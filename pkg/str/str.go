package str

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	mrand "math/rand"
	"strconv"
	"strings"
	"text/template"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	// UpperAlphabet upper alphabet chars
	UpperAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// LowerAlphabet lower alphabet chars
	LowerAlphabet = "abcdefghijklmnopqrstuvwxyz"
	// Alphabet alphabet chars with upper and lower
	Alphabet = UpperAlphabet + LowerAlphabet
	// Numerals numeral chars
	Numerals = "1234567890"
	// Alphanumeric alphabet and numeral chars
	Alphanumeric = Alphabet + Numerals
	// ASCII ascii code chars
	ASCII = Alphanumeric + "~!@#$%^&*()-_+={}[]\\|<,>.?/\"';:`"
	// Base57 base57 chars
	Base57 = "0123456789abcdefghijkmnopqrstvwxyzABCDEFGHJKLMNPQRSTVWXYZ"
)

// RandString inspired from https://github.com/jmcvetta/randutil/blob/master/randutil.go
func RandString(length int, charset string) string {
	str := make([]byte, length)
	if charset == "" {
		charset = Alphabet
	}
	charlen := big.NewInt(int64(len(charset)))
	for i := 0; i < length; i++ {
		v, _ := rand.Int(rand.Reader, charlen)
		str[i] = charset[int(v.Int64())]
	}
	return string(str)
}

// ParseInt parse decimal string
func ParseInt(s string) (int64, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	return v, err
}

// ParsePositiveInt parse positive decimal string
func ParsePositiveInt(s string) (int64, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err == nil && v <= 0 {
		err = fmt.Errorf("value expect positive, got %d", v)
	}
	return v, err
}

// MustParseInt must parse decimal string
func MustParseInt(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// MustParseBool must parse bool string
func MustParseBool(s string) bool {
	v, _ := strconv.ParseBool(s)
	return v
}

// UUIDStr return uuid string
func UUIDStr() string {
	return strings.Replace(uuid.NewV4().String(), "-", "", -1)
}

// Md5Str md5 encode string s
func Md5Str(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// Base64EncodeStr base64 encode string s
func Base64EncodeStr(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64DecodeStr base64 decode string s
func Base64DecodeStr(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// HmacSha1EncodeStr hmac-sha1 encode string s
func HmacSha1EncodeStr(key, s string) string {
	m := hmac.New(sha1.New, []byte(key))
	_, _ = m.Write([]byte(s))
	return string(m.Sum(nil))
}

// ParseAddressToHostPort parse HOST:PORT
func ParseAddressToHostPort(addr string) (string, int, error) {
	fields := strings.Split(addr, ":")
	if len(fields) < 2 {
		return "", -1, fmt.Errorf("invalid address format, not HOST:PORT")
	}
	host := fields[0]
	port, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return "", -1, fmt.Errorf("invalid port: %v", err)
	}
	return host, int(port), nil
}

// ExecuteTemplate simplify template execute
func ExecuteTemplate(tpl string, v interface{}) (string, error) {
	tp, err := template.New("tpl").Parse(tpl)
	if err != nil {
		return "", fmt.Errorf("parse template failed: %v", err)
	}

	buf := &bytes.Buffer{}
	if err = tp.Execute(buf, v); err != nil {
		return "", fmt.Errorf("template execute: %v", err)
	}
	return buf.String(), nil
}

var serial uint16

// Token32 generate 32 unique token
func Token32() string {
	serial++
	buf := make([]byte, 16)
	binary.BigEndian.PutUint16(buf[0:2], serial)                         // serial
	binary.BigEndian.PutUint64(buf[2:10], uint64(time.Now().UnixNano())) // nano timestamp
	_, _ = rand.Read(buf[10:16])

	buf[0], buf[10] = (buf[0]&0x66)+(buf[10]&0x99), (buf[10]&0x66)+(buf[0]&0x99)
	buf[1], buf[11] = (buf[1]&0x66)+(buf[11]&0x99), (buf[11]&0x66)+(buf[1]&0x99)
	buf[2], buf[12] = (buf[2]&0x66)+(buf[12]&0x99), (buf[12]&0x66)+(buf[2]&0x99)
	buf[3], buf[13] = (buf[3]&0x66)+(buf[13]&0x99), (buf[13]&0x66)+(buf[3]&0x99)
	buf[4], buf[14] = (buf[4]&0x66)+(buf[14]&0x99), (buf[14]&0x66)+(buf[4]&0x99)
	buf[5], buf[15] = (buf[5]&0x66)+(buf[15]&0x99), (buf[15]&0x66)+(buf[5]&0x99)
	buf[6], buf[12] = (buf[6]&0x99)+(buf[12]&0x66), (buf[12]&0x99)+(buf[6]&0x66)
	buf[7], buf[13] = (buf[7]&0x99)+(buf[13]&0x66), (buf[13]&0x99)+(buf[7]&0x66)
	buf[8], buf[14] = (buf[8]&0x99)+(buf[14]&0x66), (buf[14]&0x99)+(buf[8]&0x66)
	buf[9], buf[15] = (buf[9]&0x99)+(buf[15]&0x66), (buf[15]&0x99)+(buf[9]&0x66)
	buf2 := make([]byte, 32)
	hex.Encode(buf2, buf)
	return string(buf2)
}

// LengthOfStringInMinMaxOrZero reports whether length of s is in [min, max] or zero
func LengthOfStringInMinMaxOrZero(s string, min, max int) bool {
	l := len(s)
	return l == 0 || (l >= min && l <= max)
}

// LengthOfStringInMinMax reports whether length of s is in [min, max]
func LengthOfStringInMinMax(s string, min, max int) bool {
	l := len(s)
	return l >= min && l <= max
}

// ContainsRune reports whether the Unicode code point r is within s.
func ContainsRune(s string, v rune) bool {
	return strings.ContainsRune(s, v)
}

// ConstructedByCharsets reports whether s is constructed by charsets whose number is large than minMatchedCharsets
func ConstructedByCharsets(s string, minMatchedCharsets int, charsets ...string) bool {
	counts := make([]int, len(charsets))
	for _, c := range s {
		found := false
		for idx, charset := range charsets {
			if ContainsRune(charset, c) {
				found = true
				counts[idx]++
			}
		}
		if !found {
			return false
		}
	}
	count := 0
	for _, c := range counts {
		if c > 0 {
			count++
		}
	}
	return count >= minMatchedCharsets
}

func init() {
	// use UnixNano initialize global rand
	mrand.Seed(time.Now().UnixNano())
	serial = uint16(mrand.Uint32())
}
