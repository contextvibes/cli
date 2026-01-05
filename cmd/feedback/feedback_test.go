package feedback_test

import (
	"testing"

	"github.com/contextvibes/cli/cmd/feedback"
	"github.com/contextvibes/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFeedbackArgs(t *testing.T) {
	t.Parallel()

	repos := map[string]string{
		"cli":  "contextvibes/cli",
		"thea": "contextvibes/thea",
	}
	defaultRepo := "cli"

	tests := []struct {
		name          string
		args          []string
		expectedAlias string
		expectedTitle string
	}{
		{
			name:          "no args",
			args:          []string{},
			expectedAlias: "cli",
			expectedTitle: "",
		},
		{
			name:          "title only",
			args:          []string{"My bug report"},
			expectedAlias: "cli",
			expectedTitle: "My bug report",
		},
		{
			name:          "alias and title",
			args:          []string{"thea", "Docs error"},
			expectedAlias: "thea",
			expectedTitle: "Docs error",
		},
		{
			name:          "alias only (interactive title)",
			args:          []string{"thea"},
			expectedAlias: "thea",
			expectedTitle: "",
		},
		{
			name:          "unknown alias treated as title",
			args:          []string{"unknown", "stuff"},
			expectedAlias: "cli",
			expectedTitle: "unknown", // "unknown" is treated as the title
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the exported bridge function
			alias, title := feedback.ParseFeedbackArgs(tt.args, defaultRepo, repos)
			assert.Equal(t, tt.expectedAlias, alias)
			assert.Equal(t, tt.expectedTitle, title)
		})
	}
}

func TestResolveTarget(t *testing.T) {
	t.Parallel()

	cfg := &config.FeedbackSettings{
		DefaultRepository: "cli",
		Repositories: map[string]string{
			"cli": "owner/repo",
			"bad": "invalid-format",
		},
	}

	t.Run("valid resolution", func(t *testing.T) {
		t.Parallel()

		// Use the bridge constructor
		params := feedback.NewTestParams(cfg, "", "")
		
		owner, repo, err := feedback.ResolveTarget([]string{"cli", "title"}, params)
		require.NoError(t, err)
		assert.Equal(t, "owner", owner)
		assert.Equal(t, "repo", repo)
	})

	t.Run("invalid alias", func(t *testing.T) {
		t.Parallel()

		params := feedback.NewTestParams(cfg, "", "")
		owner, repo, err := feedback.ResolveTarget([]string{"missing"}, params)
		require.NoError(t, err)
		assert.Equal(t, "owner", owner)
		assert.Equal(t, "repo", repo)
	})

	t.Run("invalid config format", func(t *testing.T) {
		t.Parallel()

		params := feedback.NewTestParams(cfg, "", "")
		// Explicitly use "bad" alias
		_, _, err := feedback.ResolveTarget([]string{"bad", "title"}, params)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid repository format")
	})
}

func TestConstructWorkItem(t *testing.T) {
	t.Parallel()

	params := feedback.NewTestParams(nil, "Test Issue", "Test Body")
	user := "testuser"
	version := "v1.0.0-test"

	item := feedback.ConstructWorkItem(params, user, version)

	assert.Equal(t, "Test Issue", item.Title)
	assert.Contains(t, item.Body, "Test Body")
	assert.Contains(t, item.Body, "**Context**")
	assert.Contains(t, item.Body, "CLI Version:** `v1.0.0-test`")
	assert.Contains(t, item.Body, "Filed by:** @testuser")
	assert.Equal(t, "testuser", item.Author)
	assert.Contains(t, item.Labels, "feedback")
}
