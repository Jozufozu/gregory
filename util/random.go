package util

import (
	"math/rand"
	"time"
)

var Rand = rand.New(rand.NewSource(time.Now().Unix()))
