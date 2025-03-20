package random

import (
	"errors"
	"fmt"
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

var ErrDivideByZero = errors.New("divide by zero")

func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("boo: %w", ErrDivideByZero)
	}
	return a / b, nil
}
