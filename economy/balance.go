package economy

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/commands"
	"github.com/jozufozu/gregory/util/cache"
)

func Balance(ctx *commands.Context, raw string, args ...string) {
	var user *discordgo.User

	if len(args) > 0 {
		user = cache.GetUserOrSender(ctx, args[0])
	} else {
		user = ctx.Author
	}

	ctx.ConfirmWithMessage(user, "View your balance?", func() {
		ctx.Reply(fmt.Sprintf("%s has ₣%v", ctx.WhatDoICall(user), CheckBalance(user.ID)))
	})
}
