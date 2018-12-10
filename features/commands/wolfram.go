package commands

import (
	"fmt"
	"github.com/Krognol/go-wolfram"
	"github.com/jozufozu/gregory/util"
	"os"
	"strings"
)

var wa *wolfram.Client

func know(ctx *util.Context, raw string, args ...string) {
	res, e := wa.GetShortAnswerQuery(strings.TrimSpace(raw), wolfram.Metric, 0)

	//closer, _, e := wa.GetSimpleQuery(strings.TrimSpace(raw), url.Values{
	//	"background": []string{"36393f"},
	//	"foreground": []string{"white"},
	//})

	if e != nil {
		ctx.ChannelMessageSend(ctx.ChannelID, "bad")
	}

	ctx.ChannelMessageSend(ctx.ChannelID, res)
	//ctx.ChannelFileSend(ctx.ChannelID, "query.png", closer)
}

func init() {
	getenv, ok := os.LookupEnv("GREGORY_WOLFRAM")

	if !ok {
		fmt.Println("error creating Wolfram|Alpha client, missing AppID")
		return
	}

	wa = &wolfram.Client{AppID: getenv}
}
