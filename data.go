package main

import (
	"encoding/binary"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"log"
	"sync"
	"time"
)

var AnalyticsStore *bolt.DB

func init() {
	db, err := bolt.Open("analytics.db", 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	AnalyticsStore = db

	AnalyticsStore.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("guilds"))
		return nil
	})
}

func Save() {
	AnalyticsStore.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("guilds"))

		for id, stats := range guildAnalyses {
			guild, _ := bucket.CreateBucketIfNotExists([]byte(id))
			b := make([]byte, 8)
			binary.PutUvarint(b, guild.Sequence())
			marshal, err := json.Marshal(stats)

			if err != nil {
				continue
			}

			guild.Put(b, marshal)
		}

		return nil
	})
}

func GetLastGuildStats(guildID string) (stats *GuildStats) {
	if stats, ok := guildAnalyses[guildID]; ok {
		return stats
	}

	AnalyticsStore.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("guilds"))
		guildStats := bucket.Bucket([]byte(guildID))

		if guildStats == nil {
			return errors.New("")
		}

		b := make([]byte, 8)
		binary.PutUvarint(b, guildStats.Sequence())

		stats = new(GuildStats)
		json.Unmarshal(guildStats.Get(b), stats)

		return nil
	})

	return stats
}

type GuildStats struct {
	TimeStamp *time.Time
	Channels  map[string]*ChannelStats
}

func (gst *GuildStats) getTotals() (totals *ChannelStats) {
	totals = &ChannelStats{Users: make(map[string]*UserStats)}

	for _, stats := range gst.Channels {
		totals = joinStatList(stats, totals)
	}
	return
}

type ChannelStats struct {
	LastMessageID string
	Users         map[string]*UserStats
}

func (ch *ChannelStats) getTotals() (totals *UserStats) {
	totals = &UserStats{
		EmojisReacted:  make(map[string]uint64),
		CharactersUsed: make(map[rune]uint64),
	}

	for _, stats := range ch.Users {
		totals = joinUserStats(stats, totals)
	}
	return
}

func joinStatList(l, r *ChannelStats) *ChannelStats {
	stats := &ChannelStats{Users: make(map[string]*UserStats)}

	for k, v := range l.Users {
		stats.Users[k] = v
	}

	for k, v := range r.Users {
		if data, ok := stats.Users[k]; ok {
			stats.Users[k] = joinUserStats(data, v)
		} else {
			stats.Users[k] = v
		}
	}

	return stats
}

type UserStats struct {
	MessagesSent             uint64            `json:"msgs"`
	LinksLinked              uint64            `json:"lnks"`
	ImagesSent               uint64            `json:"imgs"`
	FilesSent                uint64            `json:"files"`
	ShouldersShruggedOverTwo uint64            `json:"shrug"`
	AdrianCommands           uint64            `json:"acmd"`
	CelestineCommands        uint64            `json:"ccmd"`
	GregoryCommands          uint64            `json:"gcmd"`
	CharactersUsed           map[rune]uint64   `json:"chuse"`
	EmojisReacted            map[string]uint64 `json:"reac"`
}

func joinUserStats(l, r *UserStats) *UserStats {
	join := &UserStats{
		MessagesSent:             l.MessagesSent + r.MessagesSent,
		LinksLinked:              l.LinksLinked + r.LinksLinked,
		ImagesSent:               l.ImagesSent + r.ImagesSent,
		FilesSent:                l.FilesSent + r.FilesSent,
		ShouldersShruggedOverTwo: l.ShouldersShruggedOverTwo + r.ShouldersShruggedOverTwo,
		AdrianCommands:           l.AdrianCommands + r.AdrianCommands,
		CelestineCommands:        l.CelestineCommands + r.CelestineCommands,
		GregoryCommands:          l.GregoryCommands + r.GregoryCommands,
		EmojisReacted:            make(map[string]uint64),
		CharactersUsed:           make(map[rune]uint64),
	}

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		for k, v := range l.CharactersUsed {
			join.CharactersUsed[k] = v
		}

		for k, v := range r.CharactersUsed {
			join.CharactersUsed[k] = join.CharactersUsed[k] + v
		}
		wg.Done()
	}()

	go func() {
		for k, v := range l.EmojisReacted {
			join.EmojisReacted[k] = v
		}

		for k, v := range r.EmojisReacted {
			join.EmojisReacted[k] = join.EmojisReacted[k] + v
		}
		wg.Done()
	}()

	wg.Wait()
	return join
}
