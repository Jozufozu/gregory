package main

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/commands"
	"github.com/jozufozu/gregory/util/cache"
	"log"
	"math"
	"strings"
	"sync"
	"time"
)

const statsColor = 0xff6611

var guildAnalyses = make(map[string]*GuildStats)
var channelAnalyses = make(map[string]*ChannelStats)

func init() {
	commands.AddCommand(&commands.Command{
		Aliases:     []string{"analyze"},
		Action:      Analytics,
		Usage:       "<channelID>",
		Description: "Preforms analysis on the given channel/server.",
	})

	commands.AddCommand(&commands.Command{
		Aliases:     []string{"stats"},
		Action:      Stats,
		Usage:       "[react|messages|images|characters] <user>",
		Description: "Preforms analysis on the given channel/server.",
	})
}

func Stats(ctx *commands.Context, raw string, args ...string) {
	guild, _ := ctx.GetGuild()

	guildStats := GetLastGuildStats(guild.ID)

	if guildStats == nil {
		Analytics(ctx, "")
		return
	}

	totals := guildStats.getTotals()

	if len(args) == 0 {
		sendChannelStats(ctx, totals)
		return
	}

	var (
		user     *discordgo.User = nil
		data     *UserStats      = nil
		name                     = ""
		function                 = ""
	)

	user = cache.GetUser(ctx, args[0])
	cut := 1

	if user != nil {
		data = totals.Users[user.ID]
		name = ctx.WhatDoICall(user)

		if len(args) > 1 {
			function = strings.ToLower(args[1])
			cut = 2
		}

	} else {
		data = totals.getTotals()
		name = guild.Name
		function = strings.ToLower(args[0])
	}

	functions := map[string]func(){
		"react": func() { sendReactionStats(ctx, data, name, args[cut:]...) },
		"shrug": func() {
			if user != nil {
				if data.ShouldersShruggedOverTwo == 0 {
					ctx.Reply(fmt.Sprintf("%s has never shrugged", name))
					return
				}

				ctx.Reply(fmt.Sprintf("%s has shrugged %v times", name, data.ShouldersShruggedOverTwo))
			} else {
				sendLeaderBoard(ctx, totals, name, "%s's %v most confused:", 0, func(stats *UserStats) float64 {
					return float64(stats.ShouldersShruggedOverTwo)
				}, args[cut:]...)
			}
		},
		"messages": func() {
			if user != nil {
				if data.MessagesSent == 0 {
					ctx.Reply(fmt.Sprintf("%s has never said a word.", name))
					return
				}

				ctx.Reply(fmt.Sprintf("%s has sent %v messages.", name, data.MessagesSent))
			} else {
				sendLeaderBoard(ctx, totals, name, "%s's %v most talkative:", 0, func(stats *UserStats) float64 {
					return float64(stats.MessagesSent)
				}, args[cut:]...)
			}
		},
		"characters": func() {
			if user != nil {
				var total uint64
				for _, count := range data.CharactersUsed {
					total += count
				}
				if total == 0 {
					ctx.Reply(fmt.Sprintf("%s has never said a word.", name))
					return
				}

				ctx.Reply(fmt.Sprintf("%s has typed %v characters.", name, total))
			} else {
				sendLeaderBoard(ctx, totals, name, "%s's top %v typists:", 0, func(stats *UserStats) float64 {
					var total uint64
					for _, count := range stats.CharactersUsed {
						total += count
					}
					return float64(total)
				}, args[cut:]...)
			}
		},
		"uniqueCharacters": func() {
			if user != nil {
				u := len(data.CharactersUsed)
				if u == 0 {
					ctx.Reply(fmt.Sprintf("%s has never said a word.", name))
					return
				}

				ctx.Reply(fmt.Sprintf("%s has pressed %s different buttons.", name, u))
			} else {
				sendLeaderBoard(ctx, totals, name, "%s's %v most diverse:", 0, func(stats *UserStats) float64 {
					return float64(len(stats.CharactersUsed))
				}, args[cut:]...)
			}
		},
		"leederboard": func() {
			if user != nil {
				u := eneleze(data)
				if u == 0 {
					ctx.Reply(fmt.Sprintf("%s has never said a word.", name))
					return
				}

				ctx.Reply(fmt.Sprintf("`%s == E%v`", name, u))
			} else {
				sendLeaderBoard(ctx, totals, name, "%s's %v most eextravagant:", 4, eneleze, args[cut:]...)
			}
		},
		"": func() {
			ctx.Reply("Sorry, I don't know what you want me to do.")
		},
	}

	for k, v := range functions {
		if strings.HasPrefix(strings.ToLower(k), function) {
			v()
			return
		}
	}

	functions[""]()
}

