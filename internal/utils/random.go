package random

import (
	"math/rand"
	"time"
)

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomString(size int) string {
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	bs := make([]byte, size)
	for i := 0; i < size; i++ {
		bs[i] = chars[rnd.Intn(len(chars))]
	}

	return string(bs)
}
