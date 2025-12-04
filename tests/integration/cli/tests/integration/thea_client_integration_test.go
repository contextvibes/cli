// Package integration_test contains integration tests for the CLI.
package integration_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/contextvibes/cli/internal/thea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newIntegrationTestLogger() *slog.Logger {
	//nolint:exhaustruct // Default handler options are sufficient.
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// Build Tag: //go:build integration.
func TestTHEAClient_Integration_FetchRealManifest(t *testing.T) {
	t.Parallel()

	if os.Getenv("RUN_INTEGRATION_TESTS") == "" && !testing.Short() {
		t.Skip("Skipping integration test: THEAClient_Integration_FetchRealManifest...")
	}

	//nolint:exhaustruct // Partial config is sufficient for test.
	cfg := thea.ServiceConfig{
		ManifestURL:       "https://raw.githubusercontent.com/contextvibes/THEA/main/thea-manifest.json", // LIVE URL
		RawContentBaseURL: "https://raw.githubusercontent.com/contextvibes/THEA",                         // LIVE URL base

		DefaultArtifactRef: "main", // Using main branch
		RequestTimeout:     60 * time.Second,
	}
	logger := newIntegrationTestLogger()
	client, err := thea.NewClient(context.Background(), &cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	logger.Info(
		"INTEGRATION TEST: Attempting to load manifest from live THEA repository (main branch)...",
	)

	manifest, err := client.LoadManifest(ctx)

	require.NoError(t, err)
	require.NotNil(t, manifest)
	assert.NotEmpty(t, manifest.ManifestSchemaVersion)
	assert.NotEmpty(t, manifest.THEAFrameworkReleaseVersion)

	// Check specifically for our kickoff prompt artifact
	foundKickoffPrompt := false

	var kickoffArtifact thea.Artifact // Corrected type

	for _, art := range manifest.Artifacts {
		if art.ID == "playbooks/project_initiation/master_strategic_kickoff_prompt" {
			foundKickoffPrompt = true
			kickoffArtifact = art // Corrected assignment

			break
		}
	}

	assert.True(
		t,
		foundKickoffPrompt,
		"Manifest should contain the 'playbooks/project_initiation/master_strategic_kickoff_prompt' artifact",
	)

	if foundKickoffPrompt {
		t.Logf(
			"Found kickoff prompt artifact: Title: '%s', Version: '%s'",
			kickoffArtifact.Title,
			kickoffArtifact.ArtifactVersion,
		)
		assert.Equal(t, "md", kickoffArtifact.FileExtension)
	}
}

// Build Tag: //go:build integration.
func TestTHEAClient_Integration_FetchRealArtifactContent(t *testing.T) {
	t.Parallel()

	if os.Getenv("RUN_INTEGRATION_TESTS") == "" && !testing.Short() {
		t.Skip("Skipping integration test: TestTHEAClient_Integration_FetchRealArtifactContent...")
	}

	testArtifactID := "playbooks/project_initiation/master_strategic_kickoff_prompt" // REAL ID
	testArtifactRef := "main"                                                        // Fetch from main branch

	expectedContentSubstring := "AI Facilitator Instructions & Persona" // REAL SUBSTRING from your kickoff prompt

	//nolint:exhaustruct // Partial config is sufficient for test.
	cfg := thea.ServiceConfig{
		ManifestURL:        "https://raw.githubusercontent.com/contextvibes/THEA/main/thea-manifest.json",
		RawContentBaseURL:  "https://raw.githubusercontent.com/contextvibes/THEA",
		DefaultArtifactRef: "main",
		RequestTimeout:     60 * time.Second,
	}
	logger := newIntegrationTestLogger()
	client, err := thea.NewClient(context.Background(), &cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	logger.Info(
		"INTEGRATION TEST: Attempting to fetch content for artifact ID from live THEA repository...",
		slog.String("id", testArtifactID),
		slog.String("ref_hint", testArtifactRef),
	)

	content, err := client.FetchArtifactContentByID(ctx, testArtifactID, testArtifactRef)

	require.NoError(t, err)
	require.NotEmpty(t, content)
	assert.Contains(t, content, expectedContentSubstring)

	t.Logf(
		"Successfully fetched content for artifact '%s'. Length: %d chars.",
		testArtifactID,
		len(content),
	)
}
