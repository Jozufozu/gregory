package commands

import "github.com/jozufozu/gregory/util"

func me(ctx *util.Context, raw string, args ...string) {
	ctx.ChannelMessageSend(ctx.ChannelID, ctx.Author.ID)
}
