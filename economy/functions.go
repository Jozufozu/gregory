package economy

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jozufozu/gregory/util"
	"log"
)

var currencyKey = []byte{23}

func CheckBalance(who string) uint64 {
	var balance uint64
	err := Data.View(func(tx *bolt.Tx) error {
		accounts := tx.Bucket([]byte("accounts"))

		user := accounts.Bucket([]byte(who))

		if user == nil {
			return errors.New("user does not have money")
		}

		balance = fromBytes(user.Get(currencyKey))

		return nil
	})

	if err != nil {
		return 0
	}

	return balance
}

func Transfer(from, to string, amount uint64) (transaction *util.Transaction) {
	err := Data.Update(func(tx *bolt.Tx) error {
		accounts := tx.Bucket([]byte("accounts"))
		transactions := tx.Bucket([]byte("transactions"))

		fromBucket, err := accounts.CreateBucketIfNotExists([]byte(from))
		toBucket, err := accounts.CreateBucketIfNotExists([]byte(to))

		if err != nil {
			return err
		}

		fromCoin := fromBytes(fromBucket.Get(currencyKey))
		toCoin := fromBytes(toBucket.Get(currencyKey))

		fromCoin -= amount
		toCoin += amount

		fromBucket.Put(currencyKey, toBytes(fromCoin))
		toBucket.Put(currencyKey, toBytes(toCoin))

		nextSequence, _ := transactions.NextSequence()

		transaction = util.NewTransaction(from, to, amount)
		transactions.Put(toBytes(nextSequence), transaction.Encode())

		return nil
	})

	if err != nil {
		log.Println(err)
		return nil
	}

	fmt.Printf("%s sent %s ₣%v\n", from, to, amount)

	return
}

func Introduce(amount uint64) {
	Data.Update(func(tx *bolt.Tx) error {
		accounts := tx.Bucket([]byte("accounts"))

		bank, err := accounts.CreateBucketIfNotExists([]byte("387810222556708865"))

		if err != nil {
			return err
		}

		bankCoin := fromBytes(bank.Get(currencyKey))
		bankCoin += amount
		bank.Put(currencyKey, toBytes(bankCoin))
		bytes, _ := json.Marshal(util.NewDate())
		bank.Put([]byte("last"), bytes)

		fmt.Printf("Added ₣%v to the economy\n", amount)

		return nil
	})
}
