package commands

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/util"
	"strings"
	"unicode"
)

func init() {

	AddCommand(&Command{
		Aliases:     []string{"help", "?"},
		Action:      help,
		Usage:       "",
		Description: "Provides a list of available commands",
	})

	AddCommand(&Command{
		Aliases:     []string{"play"},
		Action:      play,
		Usage:       "[game]",
		Description: "Sets or clears Gregory's playing status.",
	})

	AddCommand(&Command{
		Aliases:     []string{"profile"},
		Action:      profile,
		Usage:       "[user]",
		Description: "Returns the full size profile picture of [user]",
	})

	AddCommand(&Command{
		Aliases:     []string{"weeb"},
		Action:      weeb,
		Usage:       "[text]",
		Description: "Transforms text to and from japanese characters",
	})

	AddCommand(&Command{
		Aliases:     []string{"conway", "gol"},
		Action:      conway,
		Usage:       "",
		Description: "Sets whether or not to use random.org for rng",
	})

	AddCommand(&Command{
		Aliases:     []string{"d20"},
		Action:      d20,
		Usage:       "",
		Description: "Rolls a single d20",
	})

	AddCommand(&Command{
		Aliases:     []string{"roll", "dice"},
		Action:      roll,
		Usage:       "[sides] [dice]",
		Description: "Rolls some dice. Defaults are 6 sides once",
	})

	AddCommand(&Command{
		Aliases:     []string{"me", "crisis"},
		Action:      me,
		Usage:       "",
		Description: "What am I?",
	})

	AddCommand(&Command{
		Aliases:     []string{"bored", "useless", "explore"},
		Action:      bored,
		Usage:       "",
		Description: "Gives you something to do",
	})

	AddCommand(&Command{
		Aliases:     []string{"know", "learn", "k"},
		Action:      know,
		Usage:       "[query]",
		Description: "Ask wolfram alpha something",
	})

	AddCommand(&Command{
		Aliases:     []string{"echo", "say"},
		Action:      echo,
		Usage:       "<text...>",
		Description: "Says things that you say",
	})
}

const Bang = "&"

// raw is the raw input of everything after the command token
// args is all tokens after the command token separated by whitespace
type Action func(ctx *util.Context, raw string, args ...string)

type Command struct {
	Aliases     []string
	Action      Action
	Usage       string
	Description string
}

var commands = make([]*Command, 0)
var names = make([]string, 0)
var actions = make(map[string]Action)

func AddCommand(command *Command) {
	if len(command.Aliases) == 0 {
		panic(command)
	}

	for _, name := range command.Aliases {
		actions[name] = command.Action
		names = append(names, name)
	}

	commands = append(commands, command)
}

func HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if strings.HasPrefix(m.Content, Bang) {
		RunCommand(&util.Context{Session: s, Message: m.Message}, m.Content[1:])
	}
}

func RunCommand(ctx *util.Context, cmd string) {
	args := splitArgs(&cmd)

	guess := util.InferCommand(args[0], names)

	if action, ok := actions[guess]; ok {
		raw := cmd[strings.Index(cmd, args[0])+len(args[0]):]
		action(ctx, raw, args[1:]...)
	} else {
		ctx.Reply(fmt.Sprintf("Sorry, I don't know what you mean by \"%s\"", cmd))
	}
}

func splitArgs(raw *string) []string {
	out := make([]string, 0)

	buff := ""
	for _, ru := range *raw {
		if unicode.IsSpace(ru) {
			if buff != "" {
				out = append(out, buff)
			}
			buff = ""
		} else {
			buff += string(ru)
		}
	}
	if buff != "" {
		out = append(out, buff)
	}

	return out
}

func help(ctx *util.Context, raw string, args ...string) {
	if len(args) == 0 {
		fields := make([]*discordgo.MessageEmbedField, len(commands))

		buf := new(bytes.Buffer)
		for i, cmd := range commands {
			for j, name := range cmd.Aliases {
				buf.WriteString(Bang)
				buf.WriteString(name)
				if j != len(cmd.Aliases)-1 {
					buf.WriteString(" | ")
				}
			}
			buf.WriteString(" ")
			buf.WriteString(cmd.Usage)

			f := &discordgo.MessageEmbedField{
				Name:  buf.String(),
				Value: fmt.Sprintf("*%s*", cmd.Description),
			}

			fields[i] = f

			buf.Reset()
		}

		embed := &discordgo.MessageEmbed{
			Title:  "Commands:",
			Color:  0x23aaee,
			Fields: fields,
		}

		ctx.ChannelMessageSendEmbed(ctx.ChannelID, embed)
	} else {
		find := args[0]
		for _, cmd := range commands {
			for _, name := range cmd.Aliases {
				if strings.HasPrefix(name, find) {
					buf := new(bytes.Buffer)

					for j, name := range cmd.Aliases {
						buf.WriteString(Bang)
						buf.WriteString(name)
						if j != len(cmd.Aliases)-1 {
							buf.WriteString(" | ")
						}
					}
					buf.WriteString(" ")
					buf.WriteString(cmd.Usage)

					ctx.ChannelMessageSendEmbed(ctx.ChannelID, &discordgo.MessageEmbed{
						Title:       buf.String(),
						Description: fmt.Sprintf("*%s*", cmd.Description),
						Color:       0x23aaee,
					})
				}
			}
		}
	}
}
