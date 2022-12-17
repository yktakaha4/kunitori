package pkg

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFindLoginByEmail(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("user found with public email", func(t *testing.T) {
		login, err := FindLoginByEmail("audreyt@audreyt.org")
		assert.NoError(t, err)
		assert.Equal(t, "audreyt", login)
	})

	t.Run("user found with github email", func(t *testing.T) {
		login, err := FindLoginByEmail("20282867+yktakaha4@users.noreply.github.com")
		assert.NoError(t, err)
		assert.Equal(t, "yktakaha4", login)
	})

	t.Run("user not found", func(t *testing.T) {
		login, err := FindLoginByEmail("yktakaha4@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "", login)
	})

}

func TestIsGitHubAccessTokenProvided(t *testing.T) {
	t.Run("is set", func(t *testing.T) {
		t.Setenv(GitHubAccessTokenKey, "dummy-github-token")
		assert.True(t, IsGitHubAccessTokenProvided())
	})

	t.Run("is not set", func(t *testing.T) {
		t.Setenv(GitHubAccessTokenKey, "")
		assert.False(t, IsGitHubAccessTokenProvided())
	})
}
