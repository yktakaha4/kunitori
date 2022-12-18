package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/dlclark/regexp2"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	t.Setenv(KunitoriSkipRequestGitHubApi, "yes")

	options := GenerateOptions{
		RepositoryPath: testDataPath("django"),
		Region:         "JP",
		SearchCommitsOptions: &SearchCommitsOptions{
			Since:    time.Date(2021, 10, 1, 0, 0, 0, 0, time.UTC),
			Until:    time.Date(2021, 10, 30, 23, 59, 59, 0, time.UTC),
			Interval: time.Hour * 24 * 7,
			Limit:    3,
		},
		CountLinesOption: &CountLinesOption{
			Filters: []*regexp2.Regexp{
				regexp2.MustCompile("^django/apps/.+\\.py$", 0),
				regexp2.MustCompile("^\\w+\\.rst$", 0),
			},
			AuthorRegexes: []AuthorRegex{},
		},
	}

	result, err := Generate(&options)
	assert.NoError(t, err)

	serialized, err := json.Marshal(result)
	assert.NoError(t, err)

	generateJsonFilePath := testOutPath("generate.json")

	file, err := os.Create(generateJsonFilePath)
	assert.NoError(t, err)

	_, err = file.Write(serialized)
	assert.NoError(t, err)
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
			value:  "git@github.com:go-git/go-git.git",
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

func TestGetRemoteUrl(t *testing.T) {
	testCases := []struct {
		value     string
		remoteUrl string
	}{
		{
			value:     "https://github.com/yktakaha4/eduterm.git",
			remoteUrl: "https://github.com/yktakaha4/eduterm",
		},
		{
			value:     "git@github.com:go-git/go-git.git",
			remoteUrl: "https://github.com/go-git/go-git",
		},
		{
			value:     "/usr/home/repos",
			remoteUrl: "/usr/home/repos",
		},
	}

	for index, testCase := range testCases {
		t.Run(fmt.Sprintf("case_%v", index), func(t *testing.T) {
			remoteUrl := GetRemoteUrl(testCase.value)
			assert.Equal(t, testCase.remoteUrl, remoteUrl)
		})
	}
}
