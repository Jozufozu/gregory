package commands

import (
	"fmt"
	"github.com/Krognol/go-wolfram"
	"os"
	"strings"
)

var wa *wolfram.Client

func know(ctx *Context, raw string, args ...string) {
	res, e := wa.GetShortAnswerQuery(strings.TrimSpace(raw), wolfram.Metric, 0)

	if e != nil {
		ctx.ChannelMessageSend(ctx.ChannelID, "bad")
	}

	ctx.ChannelMessageSend(ctx.ChannelID, res)
}

func init() {
	getenv, ok := os.LookupEnv("GREGORY_WOLFRAM")

	if !ok {
		fmt.Println("error creating Wolfram|Alpha client, missing AppID")
		return
	}

	wa = &wolfram.Client{AppID: getenv}
}
