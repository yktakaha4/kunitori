package pkg

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestRenderChartHtml(t *testing.T) {
	generateResult := GenerateResult{
		Repository:  "dummy",
		Source:      "unknown",
		GeneratedAt: time.Now().UTC(),
		Commits: []GenerateResultCommit{
			{
				Hash:        "dummy-hash",
				CommittedAt: time.Now().UTC(),
				LineCounts: []GenerateResultCommitLineCount{
					{
						FilterRegex: "dummy regex",
						FileNames: []string{
							"dummy-file-1",
							"dummy-file-2",
						},
						Areas: []GenerateResultCommitLineCountArea{
							{
								Name:        "dummy-area",
								Size:        123,
								Ratio:       0.123,
								AuthorEmail: "dummy@example.com",
								AuthorRank:  1,
							},
						},
						Authors: []GenerateResultCommitLineCountAuthor{
							{
								Email:       "dummy2@example.com",
								Name:        "dummy-name",
								GitHubLogin: "dummy-login",
								LineCount:   123,
							},
						},
					},
				},
			},
		},
	}

	html, err := RenderChartHtml(&generateResult)
	assert.NoError(t, err)

	chartHtmlFilePath := testOutPath("chart.html")

	file, err := os.Create(chartHtmlFilePath)
	assert.NoError(t, err)

	_, err = file.Write([]byte(html))
	assert.NoError(t, err)
}
