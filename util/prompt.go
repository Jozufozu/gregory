package util

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"sync"
	"time"
)

type prompt struct {
	user, channel, message     string
	actionButton, cancelButton string
	action                     func()
	cancelTimeout              chan bool
}

var promptsMutex = new(sync.Mutex)
var prompts = map[*prompt]bool{}

func (ctx *Context) Confirm(user *discordgo.User, action func()) {
	ctx.ConfirmWithMessage(user, "", action)
}

func (ctx *Context) ConfirmWithMessage(user *discordgo.User, message string, action func()) {
	ctx.Prompt(user, message, "🆗", "🚫", action)
}

func (ctx *Context) Prompt(user *discordgo.User, message, actionButton, cancelButton string, action func()) {
	if user.Bot {
		action()
		return
	}

	if message != "" {
		message = strings.TrimSpace(message)
		message += "\n"
	}

	msg, _ := ctx.ChannelMessageSend(ctx.ChannelID, fmt.Sprintf("%s\n%sClick %s to confirm, %s to cancel.", user.Mention(), message, actionButton, cancelButton))
	ctx.MessageReactionAdd(ctx.ChannelID, msg.ID, actionButton)
	ctx.MessageReactionAdd(ctx.ChannelID, msg.ID, cancelButton)

	addPrompt(ctx, &prompt{
		user:         user.ID,
		channel:      ctx.ChannelID,
		message:      msg.ID,
		actionButton: actionButton,
		cancelButton: cancelButton,
		action:       action,
	})
}

func addPrompt(ctx *Context, prompt *prompt) {
	ch := make(chan bool)
	prompt.cancelTimeout = ch

	promptsMutex.Lock()
	prompts[prompt] = true
	promptsMutex.Unlock()

	timer := time.NewTimer(time.Minute).C

	go func() {
		select {
		case <-timer:
			promptsMutex.Lock()
			delete(prompts, prompt)
			promptsMutex.Unlock()
			ctx.ChannelMessageDelete(prompt.channel, prompt.message)
			break
		case <-ch:
			break
		}
	}()
}

func PromptHandler(s *discordgo.Session, add *discordgo.MessageReactionAdd) {
	promptsMutex.Lock()
	defer promptsMutex.Unlock()
	for prompt := range prompts {
		if prompt.user == add.UserID && prompt.channel == add.ChannelID && prompt.message == add.MessageID {
			if add.Emoji.Name == prompt.actionButton {
				s.ChannelMessageDelete(add.ChannelID, add.MessageID)
				go prompt.action()
			} else if add.Emoji.Name == prompt.cancelButton {
				s.ChannelMessageDelete(add.ChannelID, add.MessageID)
			}
			prompt.cancelTimeout <- true

			delete(prompts, prompt)
			break
		}
	}
}
