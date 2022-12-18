package pkg

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

type GenerateOptions struct {
	RepositoryUrl        string
	RepositoryPath       string
	Region               string
	SearchCommitsOptions *SearchCommitsOptions
	CountLinesOption     *CountLinesOption
}

type GenerateResultCommitLineCountAuthor struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	GitHubLogin string `json:"gitHubLogin"`
	LineCount   int    `json:"lineCount"`
	Rank        int    `json:"rank"`
}

type GenerateResultCommitLineCountArea struct {
	Name        string  `json:"name"`
	Size        float64 `json:"size"`
	Ratio       float64 `json:"ratio"`
	AuthorEmail string  `json:"authorEmail"`
	AuthorRank  int     `json:"authorRank"`
}

type GenerateResultCommitLineCount struct {
	FilterRegex string                                `json:"filterRegex"`
	FileCount   int                                   `json:"fileCount"`
	Areas       []GenerateResultCommitLineCountArea   `json:"areas"`
	Authors     []GenerateResultCommitLineCountAuthor `json:"authors"`
}

type GenerateResultCommit struct {
	Hash        string                          `json:"hash"`
	CommittedAt time.Time                       `json:"committedAt"`
	LineCounts  []GenerateResultCommitLineCount `json:"lineCounts"`
}

type GenerateResult struct {
	Repository  string                 `json:"repository"`
	Source      string                 `json:"source"`
	GeneratedAt time.Time              `json:"generatedAt"`
	Commits     []GenerateResultCommit `json:"commits"`
}

func ShowSlowMessage() {
	if !IsGitHubAccessTokenProvided() {
		fmt.Println(fmt.Sprintf(
			"warning: If the environment variable %v is not set, API searches will be very slow.",
			GitHubAccessTokenKey,
		))
	}
	if !IsUseGitCommandProvided() {
		fmt.Println(fmt.Sprintf(
			"warning: If the environment variable %v is not set, blame operation will be very slow.",
			KunitoriUseGitCommandProvidedKey,
		))
	}
}

