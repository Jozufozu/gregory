package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/commands"
	"github.com/jozufozu/gregory/economy"
	"github.com/jozufozu/gregory/util/cache"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode"
)

func main() {
	f, err := os.OpenFile("gregory.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(io.MultiWriter(f, os.Stdout))
	defer economy.Data.Close()
	defer AnalyticsStore.Close()
	defer Save()

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
	dg.AddHandler(HandleReddit)
	dg.AddHandler(commands.HandleMessage)
	dg.AddHandler(commands.PromptHandler)

	dg.AddHandlerOnce(cache.LoadUserCache)
	dg.AddHandler(cache.UpdateServers)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	defer dg.Close()

	go commands.ManageTyping(dg)

	log.Println("Bot is now running.  Press CTRL-C or enter \"exit\" to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

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

func HandleReddit(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := commands.Context{Session: s, Message: m.Message}

	if index := strings.Index(m.Content, "r/"); index >= 0 {
		end := -1
		for i, rune := range m.Content[index+2:] {
			if !(unicode.IsLetter(rune) || unicode.IsNumber(rune) || rune == '-' || rune == '_') {
				end = i + index - 2
				break
			}
		}

		if end == -1 {
			end = len(m.Content)
		}

		embed := &discordgo.MessageEmbed{
			Title: m.Content[index:end],
			URL:   "https://reddit.com/r/" + m.Content[(index+2):end],
			Color: 0xff4906,
		}
		ctx.ChannelMessageSendEmbed(ctx.ChannelID, embed)
	}
}

func HandleLog(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := commands.Context{Session: s, Message: m.Message}
	guild, _ := ctx.GetGuild()
	channel, _ := ctx.GetChannel()

	log.Printf("[%s]:[%s]: <%s> \"%s\"", guild.Name, channel.Name, m.Author.Username, m.ContentWithMentionsReplaced())
}
