package commands

import (
	"github.com/jozufozu/gregory/util"
	"math/rand"
	"strconv"
)

func d20(ctx *util.Context, raw string, args ...string) {
	ctx.Reply(strconv.FormatInt(rand.Int63n(20), 10))
}

func roll(ctx *util.Context, raw string, args ...string) {
	var sides, rolls int64 = 6, 1

	if len(args) > 0 {
		s := args[0]
		i, err := strconv.ParseInt(s, 10, 64)

		if err == nil {
			sides = i
		}
	}

	if len(args) > 1 {
		s := args[1]
		i, err := strconv.ParseInt(s, 10, 64)

		if err == nil {
			rolls = i
		}
	}

	out := int64(0)

	ch := make(chan int64, rolls)

	for i := int64(0); i < rolls; i++ {
		go func() {
			ch <- rand.Int63n(sides)
		}()
	}

	for i := range ch {
		out += i
		rolls--
		if rolls == 0 {
			break
		}
	}

	ctx.Reply(strconv.FormatInt(out, 10))
}
