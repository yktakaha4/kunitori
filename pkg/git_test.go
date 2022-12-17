package pkg

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
)

func TestCloneRepository(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	testCases := []struct {
		url string
	}{
		{
			url: "https://github.com/yktakaha4/eduterm.git",
		},
		{
			url: "git@github.com:yktakaha4/eduterm.git",
		},
	}

	for index, testCase := range testCases {
		t.Run(fmt.Sprintf("case_%v", index), func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "TestCloneRepository")
			assert.NoError(t, err)
			defer func(path string) {
				err := os.RemoveAll(path)
				assert.NoError(t, err)
			}(tempDir)

			repository, err := CloneRepository(testCase.url, tempDir)
			assert.NoError(t, err)
			assert.NotNil(t, repository)
		})
	}
}

func TestOpenRepository(t *testing.T) {
	testCases := []struct {
		path string
		head string
	}{
		{
			path: testDataPath("django"),
			head: "a1bcdc94da6d597c51b4eca0411a97a6460b482e",
		},
	}

	for index, testCase := range testCases {
		t.Run(fmt.Sprintf("case_%v", index), func(t *testing.T) {
			repository, err := OpenRepository(testCase.path)
			assert.NoError(t, err)
			assert.NotNil(t, repository)

			reference, err := repository.Head()
			assert.NoError(t, err)

			assert.Equal(t, reference.Hash().String(), testCase.head)
		})
	}
}

func TestSearchCommits(t *testing.T) {
	djangoRepository := openTestRepository("django")
	maxLimit := 50

	testCases := []struct {
		repository *git.Repository
		options    *SearchCommitsOptions
		count      int
	}{
		{
			repository: djangoRepository,
			options: &SearchCommitsOptions{
				Since:    time.UnixMilli(0),
				Until:    time.Now(),
				Interval: 0,
				Limit:    0,
			},
			count: maxLimit,
		},
		{
			repository: djangoRepository,
			options: &SearchCommitsOptions{
				Since:    time.UnixMilli(0),
				Until:    time.Now(),
				Interval: time.Hour * 24 * 30,
				Limit:    0,
			},
			count: maxLimit,
		},
		{
			repository: djangoRepository,
			options: &SearchCommitsOptions{
				Since:    time.Date(2012, 4, 1, 0, 0, 0, 0, time.UTC),
				Until:    time.Date(2012, 5, 1, 0, 0, 0, 0, time.UTC),
				Interval: time.Hour * 24,
				Limit:    20,
			},
			count: 20,
		},
		{
			repository: djangoRepository,
			options: &SearchCommitsOptions{
				Since: time.Date(2021, 4, 1, 12, 34, 56, 0, time.UTC),
				Until: time.Date(2012, 4, 1, 12, 34, 56, 0, time.UTC),
			},
			count: 0,
		},
	}

	for index, testCase := range testCases {
		t.Run(fmt.Sprintf("case_%v", index), func(t *testing.T) {
			commits, err := SearchCommits(testCase.repository, testCase.options)
			assert.NoError(t, err)
			assert.NotNil(t, commits)

			assert.Equal(t, testCase.count, len(commits))

			if testCase.options.Limit > 0 {
				assert.GreaterOrEqual(t, testCase.options.Limit, len(commits))
			}

			for index, commit := range commits {
				commitWhen := commit.Author.When.UTC()
				assert.True(t, commitWhen.After(testCase.options.Since))
				assert.True(t, commitWhen.Before(testCase.options.Until))
				if index > 0 {
					previousWhen := commits[index-1].Author.When.UTC()
					assert.True(t, previousWhen.Sub(commitWhen) >= testCase.options.Interval)
				}
			}
		})
	}
}

