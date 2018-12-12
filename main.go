package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/features/commands"
	"github.com/jozufozu/gregory/features/nani"
	"github.com/jozufozu/gregory/features/reddit"
	"github.com/jozufozu/gregory/util"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	f, err := os.OpenFile("gregory.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	defer AnalyticsStore.Close()
	defer Save()
	log.SetOutput(io.MultiWriter(f, os.Stdout))

	rand.Seed(time.Now().UnixNano())

	getenv, ok := os.LookupEnv("GREGORY_TOKEN")

	if !ok {
		panic("error creating Discord session, could not find bot token")
	}

	dg, err := discordgo.New("Bot " + getenv)

	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(HandleLog)
	dg.AddHandler(commands.HandleMessage)
	dg.AddHandler(util.HandleTipRequest)
	dg.AddHandler(util.PromptHandler)
	dg.AddHandler(util.UpdateServers)
	dg.AddHandler(reddit.HandleRedditMessage)
	dg.AddHandler(nani.HandleNani)

	dg.AddHandlerOnce(util.LoadUserCache)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	defer dg.Close()

	go util.ManageTyping(dg)

	log.Println("Bot is now running.  Press CTRL-C or enter \"exit\" to exit.")
	sc := make(chan os.Signal, 1)
	//signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	go func() {
		for {
			var str string
			fmt.Scan(&str)

			if strings.HasPrefix("exit", str) {
				sc <- nil
			}
		}
	}()

	<-sc
}

func HandleLog(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := util.Context{Session: s, Message: m.Message}
	guild, _ := ctx.GetGuild()
	channel, _ := ctx.GetChannel()

	log.Printf("[%s]:[%s]: <%s> \"%s\"", guild.Name, channel.Name, m.Author.Username, m.ContentWithMentionsReplaced())
}
