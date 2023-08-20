package indigoo

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"
)

func getStringBetween(s, start, end string) (string, error) {
	re := regexp.MustCompile(fmt.Sprintf(`(?s)%s(.*?)%s`, regexp.QuoteMeta(start), regexp.QuoteMeta(end)))
	match := re.FindStringSubmatch(s)

	if len(match) < 2 {
		return "", errors.New("no match found")
	}

	return match[1], nil
}

func generateCustomString(max int) string {
	rand.NewSource(time.Now().UnixNano())

	customRunes := make([]rune, max)
	for i := 0; i < max; i++ {
		randomIndex := rand.Intn(len(table))
		customRunes[i] = table[randomIndex]
	}

	return string(customRunes)
}

var table = [...]rune{
	'1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
}