func eneleze(stats *UserStats) float64 {
	return float64(stats.CharactersUsed['e']+stats.CharactersUsed['E']) / math.Sqrt(math.Pow(float64(stats.CharactersUsed[' ']+stats.MessagesSent), 2)+math.Pow(float64(stats.CharactersUsed['o']+stats.CharactersUsed['O']), 2))
}

func Analytics(ctx *commands.Context, raw string, args ...string) {
	channelID := ctx.ChannelID
	full := false

	if len(args) > 0 {
		full = strings.ToLower(args[0]) == "full"

		if len(args) > 1 {
			channelID = args[1]
		}
	}

	channel, err := ctx.Channel(channelID)

	if err != nil {
		ctx.Reply("Sorry, I don't know about that place.")
		return
	}

	guildID := channel.GuildID
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

	var lastGuildStats *GuildStats

	if !full {
		lastGuildStats = GetLastGuildStats(guildID)
	}

	f := func() {
		ctx.StartTyping()
		defer ctx.DoneTyping()

		t := time.Now()

		if channel.Type == discordgo.ChannelTypeGuildText {
			GetGuildStats(ctx, guildID, lastGuildStats)
		} else {
			_, err := GetChannelStats(ctx, channel, nil)

			if err != nil {
				ctx.Reply("Sorry, I had trouble with that.")
			}
		}

		ctx.Reply(fmt.Sprintf("Done! Took %s", time.Since(t)))
	}

	if lastGuildStats == nil {
		ctx.ConfirmWithMessage(ctx.Author, fmt.Sprintf("Do you want me to do a full analysis on %s?\nThis might take a while.", guild.Name), f)
	} else {
		f()
	}
}

func sendChannelStats(ctx *commands.Context, totals *ChannelStats) {
	t := time.Now()

	guild, _ := ctx.GetGuild()

	total := new(UserStats)
	for _, data := range totals.Users {
		total = joinUserStats(data, total)
	}
	var totalCharacters uint64 = 0
	for _, used := range total.CharactersUsed {
		totalCharacters += used
	}
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("there were **%v** messages sent,\n", total.MessagesSent))
	buf.WriteString(fmt.Sprintf("made of **%v** characters in total.\n", totalCharacters))
	buf.WriteString(fmt.Sprintf("With **%v** links linked,\n", total.LinksLinked))
	buf.WriteString(fmt.Sprintf("and **%v** images shared.\n", total.ImagesSent))
	buf.WriteString(fmt.Sprintf("Most importantly, the holee letter has been uttered **%v** times.", total.CharactersUsed['e']+total.CharactersUsed['E']))
	users := buildUserStatText(ctx, totals)

	embeds := []*discordgo.MessageEmbed{{
		Title:       fmt.Sprintf("In the place known as %s", guild.Name),
		Description: fmt.Sprintf("there are %v people:", len(users)),
		Color:       statsColor,
	}}
	embeds = append(embeds, getUserStatEmbeds(users)...)
	embeds = append(embeds, &discordgo.MessageEmbed{
		Title:       "In all...",
		Description: buf.String(),
		Color:       statsColor,
	})
	for _, embed := range embeds {
		ctx.ChannelMessageSendEmbed(ctx.ChannelID, embed)
	}
	log.Printf("Analysis sent, took %s\n", time.Since(t))
}