func TestCountLines(t *testing.T) {
	djangoRepository := openTestRepository("django")
	djangoHeadCommit := getHeadCommit(djangoRepository)

	testCases := []struct {
		commit  *object.Commit
		options *CountLinesOption
		results []*CountLinesResult
	}{
		{
			commit: djangoHeadCommit,
			options: &CountLinesOption{
				Filters:       []regexp.Regexp{},
				AuthorRegexes: map[string]regexp.Regexp{},
			},
		},
		{
			commit: djangoHeadCommit,
			options: &CountLinesOption{
				Filters: []regexp.Regexp{
					*regexp.MustCompile("^setup\\.py$"),
				},
				AuthorRegexes: map[string]regexp.Regexp{},
			},
			results: []*CountLinesResult{
				{
					Filter: *regexp.MustCompile("^setup\\.py$"),
					// https://github.com/django/django/blame/a1bcdc94da6d597c51b4eca0411a97a6460b482e/setup.py
					LinesByAuthor: map[string]int{
						"adrian@masked.com":       1,
						"carl@masked.com":         37,
						"carlton@masked.com":      4,
						"florian@masked.com":      3,
						"jacob@masked.com":        1,
						"jon.dufresne@masked.com": 2,
						"ops@masked.com":          6,
						"timograham@masked.com":   1,
					},
					MatchedFiles: []string{
						"setup.py",
					},
				},
			},
		},
		{
			commit: djangoHeadCommit,
			options: &CountLinesOption{
				Filters: []regexp.Regexp{
					*regexp.MustCompile("^django/__(init|main)__\\.py$|^django/shortcuts\\.py$"),
				},
				AuthorRegexes: map[string]regexp.Regexp{},
			},
			results: []*CountLinesResult{
				{
					Filter: *regexp.MustCompile("^django/__(init|main)__\\.py$|^django/shortcuts\\.py$"),
					// https://github.com/django/django/tree/a1bcdc94da6d597c51b4eca0411a97a6460b482e/django
					LinesByAuthor: map[string]int{
						"alex.gaynor@masked.com":      79,
						"anton.samarchyan@masked.com": 1,
						"aymeric.augustin@masked.com": 6,
						"carlton.gibson@masked.com":   1,
						"claude@masked.com":           33,
						"dilyanpalauzov@masked.com":   2,
						"dizballanze@masked.com":      3,
						"info@masked.com":             2,
						"marten.knbk@masked.com":      8,
						"ops@masked.com":              29,
						"ryan@masked.com":             9,
						"smithdc@masked.com":          3,
						"timograham@masked.com":       11,
						"vytis.banaitis@masked.com":   1,
					},
					MatchedFiles: []string{
						"django/__init__.py",
						"django/__main__.py",
						"django/shortcuts.py",
					},
				},
			},
		},
	}

	maskRegex := regexp.MustCompile("@.+$")

	for index, testCase := range testCases {
		t.Run(fmt.Sprintf("case_%v", index), func(t *testing.T) {
			results, err := CountLines(testCase.commit, testCase.options)
			assert.NoError(t, err)
			assert.Equal(t, len(testCase.results), len(results))

			for index, result := range results {
				t.Run(fmt.Sprintf("result_%v", index), func(t *testing.T) {
					expected := testCase.results[index]
					assert.Equal(t, expected.Filter, result.Filter)

					maskedLinesByAuthor := map[string]int{}
					for key, value := range result.LinesByAuthor {
						maskedKey := maskRegex.ReplaceAllString(key, "@masked.com")
						maskedLinesByAuthor[maskedKey] = value
					}

					assert.Equal(t, expected.LinesByAuthor, maskedLinesByAuthor)
				})
			}
		})
	}
}

func rootPath() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	relative := filepath.Dir(wd)

	absolute, err := filepath.Abs(relative)
	if err != nil {
		panic(err)
	}

	return absolute
}

func testDataPath(name string) string {
	return filepath.Join(rootPath(), "test", "testdata", name)
}

func openTestRepository(name string) *git.Repository {
	path := testDataPath(name)

	repository, err := git.PlainOpen(path)
	if err != nil {
		panic(err)
	}
	return repository
}

func getHeadCommit(repository *git.Repository) *object.Commit {
	reference, err := repository.Head()
	if err != nil {
		panic(err)
	}

	commit, err := repository.CommitObject(reference.Hash())
	if err != nil {
		panic(err)
	}

	return commit
}
