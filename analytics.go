package main

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/commands"
	"github.com/jozufozu/gregory/util/cache"
	"log"
	"strings"
	"sync"
	"time"
)

func init() {
	commands.AddCommand(&commands.Command{
		Aliases:     []string{"analyze"},
		Action:      Analytics,
		Usage:       "<channelID>",
		Description: "Preforms analysis on the given channel/server.",
	})
}

type guildStats struct {
	channels map[string]*channelStats
}

type channelStats struct {
	users map[string]*userData
}

func joinStatList(l, r *channelStats) *channelStats {
	stats := &channelStats{users: make(map[string]*userData)}

	for k, v := range l.users {
		stats.users[k] = v
	}

	for k, v := range r.users {
		if data, ok := stats.users[k]; ok {
			stats.users[k] = joinUserData(data, v)
		} else {
			stats.users[k] = v
		}
	}

	return stats
}

type userData struct {
	messagesSent      uint64
	linksLinked       uint64
	imagesSent        uint64
	filesSent         uint64
	adrianCommands    uint64
	celestineCommands uint64
	gregoryCommands   uint64
	charactersUsed    map[rune]uint64
	emojisReacted     map[string]uint64
}

func joinUserData(l, r *userData) *userData {
	join := &userData{
		messagesSent:      l.messagesSent + r.messagesSent,
		linksLinked:       l.linksLinked + r.linksLinked,
		imagesSent:        l.imagesSent + r.imagesSent,
		filesSent:         l.filesSent + r.filesSent,
		adrianCommands:    l.adrianCommands + r.adrianCommands,
		celestineCommands: l.celestineCommands + r.celestineCommands,
		gregoryCommands:   l.gregoryCommands + r.gregoryCommands,
		emojisReacted:     make(map[string]uint64),
		charactersUsed:    make(map[rune]uint64),
	}

	for k, v := range l.charactersUsed {
		join.charactersUsed[k] = v
	}

	for k, v := range r.charactersUsed {
		join.charactersUsed[k] = join.charactersUsed[k] + v
	}

	return join
}

