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
		RepositoryPath: testDataPath("yokuwakaru-grpc"),
		Region:         "__TEST",
		SearchCommitsOptions: &SearchCommitsOptions{
			Since:    time.Date(2020, 10, 1, 0, 0, 0, 0, time.UTC),
			Until:    time.Date(2022, 10, 30, 23, 59, 59, 0, time.UTC),
			Interval: time.Hour * 24 * 7,
			Limit:    1,
		},
		CountLinesOption: &CountLinesOption{
			Filters: []*regexp2.Regexp{
				regexp2.MustCompile("\\.go$", 0),
			},
			AuthorRegexes: []AuthorRegex{},
		},
	}

	expected := GenerateResult{
		Repository:  "https://github.com/yktakaha4/yokuwakaru-grpc",
		Source:      "github",
		GeneratedAt: time.Time{},
		Commits: []GenerateResultCommit{
			{
				Hash:        "2fa8fa83724e394a098890c40cc324fa90b080b5",
				CommittedAt: time.Time{},
				LineCounts: []GenerateResultCommitLineCount{
					{
						FilterRegex: "\\.go$",
						FileCount:   4,
						Areas: []GenerateResultCommitLineCountArea{
							{
								Name:        "Area30",
								Size:        30,
								Ratio:       0.5,
								AuthorEmail: "20282867+yktakaha4@users.noreply.github.com",
								AuthorRank:  1,
							},
							{
								Name:        "Area20",
								Size:        20,
								Ratio:       0.333,
								AuthorEmail: "20282867+yktakaha4@users.noreply.github.com",
								AuthorRank:  1,
							},
							{
								Name:        "Area10",
								Size:        10,
								Ratio:       0.167,
								AuthorEmail: "20282867+yktakaha4@users.noreply.github.com",
								AuthorRank:  1,
							},
						},
						Authors: []GenerateResultCommitLineCountAuthor{
							{
								Email:       "20282867+yktakaha4@users.noreply.github.com",
								Name:        "Yuuki Takahashi",
								GitHubLogin: "kunitori",
								LineCount:   495,
								Rank:        1,
							},
						},
					},
				},
			},
		},
	}

	result, err := Generate(&options)
	assert.NoError(t, err)

	expected.GeneratedAt = result.GeneratedAt
	expected.Commits[0].CommittedAt = result.Commits[0].CommittedAt
	assert.Equal(t, &expected, result)

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
