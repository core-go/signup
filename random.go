package signup

import (
	"github.com/google/uuid"
	"github.com/teris-io/shortid"
	"math"
	"math/rand"
	"strconv"
	"strings"
)

func times(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}

// Left left-pads the string with pad up to len runes
// len may be exceeded if
func padLeft(str string, length int, pad string) string {
	return times(pad, length-len(str)) + str
}

func generate(length int) string {
	max := int(math.Pow(float64(10), float64(length))) - 1
	return padLeft(strconv.Itoa(rand.Intn(max)), length, "0")
}

func shortId() (string, error) {
	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	if err != nil {
		return "", err
	}
	return sid.Generate()
}

func randomId() string {
	id := uuid.New()
	return strings.Replace(id.String(), "-", "", -1)
}