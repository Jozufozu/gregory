package commands

import (
	"github.com/gojp/kana"
	"github.com/jozufozu/gregory/util"
)

func init() {
	kana.Initialize()
}

func weeb(ctx *util.Context, raw string, args ...string) {
	if kana.IsKana(raw) {
		ctx.Reply(kana.KanaToRomaji(raw))
	} else if kana.IsLatin(raw) {
		ctx.Reply(kana.RomajiToHiragana(kana.NormalizeRomaji(raw)))
	}
}
