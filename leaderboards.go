package main

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/util"
	"strconv"
)

func sendLeaderBoard(ctx *util.Context, totals *ChannelStats, serverName, leaderBoardFormat string, prec int, heuristic func(stats *UserStats) float64, args ...string) {
	users := make(map[string]*UserStats)
	for k, v := range totals.Users {
		users[k] = v
	}

	var max, i uint64 = 0xffffffff, 0
	if len(args) > 0 {
		max, _ = strconv.ParseUint(args[0], 10, 64)
	}

	if u := uint64(len(users)); max > u {
		max = u
	}
	s := strconv.FormatUint(max, 10)
	maxPlaceLength := len(s)
	maxCountLength := 0

	buf := new(bytes.Buffer)
	buf.WriteString("```\n")
	for i < max && len(users) > 0 {
		var (
			userID string
			stats  *UserStats
		)

		for u, s := range users {
			if userID == "" || heuristic(s) > heuristic(stats) {
				userID, stats = u, s
			}
		}
		delete(users, userID)

		value := heuristic(stats)
		if value == 0 {
			max = i
			break
		}

		place := strconv.FormatUint(i+1, 10)

		buf.WriteString("#")
		buf.WriteString(place)

		times := strconv.FormatFloat(value, 'f', prec, 64)

		length := len(times)
		if length+len(place) > maxCountLength+maxPlaceLength {
			maxCountLength = length
		}

		for j := 0; j < maxPlaceLength-len(place); j++ {
			buf.WriteString(" ")
		}

		buf.WriteString(" with ")

		for j := 0; j < maxCountLength-length; j++ {
			buf.WriteString(" ")
		}

		buf.WriteString(times)
		buf.WriteString(": ")
		buf.WriteString(ctx.WhatDoICall(ctx.LazyUserGet(userID)))
		buf.WriteString("\n")
		i++
	}
	buf.WriteString("```")

	ctx.ChannelMessageSendEmbed(ctx.ChannelID, &discordgo.MessageEmbed{
		Title:       fmt.Sprintf(leaderBoardFormat, serverName, max),
		Description: buf.String(),
		Color:       statsColor,
	})
}

func sendReactionStats(ctx *util.Context, data *UserStats, name string, args ...string) {
	reactions := make(map[string]uint64)
	for k, v := range data.EmojisReacted {
		reactions[k] = v
	}

	var max, i uint64 = 10, 0
	if len(args) > 0 {
		max, _ = strconv.ParseUint(args[0], 10, 64)
	}

	if u := uint64(len(reactions)); max > u {
		max = u
	}

	s := strconv.FormatUint(max, 10)
	maxPlaceLength := len(s)
	maxCountLength := 0
	buf := new(bytes.Buffer)

	for i < max && len(reactions) > 0 {
		var emojiMax = ""
		var timesMax uint64

		for emoji, times := range reactions {
			if emoji == "" || times > timesMax {
				emojiMax, timesMax = emoji, times
			}
		}

		if timesMax == 0 {
			max = i + 1
			break
		}

		delete(reactions, emojiMax)

		place := strconv.FormatUint(i+1, 10)

		buf.WriteString("`#")
		buf.WriteString(place)

		times := strconv.FormatUint(timesMax, 10)

		length := len(times)
		if length+len(place) > maxCountLength+maxPlaceLength {
			maxCountLength = length
		}

		for j := 0; j < maxPlaceLength-len(place); j++ {
			buf.WriteString(" ")
		}

		buf.WriteString(" with ")

		for j := 0; j < maxCountLength-length; j++ {
			buf.WriteString(" ")
		}

		buf.WriteString(times)
		buf.WriteString(":` ")
		buf.WriteString(emojiMax)

		buf.WriteString("\n")

		i++
	}

	ctx.ChannelMessageSendEmbed(ctx.ChannelID, &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s's Top Reactions:", name),
		Description: buf.String(),
		Color:       statsColor,
	})
}
