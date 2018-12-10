package util

import (
	"github.com/bwmarrin/discordgo"
	"sync"
	"time"
)

type messageTip struct {
	channel, message string
	actionButton     string
	action           func()
	cancelTimeout    chan bool
}

var tipsMutex = new(sync.Mutex)
var tips = map[*messageTip]bool{}

func (ctx *Context) Tip(actionButton string, action func()) {
	ctx.MessageReactionAdd(ctx.ChannelID, ctx.ID, actionButton)

	addTip(ctx, &messageTip{
		channel:      ctx.ChannelID,
		message:      ctx.ID,
		actionButton: actionButton,
		action:       action,
	})
}

func addTip(ctx *Context, tip *messageTip) {
	ch := make(chan bool)
	tip.cancelTimeout = ch

	tipsMutex.Lock()
	tips[tip] = true
	tipsMutex.Unlock()

	timer := time.NewTimer(time.Minute).C

	go func() {
		select {
		case <-timer:
			tipsMutex.Lock()
			delete(tips, tip)
			tipsMutex.Unlock()

			ctx.MessageReactionRemove(ctx.ChannelID, ctx.ID, tip.actionButton, "@me")
			break
		case <-ch:
			break
		}
	}()
}

func HandleTipRequest(s *discordgo.Session, add *discordgo.MessageReactionAdd) {
	if add.UserID != BotID {
		tipsMutex.Lock()
		defer tipsMutex.Unlock()
		for tip := range tips {
			if tip.channel == add.ChannelID && tip.message == add.MessageID {
				if add.Emoji.Name == tip.actionButton {
					s.MessageReactionRemove(add.ChannelID, add.MessageID, tip.actionButton, "@me")
					s.MessageReactionRemove(add.ChannelID, add.MessageID, tip.actionButton, add.UserID)
					go tip.action()
				}
				tip.cancelTimeout <- true

				delete(tips, tip)
				break
			}
		}
	}
}
