package commands

func play(ctx *Context, raw string, args ...string) {
	if len(args) > 0 {
		ctx.UpdateStatus(0, raw)
	} else {
		ctx.UpdateStatus(0, "")
	}
}
