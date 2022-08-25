package random

import (
	"math/rand"
	"strconv"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandInt() int64 {
	var (
		min, max int64
	)
	min, max = 100000, 999999
	return min + rand.Int63n(max-min+1)
}

func RandString() string {
	randstring := []byte{}
	l := len(alphabet)
	for i := 0; i < 6; i++ {
		position := rand.Intn(l)
		randstring = append(randstring, alphabet[position])
	}
	return string(randstring)
}

func RandPhoneNumber() string {
	return strconv.Itoa(int(RandInt()))
}

func RandTrainingTime() time.Time {

	return time.Now().AddDate(0, 0, rand.Intn(7)+1)
}
