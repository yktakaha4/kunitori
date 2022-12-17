package pkg

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"regexp"
	"sort"
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

func SearchCommits(repository *git.Repository, options *SearchCommitsOptions) ([]*object.Commit, error) {
	maxLimit := 50

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

	limit := maxLimit
	if options.Limit > 0 && options.Limit < maxLimit {
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

type CountLinesOption struct {
	Filters       []regexp.Regexp
	AuthorRegexes map[string]regexp.Regexp
}

type CountLinesResult struct {
	Filter        regexp.Regexp
	LinesByAuthor map[string]int
}

func CountLines(commit *object.Commit, options *CountLinesOption) ([]*CountLinesResult, error) {
	log.Printf("start CountLines: commit=%+v, options=%+v", commit.Hash, options)
	results := make([]*CountLinesResult, 0)
	for _, filter := range options.Filters {
		results = append(results, &CountLinesResult{
			Filter: filter,
		})
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	fileCount, targetCount, linesCount := 0, 0, 0
	err = tree.Files().ForEach(func(file *object.File) error {
		fileCount++

		if file.Type() != plumbing.BlobObject {
			return nil
		}

		targetCount++

		for _, result := range results {
			if !result.Filter.MatchString(file.Name) {
				return nil
			}

			log.Printf("match: file=%+v, filter=%+v", file.Name, result.Filter)

			blameResult, err := git.Blame(commit, file.Name)
			if err != nil {
				return err
			}

			for _, line := range blameResult.Lines {
				author := line.Author
				for aut, autRegex := range options.AuthorRegexes {
					if autRegex.MatchString(line.Author) {
						author = aut
						break
					}
				}
				result.LinesByAuthor[author] += 1
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
