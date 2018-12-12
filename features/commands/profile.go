package commands

import (
	"bytes"
	"fmt"
	"github.com/jozufozu/gregory/util"
	"image/png"
)

func profile(ctx *util.Context, raw string, args ...string) {
	user := ctx.GetUserOrSender(args[0])

	img, err := ctx.UserAvatarDecode(user)

	if err != nil {
		name := ctx.WhatDoICall(user)
		ctx.Reply(fmt.Sprintf("Sorry, I couldn't find a profile for %s.", name))
		return
	}

	buf := new(bytes.Buffer)

	png.Encode(buf, img)

	ctx.ChannelFileSend(ctx.ChannelID, "profile.png", buf)
}
