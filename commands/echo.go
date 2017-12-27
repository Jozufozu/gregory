package commands

import "strings"

func echo(ctx *Context, raw string, args ...string) {
	ctx.ChannelMessageSend(ctx.ChannelID, strings.TrimSpace(raw))
}
