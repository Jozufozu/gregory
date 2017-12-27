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
		ctx.Reply("Payment made, stupid.")
		return
	}

	if CheckBalance(ctx.Author.ID) < amount {
		ctx.Reply(fmt.Sprintf("Sorry %s, you don't have enough.", ctx.Author.Username))
		return
	}

	ctx.ConfirmWithMessage(ctx.Author, fmt.Sprintf("Pay %s ₣%v?", user.Username, amount), func() {
		Transfer(ctx.Author.ID, user.ID, amount)
		ctx.Reply("Payment made.")
	})
}
