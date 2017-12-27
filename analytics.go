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
	"github.com/boltdb/bolt"
	"errors"
	"encoding/binary"
	"encoding/json"
)

const statsColor = 0xff6611

var AnalyticsStore *bolt.DB
var guildAnalyses = make(map[string]*GuildStats)
var channelAnalyses = make(map[string]*ChannelStats)

func init() {
	db, err := bolt.Open("analytics.db", 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	AnalyticsStore = db

	AnalyticsStore.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("guilds"))
		return nil
	})

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

func Save() {
	AnalyticsStore.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("guilds"))

		for id, stats := range guildAnalyses {
			guild, _ := bucket.CreateBucketIfNotExists([]byte(id))
			b := make([]byte, 8)
			binary.PutUvarint(b, guild.Sequence())
			marshal, err := json.Marshal(stats)

			if err != nil {
				continue
			}

			guild.Put(b, marshal)
		}

		return nil
	})
}

func GetLastGuildStats(guildID string) (stats *GuildStats) {
	AnalyticsStore.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("guilds"))
		guildStats := bucket.Bucket([]byte(guildID))

		if guildStats == nil {
			return errors.New("")
		}

		b := make([]byte, 8)
		binary.PutUvarint(b, guildStats.Sequence())

		stats = new(GuildStats)
		json.Unmarshal(guildStats.Get(b), stats)

		return nil
	})

	return stats
}

type GuildStats struct {
	TimeStamp *time.Time
	Channels  map[string]*ChannelStats
}

func (gst *GuildStats) getTotals() (totals *ChannelStats) {
	totals = &ChannelStats{Users: make(map[string]*UserStats)}

	for _, stats := range gst.Channels {
		totals = joinStatList(stats, totals)
	}
	return
}

type ChannelStats struct {
	LastMessageID string
	Users map[string]*UserStats
}

func (ch *ChannelStats) getTotals() (totals *UserStats) {
	totals = &UserStats{
		EmojisReacted:  make(map[string]uint64),
		CharactersUsed: make(map[rune]uint64),
	}

	for _, stats := range ch.Users {
		totals = joinUserStats(stats, totals)
	}
	return
}

func joinStatList(l, r *ChannelStats) *ChannelStats {
	stats := &ChannelStats{Users: make(map[string]*UserStats)}

	for k, v := range l.Users {
		stats.Users[k] = v
	}

	for k, v := range r.Users {
		if data, ok := stats.Users[k]; ok {
			stats.Users[k] = joinUserStats(data, v)
		} else {
			stats.Users[k] = v
		}
	}

	return stats
}

type UserStats struct {
	MessagesSent      uint64
	LinksLinked       uint64
	ImagesSent        uint64
	FilesSent         uint64
	AdrianCommands    uint64
	CelestineCommands uint64
	GregoryCommands   uint64
	CharactersUsed    map[rune]uint64
	EmojisReacted     map[string]uint64
}

func joinUserStats(l, r *UserStats) *UserStats {
	join := &UserStats{
		MessagesSent:      l.MessagesSent + r.MessagesSent,
		LinksLinked:       l.LinksLinked + r.LinksLinked,
		ImagesSent:        l.ImagesSent + r.ImagesSent,
		FilesSent:         l.FilesSent + r.FilesSent,
		AdrianCommands:    l.AdrianCommands + r.AdrianCommands,
		CelestineCommands: l.CelestineCommands + r.CelestineCommands,
		GregoryCommands:   l.GregoryCommands + r.GregoryCommands,
		EmojisReacted:     make(map[string]uint64),
		CharactersUsed:    make(map[rune]uint64),
	}

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		for k, v := range l.CharactersUsed {
			join.CharactersUsed[k] = v
		}

		for k, v := range r.CharactersUsed {
			join.CharactersUsed[k] = join.CharactersUsed[k] + v
		}
		wg.Done()
	}()

	go func() {
		for k, v := range l.EmojisReacted {
			join.EmojisReacted[k] = v
		}

		for k, v := range r.EmojisReacted {
			join.EmojisReacted[k] = join.EmojisReacted[k] + v
		}
		wg.Done()
	}()

	wg.Wait()
	return join
}

