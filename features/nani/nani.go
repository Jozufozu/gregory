package nani

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/util"
	"regexp"
)

const nani string = "何？！"

var naniMatch *regexp.Regexp

func init() {
	rexex, e := regexp.Compile(`[[nN]\s*[aA]\s*[nN]\s*[iI]`)

	if e == nil {
		naniMatch = rexex
	} else {
		panic(e)
	}
}

func HandleNani(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := util.Context{Session: s, Message: m.Message}

	if naniMatch.MatchString(ctx.Content) {
		ctx.Reply(nani)
	}
}