func buildUserStatText(ctx *commands.Context, totals *ChannelStats) []*discordgo.MessageEmbedField {
	builtUserFields := make([]*discordgo.MessageEmbedField, len(totals.Users))

	type send struct {
		rank  uint64
		field *discordgo.MessageEmbedField
	}

	ch := make(chan *send, len(totals.Users))
	wg := new(sync.WaitGroup)

	for userID, userStats := range totals.Users {
		wg.Add(1)
		go func(userID string, userStats *UserStats) {
			defer wg.Done()

			user := cache.LazyUserGet(ctx, userID)

			if user == nil {
				return
			}

			userField := new(discordgo.MessageEmbedField)

			buf := new(bytes.Buffer)

			userField.Name = ctx.WhatDoICall(user)

			if userStats.MessagesSent > 1 {
				buf.WriteString(fmt.Sprintf("has sent %v messages", userStats.MessagesSent))
			} else if userStats.MessagesSent == 1 {
				buf.WriteString("has sent one message")
			} else {
				buf.WriteString("has never sent a message")
			}
			if userStats.LinksLinked > 0 {
				if userStats.LinksLinked > 1 {
					buf.WriteString(fmt.Sprintf(",\n%v of which had links", userStats.LinksLinked))
				} else {
					buf.WriteString(",\none of which had a link")
				}
			}
			if userStats.ImagesSent > 0 {
				if userStats.ImagesSent > 1 {
					buf.WriteString(fmt.Sprintf(",\nand has shared %v images", userStats.ImagesSent))
				} else {
					buf.WriteString(",\nand has shared one image")
				}
			}
			buf.WriteString(".\n")

			e := userStats.CharactersUsed['e'] + userStats.CharactersUsed['E']
			if e > 1 {
				buf.WriteString(fmt.Sprintf("They've used the holee letter %v times.", e))
			} else if e == 1 {
				buf.WriteString("They've used the holee letter one time.")
			} else {
				buf.WriteString("They are the scum of this earth.")
			}

			buf.WriteString("\n\n")
			userField.Value = buf.String()
			ch <- &send{
				rank:  userStats.MessagesSent,
				field: userField,
			}
		}(userID, userStats)
	}

	wg.Wait()

	sorter := make(map[*send]bool)

	for len(ch) > 0 {
		sorter[<-ch] = true
	}

	for i := 0; len(sorter) > 0; i++ {
		var greatest *send

		for k := range sorter {
			if greatest == nil || k.rank > greatest.rank {
				greatest = k
			}
		}

		delete(sorter, greatest)

		builtUserFields[i] = greatest.field
	}

	return builtUserFields
}

func getUserStatEmbeds(users []*discordgo.MessageEmbedField) []*discordgo.MessageEmbed {
	length := len(users)
	const split = 20
	if length < split {
		return []*discordgo.MessageEmbed{{
			Color:  statsColor,
			Fields: users,
		}}
	} else {
		segments := length / split

		if segments*split < length {
			segments++
		}

		out := make([]*discordgo.MessageEmbed, segments)

		for segment := 0; segment < segments; segment++ {
			sliceFrom := segment * split
			sliceTo := (segment + 1) * split

			if sliceTo > length {
				sliceTo = length
			}

			out[segment] = &discordgo.MessageEmbed{
				Color:  statsColor,
				Fields: users[sliceFrom:sliceTo],
			}
		}
		return out
	}
}

func GetGuildStats(ctx *commands.Context, guildID string, oldGuildStats *GuildStats) *GuildStats {
	channels, _ := ctx.GuildChannels(guildID)

	statsChan := make(chan *GuildStats, len(channels))
	serverWait := new(sync.WaitGroup)

	for _, channel := range channels {
		if channel.Type == discordgo.ChannelTypeGuildVoice || channel.Type == discordgo.ChannelTypeGuildCategory {
			continue
		}
		serverWait.Add(1)

		go func(channel *discordgo.Channel) {
			defer serverWait.Done()

			var oldChannelStats *ChannelStats

			if oldGuildStats != nil {
				oldChannelStats = oldGuildStats.Channels[channel.ID]
			}

			stats, err := GetChannelStats(ctx, channel, oldChannelStats)
			if err != nil {
				fmt.Printf("Cannot access %s\n", channel.Name)
				return
			}

			statsChan <- &GuildStats{TimeStamp: nil, Channels: map[string]*ChannelStats{channel.ID: stats}}
		}(channel)
	}

	serverWait.Wait()

	masterStats := &GuildStats{TimeStamp: nil, Channels: make(map[string]*ChannelStats)}

	for len(statsChan) > 0 {
		stats := <-statsChan

		for k, v := range stats.Channels {
			masterStats.Channels[k] = v
		}
	}

	t := time.Now()
	masterStats.TimeStamp = &t
	guildAnalyses[guildID] = masterStats
	go Save()

	return masterStats
}

