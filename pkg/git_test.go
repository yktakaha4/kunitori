package pkg

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
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
