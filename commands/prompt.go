package commands

import (
	"github.com/bwmarrin/discordgo"
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
