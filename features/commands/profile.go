package commands

//import (
//	"bytes"
//	"image/png"
//	"github.com/jozufozu/gregory/util/cache"
//)
//
//func profile(ctx *Context, raw string, args ...string) {
//	user := cache.GetUserOrSender(ctx, args[0])
//
//	img, err := ctx.UserAvatarDecode(user)
//
//	if err != nil {
//		ctx.Reply("Sorry, I couldn't get anything.")
//		return
//	}
//
//	buf := new(bytes.Buffer)
//
//	png.Encode(buf, img)
//
//	ctx.ChannelFileSend(ctx.ChannelID, "profile.png", buf)
//}
