package commands

import "github.com/jozufozu/gregory/util"

func play(ctx *util.Context, raw string, args ...string) {
	if len(args) > 0 {
		ctx.UpdateStatus(0, raw)
	} else {
		ctx.UpdateStatus(0, "")
	}
}