func Generate(options *GenerateOptions) (*GenerateResult, error) {
	var repository *git.Repository

	ShowSlowMessage()

	areaInfo, err := GetAreaInfo(options.Region)
	if err != nil {
		return nil, err
	}

	var repositoryLocation string
	if options.RepositoryUrl != "" {
		tempDir, err := os.MkdirTemp("", "TestCloneRepository")
		if err != nil {
			return nil, err
		}
		defer func(path string) {
			err := os.RemoveAll(path)
			if err != nil {
				log.Println(err)
			}
		}(tempDir)

		repositoryLocation = options.RepositoryUrl
		repositoryLocation, err = filepath.Abs(repositoryLocation)
		if err != nil {
			return nil, err
		}
		fmt.Println(fmt.Sprintf("open repository: url=%v", repositoryLocation))

		repository, err = CloneRepository(repositoryLocation, tempDir)
		if err != nil {
			return nil, err
		}
	} else if options.RepositoryPath != "" {
		repositoryLocation = options.RepositoryPath
		repositoryLocation, err = filepath.Abs(repositoryLocation)
		if err != nil {
			return nil, err
		}
		fmt.Println(fmt.Sprintf("open repository: path=%v", repositoryLocation))

		repository, err = OpenRepository(repositoryLocation)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("should specify url or path")
	}

	repositoryRemoteLocation, err := GetRemoteLocation(repository)
	if err != nil {
		fmt.Println(err)
	}
	if repositoryRemoteLocation == "" {
		repositoryRemoteLocation = repositoryLocation
	}

	fmt.Println(fmt.Sprintf("location: remote=%v", repositoryRemoteLocation))

	fmt.Println(fmt.Sprintf(
		"search commit: since=%v, until=%v, interval=%v, limit=%v",
		options.SearchCommitsOptions.Since,
		options.SearchCommitsOptions.Until,
		options.SearchCommitsOptions.Interval,
		options.SearchCommitsOptions.Limit,
	))

	commits, err := SearchCommits(repository, options.SearchCommitsOptions)
	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf("matched commits: count=%v", len(commits)))

	fmt.Println(fmt.Sprintf(
		"count group: filters=%v, authors=%v",
		len(options.CountLinesOption.Filters),
		len(options.CountLinesOption.AuthorRegexes),
	))

	gitHubLoginNameCache := map[string]*string{}

	resultCommits := make([]GenerateResultCommit, 0)
	for index, commit := range commits {
		fmt.Println(fmt.Sprintf(
			"count lines: progress=%v/%v, hash=%v, when=%v",
			index+1,
			len(commits),
			commit.Hash.String(),
			commit.Author.When.UTC().String(),
		))

		results, err := CountLines(repository, commit, options.CountLinesOption)
		if err != nil {
			return nil, err
		}

		lineCounts := make([]GenerateResultCommitLineCount, 0)
		for _, result := range results {
			areaAuthors, err := AllocateAreas(areaInfo, result)
			if err != nil {
				return nil, err
			}

			areas := make([]GenerateResultCommitLineCountArea, 0)
			authors := make([]GenerateResultCommitLineCountAuthor, 0)
			for _, areaAuthor := range areaAuthors {
				areas = append(areas, GenerateResultCommitLineCountArea{
					Name:        areaAuthor.Area.Name,
					Size:        areaAuthor.Area.Size,
					Ratio:       areaAuthor.AreaRatio,
					AuthorEmail: areaAuthor.Author,
					AuthorRank:  areaAuthor.AuthorRank,
				})

				email := areaAuthor.Author

				found := false
				for _, author := range authors {
					if author.Email == email {
						found = true
						break
					}
				}
				if !found {
					if gitHubLoginNameCache[email] == nil {
						login, err := FindLoginByEmail(email)
						if err != nil {
							return nil, err
						}
						gitHubLoginNameCache[email] = &login
					}

					gitHubLogin := gitHubLoginNameCache[email]

					authors = append(authors, GenerateResultCommitLineCountAuthor{
						Email:       email,
						Name:        result.NameByAuthor[email],
						GitHubLogin: *gitHubLogin,
						LineCount:   result.LinesByAuthor[email],
						Rank:        areaAuthor.AuthorRank,
					})
				}
			}

			notAllocatedAuthors := make([]GenerateResultCommitLineCountAuthor, 0)
			for email, lineCount := range result.LinesByAuthor {
				found := false
				for _, areaAuthor := range areaAuthors {
					if areaAuthor.Author == email {
						found = true
						break
					}
				}

				if !found {
					if gitHubLoginNameCache[email] == nil {
						login, err := FindLoginByEmail(email)
						if err != nil {
							return nil, err
						}
						gitHubLoginNameCache[email] = &login
					}

					gitHubLogin := gitHubLoginNameCache[email]

					notAllocatedAuthors = append(notAllocatedAuthors, GenerateResultCommitLineCountAuthor{
						Email:       email,
						Name:        result.NameByAuthor[email],
						GitHubLogin: *gitHubLogin,
						LineCount:   lineCount,
					})
				}
			}

			sort.SliceStable(notAllocatedAuthors, func(i, j int) bool {
				if notAllocatedAuthors[i].LineCount == notAllocatedAuthors[j].LineCount {
					return notAllocatedAuthors[i].Email < notAllocatedAuthors[j].Email
				} else {
					return notAllocatedAuthors[i].LineCount > notAllocatedAuthors[j].LineCount
				}
			})

			for i, notAllocatedAuthor := range notAllocatedAuthors {
				notAllocatedAuthor.Rank = len(authors) + i + 1
				notAllocatedAuthors[i] = notAllocatedAuthor
			}

			lineCounts = append(lineCounts, GenerateResultCommitLineCount{
				FilterRegex: result.Filter.String(),
				FileCount:   len(result.MatchedFiles),
				Areas:       areas,
				Authors:     append(authors, notAllocatedAuthors...),
			})
		}

		resultCommits = append(resultCommits, GenerateResultCommit{
			Hash:        commit.Hash.String(),
			CommittedAt: commit.Author.When.UTC(),
			LineCounts:  lineCounts,
		})
	}

	return &GenerateResult{
		Repository:  GetRemoteUrl(repositoryRemoteLocation),
		Source:      GetSource(repositoryRemoteLocation),
		GeneratedAt: time.Now().UTC(),
		Commits:     resultCommits,
	}, nil
}

var gitHubRegex = regexp.MustCompile("^https://github\\.com/")
var gitHubSshRegex = regexp.MustCompile("^git@github\\.com:")
var gitSuffixRegex = regexp.MustCompile("\\.git$")

func GetSource(value string) string {
	if gitHubRegex.MatchString(value) || gitHubSshRegex.MatchString(value) {
		return "github"
	}
	return "unknown"
}

func GetRemoteUrl(value string) string {
	if gitHubRegex.MatchString(value) {
		return gitSuffixRegex.ReplaceAllString(value, "")
	} else if gitHubSshRegex.MatchString(value) {
		return gitSuffixRegex.ReplaceAllString("https://github.com/"+gitHubSshRegex.ReplaceAllString(value, ""), "")
	}
	return value
}
