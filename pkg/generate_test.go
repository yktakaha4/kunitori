package pkg

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerate(t *testing.T) {

}

func TestGetSource(t *testing.T) {
	testCases := []struct {
		value  string
		source string
	}{
		{
			value:  "https://github.com/yktakaha4/eduterm.git",
			source: "github",
		},
		{
			value:  "/usr/home/repos",
			source: "unknown",
		},
		{
			value:  "",
			source: "unknown",
		},
	}

	for index, testCase := range testCases {
		t.Run(fmt.Sprintf("case_%v", index), func(t *testing.T) {
			source := GetSource(testCase.value)
			assert.Equal(t, testCase.source, source)
		})
	}
}
