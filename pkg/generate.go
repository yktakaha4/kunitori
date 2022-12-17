package pkg

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"log"
	"os"
	"regexp"
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
	FileNames   []string                              `json:"fileNames"`
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

func Generate(options *GenerateOptions) (*GenerateResult, error) {
	var repository *git.Repository
	var repositoryLocation string

	ShowSlowMessage()

	areaInfo, err := GetAreaInfo(options.Region)
	if err != nil {
		return nil, err
	}

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
		fmt.Println(fmt.Sprintf("open repository: url=%v", options.RepositoryUrl))

		repository, err = CloneRepository(options.RepositoryUrl, tempDir)
		if err != nil {
			return nil, err
		}
	} else if options.RepositoryPath != "" {
		repositoryLocation = options.RepositoryPath
		fmt.Println(fmt.Sprintf("open repository: path=%v", options.RepositoryPath))

		repository, err = OpenRepository(options.RepositoryPath)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("should specify url or path")
	}

	fmt.Println("search commit...")

	commits, err := SearchCommits(repository, options.SearchCommitsOptions)
	if err != nil {
		return nil, err
	}

	fmt.Println(fmt.Sprintf("matched commits: count=%v", len(commits)))

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
					})
				}
			}

			lineCounts = append(lineCounts, GenerateResultCommitLineCount{
				FilterRegex: result.Filter.String(),
				FileNames:   result.MatchedFiles,
				Areas:       areas,
				Authors:     authors,
			})
		}

		resultCommits = append(resultCommits, GenerateResultCommit{
			Hash:        commit.Hash.String(),
			CommittedAt: commit.Author.When.UTC(),
			LineCounts:  lineCounts,
		})
	}

	return &GenerateResult{
		Repository:  repositoryLocation,
		Source:      GetSource(repositoryLocation),
		GeneratedAt: time.Now().UTC(),
		Commits:     resultCommits,
	}, nil
}

func GetSource(value string) string {
	gitHubRegex := regexp.MustCompile("^https://github\\.com/")
	if gitHubRegex.MatchString(value) {
		return "github"
	}
	return "unknown"
}
