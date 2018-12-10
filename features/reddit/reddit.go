package reddit

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/util"
	"regexp"
	"strings"
)

var subredditMatch *regexp.Regexp

func init() {
	rexex, e := regexp.Compile(`[rR]\/[\w\d-]{3,}`)

	if e == nil {
		subredditMatch = rexex
	} else {
		panic(e)
	}
}

func HandleRedditMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := util.Context{Session: s, Message: m.Message}

	matches := subredditMatch.FindAllString(ctx.Content, -1)

	if len(matches) == 0 {
		return
	}

	ctx.Tip("💡", func() {
		msg := strings.Builder{}

		for _, e := range matches {
			msg.WriteString("https://reddit.com/r/")
			msg.WriteString(e[2:])
			msg.WriteString("\n")
		}

		embed := &discordgo.MessageEmbed{
			Description: msg.String(),
			Color:       0xff4906,
		}
		ctx.ChannelMessageSendEmbed(ctx.ChannelID, embed)
	})
}
