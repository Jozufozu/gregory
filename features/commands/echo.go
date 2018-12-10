package commands

import (
	"github.com/jozufozu/gregory/util"
	"strings"
)

func echo(ctx *util.Context, raw string, args ...string) {
	ctx.ChannelMessageSend(ctx.ChannelID, strings.TrimSpace(raw))
}
