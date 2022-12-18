package pkg

import (
	"fmt"
	"github.com/dlclark/regexp2"
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
			count: SearchCommitMaxLimit,
		},
		{
			repository: djangoRepository,
			options: &SearchCommitsOptions{
				Since:    time.UnixMilli(0),
				Until:    time.Now(),
				Interval: time.Hour * 24 * 30,
				Limit:    0,
			},
			count: SearchCommitMaxLimit,
		},
		{
			repository: djangoRepository,
			options: &SearchCommitsOptions{
				Since:    time.Date(2012, 4, 1, 0, 0, 0, 0, time.UTC),
				Until:    time.Date(2012, 5, 1, 0, 0, 0, 0, time.UTC),
				Interval: time.Hour * 24,
				Limit:    7,
			},
			count: 7,
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
	t.Setenv(KunitoriUseGitCommandProvidedKey, "")

	djangoRepository := openTestRepository("django")
	djangoHeadCommit := getHeadCommit(djangoRepository)

	testCases := []struct {
		repository *git.Repository
		commit     *object.Commit
		options    *CountLinesOption
		results    []*CountLinesResult
	}{
		{
			repository: djangoRepository,
			commit:     djangoHeadCommit,
			options: &CountLinesOption{
				Filters:       []*regexp2.Regexp{},
				AuthorRegexes: []AuthorRegex{},
			},
		},
		{
			repository: djangoRepository,
			commit:     djangoHeadCommit,
			options: &CountLinesOption{
				Filters: []*regexp2.Regexp{
					regexp2.MustCompile("^setup\\.py$", 0),
				},
				AuthorRegexes: []AuthorRegex{},
			},
			results: []*CountLinesResult{
				{
					Filter: regexp2.MustCompile("^setup\\.py$", 0),
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
					NameByAuthor: map[string]string{
						"adrian@masked.com":       "Adrian Holovaty",
						"carl@masked.com":         "Carl Meyer",
						"carlton@masked.com":      "Carlton Gibson",
						"florian@masked.com":      "Florian Apolloner",
						"jacob@masked.com":        "Jacob Kaplan-Moss",
						"jon.dufresne@masked.com": "Jon Dufresne",
						"ops@masked.com":          "django-bot",
						"timograham@masked.com":   "Tim Graham",
					},
					MatchedFiles: []string{
						"setup.py",
					},
				},
			},
		},
		{
			repository: djangoRepository,
			commit:     djangoHeadCommit,
			options: &CountLinesOption{
				Filters: []*regexp2.Regexp{
					regexp2.MustCompile("^setup\\.py$", 0),
				},
				AuthorRegexes: []AuthorRegex{
					{
						Condition: regexp2.MustCompile("^a.+", 0),
						Author:    "aGroup",
					},
					{
						Condition: regexp2.MustCompile("^c.+", 0),
						Author:    "cGroup",
					},
					{
						Condition: regexp2.MustCompile("^[^f].+", 0),
						Author:    "otherGroup",
					},
				},
			},
			results: []*CountLinesResult{
				{
					Filter: regexp2.MustCompile("^setup\\.py$", 0),
					// https://github.com/django/django/blame/a1bcdc94da6d597c51b4eca0411a97a6460b482e/setup.py
					LinesByAuthor: map[string]int{
						"aGroup":             1,
						"cGroup":             41,
						"florian@masked.com": 3,
						"otherGroup":         10,
					},
					NameByAuthor: map[string]string{
						"aGroup":             "Adrian Holovaty",
						"cGroup":             "Carlton Gibson",
						"florian@masked.com": "Florian Apolloner",
						"otherGroup":         "Tim Graham",
					},
					MatchedFiles: []string{
						"setup.py",
					},
				},
			},
		},
		{
			repository: djangoRepository,
			commit:     djangoHeadCommit,
			options: &CountLinesOption{
				Filters: []*regexp2.Regexp{
					regexp2.MustCompile("^django/__(init|main)__\\.py$|^django/shortcuts\\.py$", 0),
				},
				AuthorRegexes: []AuthorRegex{},
			},
			results: []*CountLinesResult{
				{
					Filter: regexp2.MustCompile("^django/__(init|main)__\\.py$|^django/shortcuts\\.py$", 0),
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
					NameByAuthor: map[string]string{
						"alex.gaynor@masked.com":      "Alex Gaynor",
						"anton.samarchyan@masked.com": "Anton Samarchyan",
						"aymeric.augustin@masked.com": "Aymeric Augustin",
						"carlton.gibson@masked.com":   "Carlton Gibson",
						"claude@masked.com":           "Claude Paroz",
						"dilyanpalauzov@masked.com":   "Дилян Палаузов",
						"dizballanze@masked.com":      "dizballanze",
						"info@masked.com":             "Martin Thoma",
						"marten.knbk@masked.com":      "Marten Kenbeek",
						"ops@masked.com":              "django-bot",
						"ryan@masked.com":             "Ryan Hiebert",
						"smithdc@masked.com":          "David Smith",
						"timograham@masked.com":       "Tim Graham",
						"vytis.banaitis@masked.com":   "Vytis Banaitis",
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
		if index > 1 && testing.Short() {
			t.SkipNow()
		}

		t.Run(fmt.Sprintf("case_%v", index), func(t *testing.T) {
			results, err := CountLines(testCase.repository, testCase.commit, testCase.options)
			assert.NoError(t, err)
			assert.Equal(t, len(testCase.results), len(results))

			for index, result := range results {
				t.Run(fmt.Sprintf("result_%v", index), func(t *testing.T) {
					expected := testCase.results[index]
					assert.Equal(t, expected.Filter.String(), result.Filter.String())

					maskedLinesByAuthor := map[string]int{}
					for key, value := range result.LinesByAuthor {
						maskedKey := maskRegex.ReplaceAllString(key, "@masked.com")
						maskedLinesByAuthor[maskedKey] = value
					}
					assert.Equal(t, expected.LinesByAuthor, maskedLinesByAuthor)

					maskedNameByAuthor := map[string]string{}
					for key, value := range result.NameByAuthor {
						maskedKey := maskRegex.ReplaceAllString(key, "@masked.com")
						maskedNameByAuthor[maskedKey] = value
					}
					assert.Equal(t, expected.NameByAuthor, maskedNameByAuthor)
				})
			}
		})
	}
}

func TestCountLines__diff(t *testing.T) {
	t.Setenv(KunitoriUseGitCommandProvidedKey, "")

	djangoRepository := openTestRepository("django")
	djangoHeadCommit := getHeadCommit(djangoRepository)

	option := CountLinesOption{
		Filters: []*regexp2.Regexp{
			regexp2.MustCompile("^\\w+\\.\\w+$", 0),
		},
		AuthorRegexes: []AuthorRegex{},
	}

	goGitResult, err := CountLines(djangoRepository, djangoHeadCommit, &option)
	assert.NoError(t, err)

	t.Setenv(KunitoriUseGitCommandProvidedKey, "1")
	gitCommandResult, err := CountLines(djangoRepository, djangoHeadCommit, &option)
	assert.NoError(t, err)

	assert.NotEqual(t, goGitResult, gitCommandResult)
}

func TestBlameWithGitCommand(t *testing.T) {
	t.Setenv(KunitoriUseGitCommandProvidedKey, "1")

	djangoRepository := openTestRepository("django")
	djangoHeadCommit := getHeadCommit(djangoRepository)

	option := CountLinesOption{
		Filters: []*regexp2.Regexp{
			regexp2.MustCompile("^setup\\.\\w+$", 0),
		},
		AuthorRegexes: []AuthorRegex{},
	}

	maskRegex := regexp.MustCompile("@.+$")

	results, err := CountLines(djangoRepository, djangoHeadCommit, &option)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))

	result := results[0]
	expected := CountLinesResult{
		Filter: option.Filters[0],
		LinesByAuthor: map[string]int{
			"adrian@masked.com":           4,
			"bruno@masked.com":            1,
			"carl@masked.com":             36,
			"carlton.gibson@masked.com":   11,
			"carlton@masked.com":          4,
			"felisiak.mariusz@masked.com": 4,
			"florian@masked.com":          4,
			"jacob@masked.com":            1,
			"jon.dufresne@masked.com":     47,
			"ops@masked.com":              6,
			"smithdc@masked.com":          1,
			"timograham@masked.com":       7,
			"ville.skytta@masked.com":     1,
		},
		NameByAuthor: map[string]string{
			"adrian@masked.com":           "Adrian Holovaty",
			"bruno@masked.com":            "Bruno Renié",
			"carl@masked.com":             "Carl Meyer",
			"carlton.gibson@masked.com":   "Carlton Gibson",
			"carlton@masked.com":          "Carlton Gibson",
			"felisiak.mariusz@masked.com": "Mariusz Felisiak",
			"florian@masked.com":          "Florian Apolloner",
			"jacob@masked.com":            "Jacob Kaplan-Moss",
			"jon.dufresne@masked.com":     "Jon Dufresne",
			"ops@masked.com":              "django-bot",
			"smithdc@masked.com":          "David Smith",
			"timograham@masked.com":       "Tim Graham",
			"ville.skytta@masked.com":     "Ville Skyttä",
		},
		MatchedFiles: []string{
			"setup.cfg",
			"setup.py",
		},
	}
	assert.Equal(t, expected.Filter, result.Filter)

	maskedLinesByAuthor := map[string]int{}
	for key, value := range result.LinesByAuthor {
		maskedKey := maskRegex.ReplaceAllString(key, "@masked.com")
		maskedLinesByAuthor[maskedKey] = value
	}
	assert.Equal(t, expected.LinesByAuthor, maskedLinesByAuthor)

	maskedNameByAuthor := map[string]string{}
	for key, value := range result.NameByAuthor {
		maskedKey := maskRegex.ReplaceAllString(key, "@masked.com")
		maskedNameByAuthor[maskedKey] = value
	}
	assert.Equal(t, expected.NameByAuthor, maskedNameByAuthor)
}

func TestIsUseGitCommandProvided(t *testing.T) {
	t.Run("is set", func(t *testing.T) {
		t.Setenv(KunitoriUseGitCommandProvidedKey, "1")
		assert.True(t, IsUseGitCommandProvided())
	})

	t.Run("is not set", func(t *testing.T) {
		t.Setenv(KunitoriUseGitCommandProvidedKey, "")
		assert.False(t, IsUseGitCommandProvided())
	})
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

func testOutPath(name string) string {
	return filepath.Join(rootPath(), "test", "out", name)
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
