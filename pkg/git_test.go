package pkg

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestCloneRepository(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

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

func TestOpenRepository(t *testing.T) {
	paths := []string{
		testDataPath("django"),
	}

	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			repository, err := OpenRepository(p)
			assert.NoError(t, err)
			assert.NotNil(t, repository)
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
