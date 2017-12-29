package cache

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jozufozu/gregory/commands"
	"strings"
	"sync"
)

const BotID = "387810222556708865"

var (
	SaveUser         = make(chan *discordgo.User, 256)
	ConnectedServers = make(map[string]bool)
	ConnectedUsers   = make(map[string]bool)

	namesToIDs           = make(map[string]string)
	namesToUsers         = make(map[string]*discordgo.User)
	guildsToNicksToUsers = make(map[string]map[string]*discordgo.User)
	idsToUsers           = make(map[string]*discordgo.User)
	mu                   = new(sync.Mutex)
)

func init() {
	go func() {
		for {
			select {
			case user := <-SaveUser:
				mu.Lock()
				idsToUsers[user.ID] = user
				mu.Unlock()
			}
		}
	}()
}

func KnowUser(user *discordgo.User) bool {
	mu.Lock()
	_, know := idsToUsers[user.ID]
	mu.Unlock()
	if !know {
		SaveUser <- user
	}
	return know
}

func LazyUserGet(ctx *commands.Context, userID string) *discordgo.User {
	mu.Lock()
	defer mu.Unlock()

	if user, ok := idsToUsers[userID]; ok {
		return user
	}

	defer func() { ctx.StateEnabled = true }()
	ctx.StateEnabled = false
	if user, err := ctx.User(userID); err != nil {
		return user
	}

	return nil
}

func GetUserOrSender(ctx *commands.Context, arg string) *discordgo.User {
	user := GetUser(ctx, arg)

	if user != nil {
		return user
	}

	return ctx.Author
}

func GetUser(ctx *commands.Context, arg string) *discordgo.User {
	if len(arg) == len(BotID) {
		user, err := ctx.User(arg)

		if err == nil {
			return user
		}
	}
	guild, _ := ctx.GetGuild()

	if user, ok := namesToUsers[arg]; ok {
		return user
	}

	nicksToUsers := guildsToNicksToUsers[guild.ID]

	lowerArg := strings.ToLower(arg)
	for name, user := range namesToUsers {
		if strings.HasPrefix(strings.ToLower(name), lowerArg) {
			return user
		}
	}

	for name, user := range nicksToUsers {
		if strings.HasPrefix(strings.ToLower(name), lowerArg) {
			return user
		}
	}

	return nil
}

func UpdateServers(session *discordgo.Session, ready *discordgo.GuildMemberAdd) {
	ConnectedServers[ready.GuildID] = true
}

func LoadUserCache(session *discordgo.Session, ready *discordgo.Ready) {
	session.StateEnabled = false
	for _, guild := range ready.Guilds {
		ConnectedServers[guild.ID] = true
		guildsToNicksToUsers[guild.ID] = make(map[string]*discordgo.User)

		members, _ := session.GuildMembers(guild.ID, "0", 1000)

		for _, member := range members {
			user := member.User

			name := user.Username

			if member.Nick != "" {
				name = member.Nick
			}

			ConnectedUsers[user.ID] = true

			idsToUsers[user.ID] = user
			namesToIDs[user.Username] = user.ID
			namesToUsers[user.Username] = user
			guildsToNicksToUsers[guild.ID][name] = user
		}
	}
	session.StateEnabled = true
}
