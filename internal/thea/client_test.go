// Package thea_test contains tests for the thea package.
package thea_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/contextvibes/cli/internal/thea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a test logger.
func newTestLogger() *slog.Logger {
	// For CI or quiet tests, use io.Discard. For local debugging, os.Stdout is fine.
	// return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	//nolint:exhaustruct // Default handler options are sufficient.
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestFetchManifest_Success(t *testing.T) {
	t.Parallel()
	// These types are defined in client.go, in the same 'thea' package
	//nolint:exhaustruct // Partial initialization is sufficient for test.
	expectedManifest := thea.Manifest{
		ManifestSchemaVersion:       "1.3.0",
		THEAFrameworkReleaseVersion: "v0.7.0",
		//nolint:exhaustruct // Partial initialization is sufficient for test.
		Artifacts: []thea.Artifact{
			{ID: "test-id", Title: "Test Artifact"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/thea-manifest.json", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		//nolint:musttag // Struct tags are present on Manifest struct.
		err := json.NewEncoder(w).Encode(expectedManifest)
		//nolint:testifylint // require is acceptable in handler for test setup failure.
		require.NoError(t, err)
	}))
	defer server.Close()

	// ServiceConfig is defined in client.go, in the same 'thea' package
	//nolint:exhaustruct // Partial config is sufficient for test.
	cfg := thea.ServiceConfig{
		ManifestURL:        server.URL + "/thea-manifest.json",
		RawContentBaseURL:  "http://dummy-raw-base.com", // Provide a value
		DefaultArtifactRef: "main",                      // Provide a value
		// Other fields can be zero if NewClient handles defaults or they aren't relevant here
	}
	logger := newTestLogger()
	// NewClient is defined in client.go, in the same 'thea' package
	client, err := thea.NewClient(context.Background(), &cfg, logger)
	require.NoError(t, err)
	require.NotNil(t, client)

	manifest, err := client.LoadManifest(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, manifest)
	assert.Equal(t, expectedManifest.ManifestSchemaVersion, manifest.ManifestSchemaVersion)
	assert.Len(t, manifest.Artifacts, 1)

	if len(manifest.Artifacts) > 0 { // Guard against panic if artifacts slice is unexpectedly empty
		assert.Equal(t, "test-id", manifest.Artifacts[0].ID)
	}
}

func TestFetchManifest_ServerReturns404(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// ServiceConfig is local to the 'thea' package
	//nolint:exhaustruct // Partial config is sufficient for test.
	cfg := thea.ServiceConfig{
		ManifestURL:        server.URL + "/thea-manifest.json",
		RawContentBaseURL:  "http://dummy-raw-base.com",
		DefaultArtifactRef: "main",
	}
	logger := newTestLogger()
	// NewClient is local to the 'thea' package
	client, err := thea.NewClient(context.Background(), &cfg, logger)
	require.NoError(
		t,
		err,
	) // NewClient itself should not error with this config
	require.NotNil(t, client)

	_, err = client.LoadManifest(context.Background())
	require.Error(t, err)

	if err != nil { // Guard for err being non-nil before calling Contains
		assert.Contains(t, err.Error(), "received status 404")
	}
}