func Stats(ctx *commands.Context, raw string, args ...string) {
	guild, _ := ctx.GetGuild()

	guildStats, ok := guildAnalyses[guild.ID]

	if !ok {
		guildStats = GetLastGuildStats(guild.ID)
	}

	if guildStats == nil {
		Analytics(ctx, "")
		return
	}

	ctx.StartTyping()
	defer ctx.DoneTyping()

	totals := guildStats.getTotals()

	if len(args) == 0 {
		sendChannelStats(ctx, totals)
		return
	}

	var (
		user *discordgo.User = nil
		data *UserStats      = nil
		name                 = ""
		function             = ""
	)

	user = cache.GetUser(ctx, args[0])

	if user != nil {
		data = totals.Users[user.ID]
		name = ctx.WhatDoICall(user)

		if len(args) > 1 {
			function = args[1]
		}

	} else {
		data = totals.getTotals()
		name = guild.Name
		function = args[0]
	}

	functions := map[string]func(){
		"react": func() {
			buf := new(bytes.Buffer)

			reactions := make(map[string]uint64)
			for k, v := range data.EmojisReacted {
				reactions[k] = v
			}

			for len(reactions) > 0 {
				var emojiMax = ""
				var timesMax uint64

				for emoji, times := range reactions {
					if emoji == "" || times > timesMax {
						emojiMax, timesMax = emoji, times
					}
				}
				delete(reactions, emojiMax)

				buf.WriteString(fmt.Sprintf("%s: %v\n", emojiMax, timesMax))
			}

			ctx.ChannelMessageSendEmbed(ctx.ChannelID, &discordgo.MessageEmbed{
				Title:       fmt.Sprintf("%s's Reactions:", name),
				Description: buf.String(),
				Color:       statsColor,
			})
		},
		"": func() {

		},
	}

	if whatToDo, do := functions[function]; do {
		whatToDo()
	} else {
		functions[""]()
	}
}

func Analytics(ctx *commands.Context, raw string, args ...string) {
	channelID := ctx.ChannelID

	if len(args) > 0 {
		channelID = args[0]
	}

	channel, err := ctx.Channel(channelID)

	if err != nil {
		ctx.Reply("Sorry, I don't know about that place.")
		return
	}

	guildID := channel.GuildID
	guild, err := ctx.Guild(guildID)

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
		defer ctx.DoneTyping()

		t := time.Now()

		if channel.Type == discordgo.ChannelTypeGuildText {
			GetGuildStats(ctx, guildID)
		} else {
			_, err := GetChannelStats(ctx, channel, nil)

			if err != nil {
				ctx.Reply("Sorry, I had trouble with that.")
			}
		}

		ctx.Reply(fmt.Sprintf("Done! Took %s", time.Since(t)))
	})
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
		Type:        "rich",
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

	users := make(map[string]*UserStats)
	for k, v := range totals.Users {
		users[k] = v
	}

	i := 0
	for len(users) > 0 {
		var userID = ""
		var userStats *UserStats = nil

		for id, stats := range users {
			if userStats == nil || stats.MessagesSent > userStats.MessagesSent {
				userID, userStats = id, stats
			}
		}

		delete(users, userID)

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

		builtUserFields[i] = userField
		i++
	}
	return builtUserFields
}
func getUserStatEmbeds(users []*discordgo.MessageEmbedField) []*discordgo.MessageEmbed {
	length := len(users)
	if length < 25 {
		return []*discordgo.MessageEmbed{{
			Color:  statsColor,
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

func GetGuildStats(ctx *commands.Context, guildID string) *GuildStats {
	channels, _ := ctx.GuildChannels(guildID)

	statsChan := make(chan *GuildStats, len(channels))
	serverWait := new(sync.WaitGroup)

	oldGuildStats := GetLastGuildStats(guildID)

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

	return masterStats
}

func GetChannelStats(ctx *commands.Context, channel *discordgo.Channel, existing *ChannelStats) (*ChannelStats, error) {
	messages, err := getChannelMessages(ctx, channel, existing)

	if err != nil {
		return nil, err
	}

	if len(messages) == 0 {
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
				users, _ := ctx.MessageReactions(message.ChannelID, message.ID, react.Emoji.APIName(), 100)

				for _, reactor := range users {
					data := getStatsForID(stats, reactor.ID)
					name := ctx.HowDoISay(react.Emoji)
					data.EmojisReacted[name] = data.EmojisReacted[name] + 1
				}
			}
		}()

		data := getStatsForID(stats, message.Author.ID)

		data.MessagesSent++
		content := message.ContentWithMentionsReplaced()

		go func() {
			defer wg.Done()
			for _, char := range content {
				data.CharactersUsed[char] = data.CharactersUsed[char] + 1
			}
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

		for _, embed := range message.Embeds {
			if embed.URL != "" {
				data.LinksLinked++
			}
		}

		wg.Wait()
	}

	return stats
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

func getChannelMessages(ctx *commands.Context, channel *discordgo.Channel, existing *ChannelStats) ([]*discordgo.Message, error) {
	t := time.Now()

	after := "0"

	if existing != nil && existing.LastMessageID != "" {
		after = existing.LastMessageID
	}

	messages, err := ctx.ChannelMessages(channel.ID, 100, "", after, "")

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
