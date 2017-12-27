package cache

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/commands"
	"strings"
)

const BotID = "387810222556708865"

var ConnectedServers = make(map[string]bool)
var ConnectedUsers = make(map[string]bool)

var namesToIDs = make(map[string]string)
var namesToUsers = make(map[string]*discordgo.User)
var idsToUsers = make(map[string]*discordgo.User)

func LazyUserGet(id string) *discordgo.User {
	if user, ok := idsToUsers[id]; ok {
		return user
	}
	return nil
}

func GetUser(ctx *commands.Context, arg string) *discordgo.User {
	if len(arg) == len("387810222556708865") {
		user, err := ctx.User(arg)

		if err == nil {
			return user
		}
	}

	if user, ok := namesToUsers[arg]; ok {
		return user
	}

	lowerArg := strings.ToLower(arg)
	for name, user := range namesToUsers {
		if strings.HasPrefix(strings.ToLower(name), lowerArg) {
			return user
		}
	}

	return nil
}

func GetUserOrSender(ctx *commands.Context, arg string) *discordgo.User {
	if len(arg) == len(BotID) {
		user, err := ctx.User(arg)

		if err == nil {
			return user
		}
	}

	if user, ok := namesToUsers[arg]; ok {
		return user
	}

	lowerArg := strings.ToLower(arg)
	for name, user := range namesToUsers {
		if strings.HasPrefix(strings.ToLower(name), lowerArg) {
			return user
		}
	}

	return ctx.Author
}

func UpdateServers(session *discordgo.Session, ready *discordgo.GuildMemberAdd) {
	ConnectedServers[ready.GuildID] = true
}

func LoadUserCache(session *discordgo.Session, ready *discordgo.Ready) {
	session.StateEnabled = false
	for _, guild := range ready.Guilds {
		ConnectedServers[guild.ID] = true
		members, _ := session.GuildMembers(guild.ID, "0", 1000)

		for _, member := range members {
			user := member.User

			ConnectedUsers[user.ID] = true
			idsToUsers[user.ID] = user
			namesToIDs[user.Username] = user.ID
			namesToUsers[user.Username] = user
		}
	}
	session.StateEnabled = true
}
