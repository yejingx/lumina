package str

import (
	"crypto/rand"
	"math/big"
)

const (
	UpperAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LowerAlphabet = "abcdefghijklmnopqrstuvwxyz"
	Alphabet      = UpperAlphabet + LowerAlphabet
	Numerals      = "1234567890"
	Alphanumeric  = Alphabet + Numerals
)

func RandStr(length int, charset string) string {
	str := make([]byte, length)

	charlen := big.NewInt(int64(len(charset)))
	for i := range length {
		v, _ := rand.Int(rand.Reader, charlen)
		str[i] = charset[int(v.Int64())]
	}

	return string(str)
}

func GenDeviceId(length int) string {
	return RandStr(length, UpperAlphabet+Numerals)
}

func GenToken(length int) string {
	return RandStr(length, Alphanumeric)
}
