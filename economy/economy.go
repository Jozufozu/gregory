package economy

import (
	"encoding/binary"
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/jozufozu/gregory/commands"
	"github.com/jozufozu/gregory/util"
	"log"
	"time"
)

var Data *bolt.DB

func checkAndUpdate() {
	var (
		b    bool
		days int
	)
	Data.View(func(tx *bolt.Tx) error {
		accounts := tx.Bucket([]byte("accounts"))

		bank := accounts.Bucket([]byte("387810222556708865"))

		get := bank.Get([]byte("last"))

		if get == nil {
			b = true
			return nil
		}

		last := new(util.Date)
		json.Unmarshal(get, last)

		date := util.NewDate()
		b = date.NewerThan(last)
		days = date.DaysSince(last)

		return nil
	})

	if b {
		Introduce(1000)
	}
}

func fromBytes(data []byte) uint64 {
	u, _ := binary.Uvarint(data)
	return u
}

func toBytes(coin uint64) []byte {
	b := make([]byte, 8)
	binary.PutUvarint(b, coin)
	return b
}

func init() {
	db, err := bolt.Open("econ.db", 0600, nil)

	if err != nil {
		log.Fatal(err)
	}

	Data = db

	Data.Update(func(tx *bolt.Tx) error {
		accounts, _ := tx.CreateBucketIfNotExists([]byte("accounts"))
		accounts.CreateBucketIfNotExists([]byte("387810222556708865"))
		tx.CreateBucketIfNotExists([]byte("transactions"))
		return nil
	})

	go func() {
		ticker := time.NewTicker(time.Minute * 5).C

		checkAndUpdate()

		for e := range ticker {
			e.Unix()
			checkAndUpdate()
		}
	}()

	commands.AddCommand(&commands.Command{
		Aliases:     []string{"balance"},
		Action:      Balance,
		Usage:       "",
		Description: "How much stuff you got",
	})

	commands.AddCommand(&commands.Command{
		Aliases:     []string{"pay"},
		Action:      Pay,
		Usage:       "<user> <amount>",
		Description: "Give some of your hard earned cash to some loser.",
	})

	commands.AddCommand(&commands.Command{
		Aliases:     []string{"request"},
		Action:      Request,
		Usage:       "<user> <amount>",
		Description: "Ask somebody to give some of their hard earned cash to some loser.",
	})
}
