package internal

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestCloneRepository(t *testing.T) {
	urls := []string{
		"https://github.com/yktakaha4/eduterm.git",
		"git@github.com:yktakaha4/eduterm.git",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "TestCloneRepository")
			assert.NoError(t, err)
			defer func(path string) {
				err := os.RemoveAll(path)
				assert.NoError(t, err)
			}(tempDir)

			repository, err := CloneRepository(url, tempDir)
			assert.NoError(t, err)
			assert.NotNil(t, repository)
		})
	}
}
