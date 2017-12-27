package commands

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

type Context struct {
	*discordgo.Session
	*discordgo.Message
	guild *discordgo.Guild
}

var (
	typingStarter  = make(chan string)
	typingFinisher = make(chan string)
)

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

func (ctx *Context) Confirm(user *discordgo.User, action func()) {
	ctx.ConfirmWithMessage(user, "", action)
}

func (ctx *Context) ConfirmWithMessage(user *discordgo.User, message string, action func()) {
	ctx.Prompt(user, message, "🆗", "🚫", action)
}

func (ctx *Context) Prompt(user *discordgo.User, message, actionButton, cancelButton string, action func()) {
	if user.Bot {
		action()
		return
	}

	if message != "" {
		message = strings.TrimSpace(message)
		message += "\n"
	}

	msg, _ := ctx.ChannelMessageSend(ctx.ChannelID, fmt.Sprintf("%s\n%sClick %s to confirm, %s to cancel.", user.Mention(), message, actionButton, cancelButton))
	ctx.MessageReactionAdd(ctx.ChannelID, msg.ID, actionButton)
	ctx.MessageReactionAdd(ctx.ChannelID, msg.ID, cancelButton)

	addPrompt(ctx, &prompt{
		user:         user.ID,
		channel:      ctx.ChannelID,
		message:      msg.ID,
		actionButton: actionButton,
		cancelButton: cancelButton,
		action:       action,
	})
}
