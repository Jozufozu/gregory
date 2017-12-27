package util

import (
	"encoding/json"
	"log"
	"time"
)

type Transaction struct {
	Time   time.Time `json:"time"`
	From   string    `json:"from"`
	To     string    `json:"to"`
	Amount uint64    `json:"amount"`
}

func NewTransaction(from, to string, amount uint64) *Transaction {
	return &Transaction{
		Time:   time.Now(),
		From:   from,
		To:     to,
		Amount: amount,
	}
}

func (t *Transaction) Encode() []byte {
	bytes, err := json.Marshal(t)
	if err != nil {
		log.Println(err)
		return nil
	}
	return bytes
}
