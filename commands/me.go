package commands

func me(ctx *Context, raw string, args ...string) {
	ctx.ChannelMessageSend(ctx.ChannelID, ctx.Author.ID)
}
