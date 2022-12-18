package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/dlclark/regexp2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"
)

func CloneRepository(url string, path string) (*git.Repository, error) {
	log.Printf("start CloneRepository: url=%+v, path=%+v", url, path)
	repository, err := git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("clone completed: repository=%+v", repository)

	return repository, nil
}

func OpenRepository(path string) (*git.Repository, error) {
	log.Printf("start OpenRepository: path=%+v", path)

	repository, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

type SearchCommitsOptions struct {
	Since    time.Time
	Until    time.Time
	Interval time.Duration
	Limit    int
}

const SearchCommitMaxLimit = 15

func SearchCommits(repository *git.Repository, options *SearchCommitsOptions) ([]*object.Commit, error) {
	log.Printf("start SearchCommits: repository=%+v, options=%+v", repository, options)

	reference, err := repository.Head()
	if err != nil {
		return nil, err
	}

	hash := reference.Hash()

	since := time.UnixMilli(0)
	if !options.Since.IsZero() {
		since = options.Since
	}
	since = since.UTC()

	until := time.Now()
	if !options.Until.IsZero() {
		until = options.Until
	}
	until = until.UTC()

	log.Printf("filter commits: hash=%v, since=%+v, until=%+v", hash, since, until)

	commitIter, err := repository.Log(&git.LogOptions{
		From: hash,
	})
	if err != nil {
		return nil, err
	}

	filteredCommits := make([]*object.Commit, 0)
	commitCount, pickCount, skipCount := 0, 0, 0
	err = commitIter.ForEach(func(commit *object.Commit) error {
		commitCount++

		commitWhen := commit.Author.When.UTC()
		if commitWhen.After(since) && commitWhen.Before(until) {
			pickCount++
			filteredCommits = append(filteredCommits, commit)
		} else {
			skipCount++
		}

		return nil
	})

	log.Printf(
		"filter complete: commitCount=%+v, pickCount=%+v, skipCount=%+v",
		commitCount, pickCount, skipCount,
	)

	if len(filteredCommits) == 0 {
		return filteredCommits, nil
	}

	log.Printf("sort commits: count=%+v", len(filteredCommits))

	sort.SliceStable(filteredCommits, func(i, j int) bool {
		return filteredCommits[i].Author.When.UTC().After(filteredCommits[j].Author.When.UTC())
	})

	log.Printf(
		"sort complete: latest=%+v, least=%+v",
		filteredCommits[0].Author.When.UTC(),
		filteredCommits[len(filteredCommits)-1].Author.When.UTC(),
	)

	interval := options.Interval
	if interval < 0 {
		interval = 0
	}

	limit := SearchCommitMaxLimit
	if options.Limit > 0 && options.Limit < SearchCommitMaxLimit {
		limit = options.Limit
	}

	log.Printf("thin commits: interval=%+v, limit=%+v", interval, limit)

	commits := make([]*object.Commit, 0)
	pickCount, skipCount = 0, 0
	for _, commit := range filteredCommits {
		hash := commit.Hash.String()
		commitWhen := commit.Author.When.UTC()
		if len(commits) > 0 {
			recentWhen := commits[len(commits)-1].Author.When.UTC()

			whenDiff := recentWhen.Sub(commitWhen)
			if whenDiff < interval {
				skipCount++
				continue
			}
		}

		log.Printf("hash=%+v, commitWhen=%+v", hash, commitWhen)
		commits = append(commits, commit)
		pickCount++

		if pickCount >= limit {
			break
		}
	}

	log.Printf(
		"thin completed: commitCount=%+v, pickCount=%+v, skipCount=%+v",
		len(commits), pickCount, skipCount,
	)

	return commits, nil
}

type AuthorRegex struct {
	Condition *regexp2.Regexp
	Author    string
}

type CountLinesOption struct {
	Filters       []*regexp2.Regexp
	AuthorRegexes []AuthorRegex
}

type CountLinesResult struct {
	Filter        *regexp2.Regexp
	LinesByAuthor map[string]int
	NameByAuthor  map[string]string
	MatchedFiles  []string
}

func CountLines(repository *git.Repository, commit *object.Commit, options *CountLinesOption) ([]*CountLinesResult, error) {
	log.Printf("start CountLines: commit=%+v, options=%+v", commit.Hash, options)
	results := make([]*CountLinesResult, 0)
	for _, filter := range options.Filters {
		results = append(results, &CountLinesResult{
			Filter:        filter,
			LinesByAuthor: map[string]int{},
			NameByAuthor:  map[string]string{},
			MatchedFiles:  make([]string, 0),
		})
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	fileCount, targetCount, linesCount, errorCount := 0, 0, 0, 0
	err = tree.Files().ForEach(func(file *object.File) error {
		fileCount++

		if file.Type() != plumbing.BlobObject {
			return nil
		}

		for _, result := range results {
			isMatch, err := result.Filter.MatchString(file.Name)
			if err != nil {
				return err
			}
			if !isMatch {
				continue
			}

			result.MatchedFiles = append(result.MatchedFiles, file.Name)
			targetCount++

			log.Printf("match: file=%+v, filter=%+v", file.Name, result.Filter.String())

			lines := make([]*git.Line, 0)
			if IsUseGitCommandProvided() {
				lines, err = BlameWithGitCommand(repository, commit, file.Name)
				if err != nil {
					return err
				}
			} else {
				blameResult, err := git.Blame(commit, file.Name)
				if err != nil {
					log.Printf("failed to blame: err=%v", err)
					errorCount++
					continue
				}
				lines = blameResult.Lines
			}

			for _, line := range lines {
				author := line.Author
				for _, autRegex := range options.AuthorRegexes {
					isMatch, err := autRegex.Condition.MatchString(line.Author)
					if err != nil {
						return err
					}
					if isMatch {
						author = autRegex.Author
						break
					}
				}
				result.LinesByAuthor[author] += 1

				if result.NameByAuthor[author] == "" {
					lineCommit, err := repository.CommitObject(line.Hash)
					if err != nil {
						log.Printf("failed to get line commit: err=%v", err)
						continue
					}
					result.NameByAuthor[author] = lineCommit.Author.Name
				}

				linesCount++
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	log.Printf(
		"traverse complete: fileCount=%+v, targetCount=%v, linesCount=%+v",
		fileCount, targetCount, linesCount,
	)

	return results, nil
}

const KunitoriUseGitCommandProvidedKey = "KUNITORI_USE_GIT_COMMAND"

func IsUseGitCommandProvided() bool {
	return len(os.Getenv(KunitoriUseGitCommandProvidedKey)) > 0
}

var blameLineRegexp = regexp.MustCompile("^\\w+\\s\\d+\\)\\s")
var invalidCharacterRegexp = regexp.MustCompile("\\W")

func BlameWithGitCommand(repository *git.Repository, commit *object.Commit, file string) ([]*git.Line, error) {
	workTree, err := repository.Worktree()
	if err != nil {
		return nil, err
	}

	repoRoot := workTree.Filesystem.Root()
	hash := commit.Hash.String()
	blameResult, err := exec.Command("git", "-C", repoRoot, "blame", "-sl", hash, file).Output()
	if err != nil {
		return nil, err
	}

	badCommit := map[string]bool{}
	lineCommitCache := map[string]*object.Commit{}

	lines := make([]*git.Line, 0)
	scanner := bufio.NewScanner(bytes.NewReader(blameResult))
	for scanner.Scan() {
		line := scanner.Text()
		hashStr := invalidCharacterRegexp.ReplaceAllString(strings.Split(line, " ")[0], "")

		if badCommit[hashStr] {
			continue
		} else if lineCommitCache[hashStr] == nil {
			lineHash := plumbing.NewHash(hashStr)
			lineCommit, err := repository.CommitObject(lineHash)
			if err != nil {
				log.Printf(fmt.Sprintf("invalid commit: hash=%v, error=%v", hashStr, err))
				badCommit[hashStr] = true
				continue
			}

			lineCommitCache[hashStr] = lineCommit
		}

		lineCommit := lineCommitCache[hashStr]

		text := blameLineRegexp.ReplaceAllString(line, "")

		lines = append(lines, &git.Line{
			Author: lineCommit.Author.Email,
			Text:   text,
			Date:   lineCommit.Author.When.UTC(),
			Hash:   lineCommit.Hash,
		})
	}

	return lines, nil
}
