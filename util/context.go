package util

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"sync"
	"time"
)

type Context struct {
	*discordgo.Session
	*discordgo.Message
	guild *discordgo.Guild
}

const BotID = "387810222556708865"

var (
	typingStarter  = make(chan string)
	typingFinisher = make(chan string)

	SaveUser         = make(chan *discordgo.User, 256)
	ConnectedServers = make(map[string]bool)
	ConnectedUsers   = make(map[string]bool)

	namesToIDs           = make(map[string]string)
	namesToUsers         = make(map[string]*discordgo.User)
	guildsToNicksToUsers = make(map[string]map[string]*discordgo.User)
	idsToUsers           = make(map[string]*discordgo.User)
	mu                   = new(sync.Mutex)
)

func init() {
	go func() {
		for {
			select {
			case user := <-SaveUser:
				mu.Lock()
				idsToUsers[user.ID] = user
				mu.Unlock()
			}
		}
	}()
}

func ManageTyping(s *discordgo.Session) {
	typingChannels := map[string]bool{}
	timer := time.Tick(time.Second * 4)

	for {
		select {
		case <-timer:
			for channel, typing := range typingChannels {
				if typing {
					s.ChannelTyping(channel)
				}
			}
		case id := <-typingStarter:
			typingChannels[id] = true
			s.ChannelTyping(id)
		case id := <-typingFinisher:
			delete(typingChannels, id)
		}
	}
}

func (ctx *Context) WhatDoICall(user *discordgo.User) (name string) {
	if user == nil {
		return "An old friend"
	}
	name = user.Username

	if channel, err := ctx.Channel(ctx.ChannelID); err == nil {
		if member, err := ctx.GuildMember(channel.GuildID, user.ID); err == nil && member.Nick != "" {
			name = member.Nick
		}
	}

	return
}

func (ctx *Context) HowDoISay(emoji *discordgo.Emoji) (name string) {
	if emoji.ID == "" {
		name = emoji.Name
	} else {
		name = fmt.Sprintf("<:%s:%s>", emoji.Name, emoji.ID)
	}

	return
}

func (ctx *Context) GetGuild() (*discordgo.Guild, error) {

	if ctx.guild != nil {
		return ctx.guild, nil
	}

	channel, err := ctx.GetChannel()

	if err != nil {
		return nil, err
	}

	ctx.StateEnabled = false
	guild, err := ctx.Guild(channel.GuildID)
	ctx.StateEnabled = true

	if err != nil {
		return nil, err
	}
	ctx.guild = guild

	return guild, err
}

func (ctx *Context) GetChannel() (*discordgo.Channel, error) {
	return ctx.Channel(ctx.ChannelID)
}

func (ctx *Context) StartTyping() {
	typingStarter <- ctx.ChannelID
}

func (ctx *Context) DoneTyping() {
	typingFinisher <- ctx.ChannelID
}

func (ctx *Context) Reply(msg string) {
	ctx.ChannelMessageSend(ctx.ChannelID, msg)
}

func (ctx *Context) ReplyDelete(msg string, after time.Duration) {
	message, _ := ctx.ChannelMessageSend(ctx.ChannelID, msg)

	timer := time.NewTimer(after).C

	go func() {
		select {
		case <-timer:
			ctx.ChannelMessageDelete(message.ChannelID, message.ID)
			break
		}
	}()
}

func KnowUser(user *discordgo.User) bool {
	mu.Lock()
	_, know := idsToUsers[user.ID]
	mu.Unlock()
	if !know {
		SaveUser <- user
	}
	return know
}

func (ctx *Context) LazyUserGet(userID string) *discordgo.User {
	mu.Lock()
	defer mu.Unlock()

	if user, ok := idsToUsers[userID]; ok {
		return user
	}

	defer func() { ctx.StateEnabled = true }()
	ctx.StateEnabled = false
	if user, err := ctx.User(userID); err != nil {
		return user
	}

	return nil
}

func (ctx *Context) GetUserOrSender(arg string) *discordgo.User {
	user := ctx.GetUser(arg)

	if user != nil {
		return user
	}

	return ctx.Author
}

func (ctx *Context) GetUser(arg string) *discordgo.User {
	if len(arg) == len(BotID) {
		user, err := ctx.User(arg)

		if err == nil {
			return user
		}
	}
	guild, _ := ctx.GetGuild()

	if user, ok := namesToUsers[arg]; ok {
		return user
	}

	nicksToUsers := guildsToNicksToUsers[guild.ID]

	lowerArg := strings.ToLower(arg)
	for name, user := range namesToUsers {
		if strings.HasPrefix(strings.ToLower(name), lowerArg) {
			return user
		}
	}

	for name, user := range nicksToUsers {
		if strings.HasPrefix(strings.ToLower(name), lowerArg) {
			return user
		}
	}

	return nil
}

func UpdateServers(session *discordgo.Session, ready *discordgo.GuildMemberAdd) {
	ConnectedServers[ready.GuildID] = true
}

func LoadUserCache(session *discordgo.Session, ready *discordgo.Ready) {
	session.StateEnabled = false
	for _, guild := range ready.Guilds {
		ConnectedServers[guild.ID] = true
		guildsToNicksToUsers[guild.ID] = make(map[string]*discordgo.User)

		members, _ := session.GuildMembers(guild.ID, "0", 1000)

		for _, member := range members {
			user := member.User

			name := user.Username

			if member.Nick != "" {
				name = member.Nick
			}

			ConnectedUsers[user.ID] = true

			idsToUsers[user.ID] = user
			namesToIDs[user.Username] = user.ID
			namesToUsers[user.Username] = user
			guildsToNicksToUsers[guild.ID][name] = user
		}
	}
	session.StateEnabled = true
}