func GetChannelStats(ctx *commands.Context, channel *discordgo.Channel, existing *ChannelStats) (*ChannelStats, error) {
	messages, err := getChannelMessages(ctx, channel, existing)

	if err != nil {
		return nil, err
	}

	if messages == nil || len(messages) == 0 {
		if existing != nil {
			return existing, nil
		} else {
			return &ChannelStats{LastMessageID: "0", Users: make(map[string]*UserStats)}, nil
		}
	}

	channelStats := parallelAnalyze(ctx, messages)

	if existing != nil {
		channelStats = joinStatList(existing, channelStats)
	}

	channelStats.LastMessageID = messages[0].ID
	channelAnalyses[channel.ID] = channelStats
	return channelStats, nil
}

func getChannelMessages(ctx *commands.Context, channel *discordgo.Channel, existing *ChannelStats) ([]*discordgo.Message, error) {
	t := time.Now()

	after := "0"

	if existing != nil && existing.LastMessageID != "" {
		after = existing.LastMessageID
	}

	messages, err := ctx.ChannelMessages(channel.ID, 100, "", after, "")

	if len(messages) == 0 {
		log.Printf("No new messages in %s.\n", channel.Name)
		return nil, nil
	}

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

func parallelAnalyze(ctx *commands.Context, messages []*discordgo.Message) *ChannelStats {
	length := len(messages)
	if length <= 500 {
		return analyze(ctx, messages)
	}

	segments := length / 500

	if segments*500 < length {
		segments++
	}

	ch := make(chan *ChannelStats, segments)

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

	stats := &ChannelStats{Users: make(map[string]*UserStats)}

	i := 0
	for subStat := range ch {
		stats = joinStatList(stats, subStat)
		if i++; i == segments {
			break
		}
	}

	return stats
}

func analyze(ctx *commands.Context, messages []*discordgo.Message) *ChannelStats {
	stats := &ChannelStats{Users: make(map[string]*UserStats)}
	wg := new(sync.WaitGroup)
	mu := new(sync.Mutex)

	for _, message := range messages {
		if message.Type != discordgo.MessageTypeDefault {
			continue
		}
		go cache.KnowUser(message.Author)

		wg.Add(2)
		go func() {
			defer wg.Done()
			if message.Reactions == nil || len(message.Reactions) == 0 {
				return
			}

			for _, react := range message.Reactions {
				users, _ := ctx.MessageReactions(message.ChannelID, message.ID, react.Emoji.APIName(), 100)

				for _, reactor := range users {
					data := getStatsForID(stats, reactor.ID)
					name := ctx.HowDoISay(react.Emoji)
					data.EmojisReacted[name] = data.EmojisReacted[name] + 1
				}
			}
		}()

		data := getStatsForID(stats, message.Author.ID)

		for _, embed := range message.Embeds {
			if embed.URL != "" {
				data.LinksLinked++
			}
			if embed.Type == "rich" {
				wg.Add(1)
				go func() {
					defer wg.Done()
					defer mu.Unlock()
					mu.Lock()
					countCharacters(data, embed.Description)
					countCharacters(data, embed.Title)

					if len(embed.Fields) > 0 {
						for _, field := range embed.Fields {
							countCharacters(data, field.Name)
							countCharacters(data, field.Value)
						}
					}
				}()
			}
		}

		data.MessagesSent++
		content := message.ContentWithMentionsReplaced()

		go func() {
			defer wg.Done()
			defer mu.Unlock()
			mu.Lock()
			countCharacters(data, content)
		}()

		if strings.HasPrefix(content, "!") {
			data.AdrianCommands++
		} else if strings.HasPrefix(content, "%") {
			data.CelestineCommands++
		} else if strings.HasPrefix(content, "&") {
			data.GregoryCommands++
		}

		for _, attachment := range message.Attachments {
			if attachment.Height != 0 && attachment.Width != 0 {
				data.ImagesSent++
			}
			data.FilesSent++
		}

		wg.Wait()
	}

	return stats
}
func countCharacters(data *UserStats, content string) {
	data.ShouldersShruggedOverTwo += uint64(strings.Count(content, "¯\\_(ツ)_/¯"))
	for _, char := range content {
		data.CharactersUsed[char] = data.CharactersUsed[char] + 1
	}
}

func getStatsForID(stats *ChannelStats, userID string) *UserStats {
	var data *UserStats
	if existing, ok := stats.Users[userID]; !ok {
		data = &UserStats{
			EmojisReacted:  make(map[string]uint64),
			CharactersUsed: make(map[rune]uint64),
		}
		stats.Users[userID] = data
	} else {
		data = existing
	}
	return data
}