func Analytics(ctx *commands.Context, raw string, args ...string) {
	channelID := ctx.ChannelID

	if len(args) > 0 {
		channelID = args[0]
	}

	channel, err := ctx.Channel(channelID)
	guildID := channel.GuildID

	if err != nil {
		ctx.Reply("Sorry, I don't know about that place.")
		return
	}

	ctx.StateEnabled = false
	guild, err := ctx.Guild(guildID)
	ctx.StateEnabled = true

	if err != nil {
		ctx.Reply("Sorry, I don't know about that place.")
		return
	}

	if channel.Type == discordgo.ChannelTypeGuildVoice || channel.Type == discordgo.ChannelTypeGuildCategory {
		ctx.Reply(fmt.Sprintf("Sorry, I cant analyze %s.", channel.Name))
		return
	}

	ctx.ConfirmWithMessage(ctx.Author, fmt.Sprintf("Ananlyzing %s.\nThis might take a while.", guild.Name), func() {
		ctx.StartTyping()

		var totals *channelStats = nil

		t := time.Now()

		if channel.Type == discordgo.ChannelTypeGuildText {
			stats := GetGuildStats(ctx, guildID)

			totals = &channelStats{users: make(map[string]*userData)}

			for _, stats := range stats.channels {
				totals = joinStatList(stats, totals)
			}
		} else {
			var err error = nil
			totals, err = GetChannelStats(ctx, channel)

			if err != nil {
				ctx.Reply("Sorry, I had trouble with that.")
			}
		}

		log.Printf("Analysis of %s complete, took %s\n", guild.Name, time.Since(t))
		t = time.Now()

		total := new(userData)

		for _, data := range totals.users {
			total = joinUserData(data, total)
		}

		var totalCharacters uint64 = 0

		for _, used := range total.charactersUsed {
			totalCharacters += used
		}

		buf := new(bytes.Buffer)
		buf.WriteString(fmt.Sprintf("there were **%v** messages sent,\n", total.messagesSent))
		buf.WriteString(fmt.Sprintf("made of **%v** characters in total.\n", totalCharacters))
		buf.WriteString(fmt.Sprintf("With **%v** links linked,\n", total.linksLinked))
		buf.WriteString(fmt.Sprintf("and **%v** images shared.\n", total.imagesSent))
		buf.WriteString(fmt.Sprintf("Most importantly, the holee letter has been uttered **%v** times.", total.charactersUsed['e']+total.charactersUsed['E']))

		users := buildUserStatText(ctx, totals)

		ctx.DoneTyping()

		embeds := []*discordgo.MessageEmbed{{
			Title:       fmt.Sprintf("In the place known as %s", guild.Name),
			Description: fmt.Sprintf("there are %v people:", len(users)),
			Color:       0xff6611,
		}}

		embeds = append(embeds, getUserStatEmbeds(users)...)

		embeds = append(embeds, &discordgo.MessageEmbed{
			Title:       "In all...",
			Type:        "rich",
			Description: buf.String(),
			Color:       0xff6611,
		})

		for _, embed := range embeds {
			ctx.ChannelMessageSendEmbed(ctx.ChannelID, embed)
		}

		log.Printf("Analysis sent, took %s\n", time.Since(t))
	})
}
func buildUserStatText(ctx *commands.Context, totals *channelStats) []*discordgo.MessageEmbedField {
	users := make([]*discordgo.MessageEmbedField, len(totals.users))

	i := 0
	for len(totals.users) > 0 {
		var userID = ""
		var userStats *userData = nil

		for id, stats := range totals.users {
			if userStats == nil || stats.messagesSent > userStats.messagesSent {
				userID, userStats = id, stats
			}
		}

		delete(totals.users, userID)

		user := cache.LazyUserGet(userID)

		if user == nil {
			user, _ = ctx.User(userID)
		}

		if user == nil {
			continue
		}

		userField := new(discordgo.MessageEmbedField)

		buf := new(bytes.Buffer)

		userField.Name = ctx.WhatDoICall(user)

		if userStats.messagesSent > 1 {
			buf.WriteString(fmt.Sprintf("has sent %v messages", userStats.messagesSent))
		} else if userStats.messagesSent == 1 {
			buf.WriteString("has sent one message")
		} else {
			buf.WriteString("has never sent a message")
		}
		if userStats.linksLinked > 0 {
			if userStats.linksLinked > 1 {
				buf.WriteString(fmt.Sprintf(",\n%v of which had links", userStats.linksLinked))
			} else {
				buf.WriteString(",\none of which had a link")
			}
		}
		if userStats.imagesSent > 0 {
			if userStats.imagesSent > 1 {
				buf.WriteString(fmt.Sprintf(",\nand has shared %v images", userStats.imagesSent))
			} else {
				buf.WriteString(",\nand has shared one image")
			}
		}
		buf.WriteString(".\n")

		e := userStats.charactersUsed['e'] + userStats.charactersUsed['E']
		if e > 1 {
			buf.WriteString(fmt.Sprintf("They've used the holee letter %v times.", e))
		} else if e == 1 {
			buf.WriteString("They've used the holee letter one time.")
		} else {
			buf.WriteString("They are the scum of this earth.")
		}

		buf.WriteString("\n\n")
		userField.Value = buf.String()

		users[i] = userField
		i++
	}
	return users
}
func getUserStatEmbeds(users []*discordgo.MessageEmbedField) []*discordgo.MessageEmbed {
	length := len(users)
	if length < 25 {
		return []*discordgo.MessageEmbed{{
			Color:  0xff6611,
			Fields: users,
		}}
	} else {
		segments := length / 25

		if segments*25 < length {
			segments++
		}

		out := make([]*discordgo.MessageEmbed, segments)

		for segment := 0; segment < segments; segment++ {
			sliceFrom := segment * 25
			sliceTo := (segment + 1) * 25

			if sliceTo > length {
				sliceTo = length
			}

			out[segment] = &discordgo.MessageEmbed{
				Color:  0xff6611,
				Fields: users[sliceFrom:sliceTo],
			}
		}
		return out
	}
}

