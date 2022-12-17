package pkg

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"log"
	"regexp"
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
	Until    time.Time
	Interval time.Duration
}

func SearchCommits(repository *git.Repository, options *SearchCommitsOptions) ([]*object.Commit, error) {
	log.Printf("start SearchCommits: repository=%+v, options=%+v", repository, options)
	reference, err := repository.Head()
	if err != nil {
		return nil, err
	}

	hash := reference.Hash()

	log.Printf("head: hash=%+v", hash)

	commitIter, err := repository.Log(&git.LogOptions{
		From:  hash,
		Order: git.LogOrderCommitterTime,
		Until: &options.Until,
	})
	if err != nil {
		return nil, err
	}

	log.Printf("search commits...")

	commits := make([]*object.Commit, 0)
	commitCount, targetCount, skipCount := 0, 0, 0
	err = commitIter.ForEach(func(commit *object.Commit) error {
		commitCount++

		commitWhen := commit.Author.When
		if len(commits) > 0 {
			recentWhen := commits[len(commits)-1].Author.When

			whenDiff := recentWhen.Sub(commitWhen)
			if whenDiff < options.Interval {
				skipCount++
				return nil
			}
		}

		log.Printf("commit=%+v, when=%+v", commit, commitWhen)
		commits = append(commits, commit)
		targetCount++

		return nil
	})
	if err != nil {
		return nil, err
	}

	log.Printf(
		"search completed: commitCount=%+v, targetCount=%+v, skipCount=%+v",
		commitCount, targetCount, skipCount,
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
	log.Printf("start CountLines: commit=%+v, options=%+v", commit, options)
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

	log.Printf("traverse: tree=%+v", tree)

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
