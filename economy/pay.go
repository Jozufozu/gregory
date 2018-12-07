package economy

import (
	"fmt"
	"github.com/jozufozu/gregory/commands"
	"github.com/jozufozu/gregory/util/cache"
	"strconv"
)

func Pay(ctx *commands.Context, raw string, args ...string) {
	if len(args) < 2 {
		ctx.Reply("Sorry, pay who how much?")
		return
	}

	amount, err := strconv.ParseUint(args[1], 10, 64)

	if err != nil {
		ctx.Reply("That's no number!")
		return
	}

	if amount == 0 {
		ctx.Reply("Transfer completed before it began.")
		return
	}

	user := cache.GetUser(ctx, args[0])

	if user == nil {
		ctx.Reply("Sorry, I don't know that person.")
		return
	}

	if user.ID == ctx.Author.ID {
		ctx.Reply("Go ahead, _please_.")
		return
	}

	if CheckBalance(ctx.Author.ID) < amount {
		ctx.Reply(fmt.Sprintf("Sorry %s, you don't have enough.", ctx.WhatDoICall(ctx.Author)))
		return
	}

	ctx.ConfirmWithMessage(ctx.Author, fmt.Sprintf("Pay %s ₣%v?", ctx.WhatDoICall(user), amount), func() {
		Transfer(ctx.Author.ID, user.ID, amount)
		ctx.Reply("Payment made.")
	})
}

func Request(ctx *commands.Context, raw string, args ...string) {
	if len(args) < 2 {
		ctx.Reply("Sorry, you want how much from who?")
		return
	}

	amount, err := strconv.ParseUint(args[1], 10, 64)

	if err != nil {
		ctx.Reply("That's no number!")
		return
	}

	if amount == 0 {
		ctx.Reply("Transfer completed before it began.")
		return
	}

	user := cache.GetUser(ctx, args[0])

	if user == nil {
		ctx.Reply("Sorry, I don't know that person.")
		return
	}

	if user.Bot {
		ctx.Reply(fmt.Sprintf("I'm sorry %s, I can't do that for you.", ctx.WhatDoICall(ctx.Author)))
		return
	}

	if user.ID == ctx.Author.ID {
		ctx.Reply("Payment made.")
		return
	}

	if CheckBalance(user.ID) < amount {
		ctx.Reply(fmt.Sprintf("Sorry %s, %s doesn't have enough.", ctx.WhatDoICall(ctx.Author), ctx.WhatDoICall(user)))
		return
	}

	ctx.ConfirmWithMessage(user, fmt.Sprintf("Pay %s ₣%v?", ctx.WhatDoICall(user), amount), func() {
		Transfer(user.ID, ctx.Author.ID, amount)
		ctx.Reply("Payment made.")
	})
}