func GetGuildStats(ctx *commands.Context, guildID string) *guildStats {
	channels, _ := ctx.GuildChannels(guildID)

	statsChan := make(chan *guildStats, len(channels))
	serverWait := new(sync.WaitGroup)

	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildVoice || channel.Type == discordgo.ChannelTypeGuildCategory {
			continue
		}
		serverWait.Add(1)

		go func(channel *discordgo.Channel) {
			defer serverWait.Done()

			stats, err := GetChannelStats(ctx, channel)
			if err != nil {
				fmt.Printf("Cannot access %s\n", channel.Name)
				return
			}

			statsChan <- &guildStats{map[string]*channelStats{channel.ID: stats}}
		}(channel)
	}

	serverWait.Wait()

	masterStats := &guildStats{make(map[string]*channelStats)}

	for len(statsChan) > 0 {
		stats := <-statsChan

		for k, v := range stats.channels {
			masterStats.channels[k] = v
		}
	}

	return masterStats
}

func GetChannelStats(ctx *commands.Context, channel *discordgo.Channel) (*channelStats, error) {
	messages, err := getChannelMessages(ctx, channel)

	if err != nil {
		return nil, err
	}

	return parallelAnalyze(ctx, messages), nil
}

func parallelAnalyze(ctx *commands.Context, messages []*discordgo.Message) *channelStats {
	length := len(messages)
	if length <= 500 {
		return analyze(ctx, messages)
	}

	segments := length / 500

	if segments*500 < length {
		segments++
	}

	ch := make(chan *channelStats, segments)

	for segment := 0; segment < segments; segment++ {
		sliceFrom := segment * 500
		sliceTo := (segment + 1) * 500

		if sliceTo > length {
			sliceTo = length
		}

		go func(messages []*discordgo.Message) {
			ch <- analyze(ctx, messages)
		}(messages[sliceFrom:sliceTo])
	}

	stats := &channelStats{users: make(map[string]*userData)}

	i := 0
	for subStat := range ch {
		stats = joinStatList(stats, subStat)
		if i++; i == segments {
			break
		}
	}

	return stats
}

func analyze(ctx *commands.Context, messages []*discordgo.Message) *channelStats {
	stats := &channelStats{users: make(map[string]*userData)}
	wg := new(sync.WaitGroup)

	for _, message := range messages {
		if message.Type != discordgo.MessageTypeDefault {
			continue
		}

		wg.Add(2)

		go func() {
			defer wg.Done()
			if message.Reactions == nil || len(message.Reactions) == 0 {
				return
			}

			for _, react := range message.Reactions {
				users, _ := ctx.MessageReactions(message.ChannelID, message.ID, react.Emoji.ID, 100)

				for _, reactor := range users {
					data := getStatsForID(stats, reactor.ID)
					data.emojisReacted[react.Emoji.Name] = data.emojisReacted[react.Emoji.Name] + 1
				}
			}
		}()

		data := getStatsForID(stats, message.Author.ID)

		data.messagesSent++
		content := message.ContentWithMentionsReplaced()

		go func() {
			defer wg.Done()
			for _, char := range content {
				data.charactersUsed[char] = data.charactersUsed[char] + 1
			}
		}()

		if strings.HasPrefix(content, "!") {
			data.adrianCommands++
		} else if strings.HasPrefix(content, "%") {
			data.celestineCommands++
		} else if strings.HasPrefix(content, "&") {
			data.gregoryCommands++
		}

		for _, attachment := range message.Attachments {
			if attachment.Height != 0 && attachment.Width != 0 {
				data.imagesSent++
			}
			data.filesSent++
		}

		for _, embed := range message.Embeds {
			if embed.URL != "" {
				data.linksLinked++
			}
		}

		wg.Wait()
	}

	return stats
}

func getStatsForID(stats *channelStats, userID string) *userData {
	var data *userData
	if existing, ok := stats.users[userID]; !ok {
		data = &userData{
			emojisReacted:  make(map[string]uint64),
			charactersUsed: make(map[rune]uint64),
		}
		stats.users[userID] = data
	} else {
		data = existing
	}
	return data
}

func getChannelMessages(ctx *commands.Context, channel *discordgo.Channel) ([]*discordgo.Message, error) {
	t := time.Now()
	messages, err := ctx.ChannelMessages(channel.ID, 100, "", "0", "")

	if err != nil {
		return nil, err
	}

	master := make([]*discordgo.Message, len(messages))

	copy(master, messages)

	for len(messages) == 100 {
		messages, _ = ctx.ChannelMessages(channel.ID, 100, "", messages[0].ID, "")

		master = append(messages, master...)
	}

	log.Printf("Got all %v messages in %s, took %s\n", len(master), channel.Name, time.Since(t))

	return master, nil
}
