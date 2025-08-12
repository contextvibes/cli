// In internal/thea/client_test.go
package thea // Ensures this file is part of the 'thea' package

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create a test logger.
func newTestLogger() *slog.Logger {
	// For CI or quiet tests, use io.Discard. For local debugging, os.Stdout is fine.
	// return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func TestFetchManifest_Success(t *testing.T) {
	// These types are defined in client.go, in the same 'thea' package
	expectedManifest := Manifest{ // No package qualifier
		ManifestSchemaVersion:       "1.3.0",
		THEAFrameworkReleaseVersion: "v0.7.0",
		Artifacts: []Artifact{ // No package qualifier
			{ID: "test-id", Title: "Test Artifact"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/thea-manifest.json", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(expectedManifest)
		require.NoError(t, err)
	}))
	defer server.Close()

	// THEAServiceConfig is defined in client.go, in the same 'thea' package
	cfg := THEAServiceConfig{ // No package qualifier
		ManifestURL:        server.URL + "/thea-manifest.json",
		RawContentBaseURL:  "http://dummy-raw-base.com", // Provide a value
		DefaultArtifactRef: "main",                      // Provide a value
		// Other fields can be zero if NewClient handles defaults or they aren't relevant here
	}
	logger := newTestLogger()
	// NewClient is defined in client.go, in the same 'thea' package
	client, err := NewClient(context.Background(), &cfg, logger) // No package qualifier
	require.NoError(t, err)
	require.NotNil(t, client)

	manifest, err := client.fetchManifest(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, manifest)
	assert.Equal(t, expectedManifest.ManifestSchemaVersion, manifest.ManifestSchemaVersion)
	assert.Len(t, manifest.Artifacts, 1)

	if len(manifest.Artifacts) > 0 { // Guard against panic if artifacts slice is unexpectedly empty
		assert.Equal(t, "test-id", manifest.Artifacts[0].ID)
	}
}

func TestFetchManifest_ServerReturns404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// THEAServiceConfig is local to the 'thea' package
	cfg := THEAServiceConfig{ // No package qualifier
		ManifestURL:        server.URL + "/thea-manifest.json",
		RawContentBaseURL:  "http://dummy-raw-base.com",
		DefaultArtifactRef: "main",
	}
	logger := newTestLogger()
	// NewClient is local to the 'thea' package
	client, err := NewClient(context.Background(), &cfg, logger) // No package qualifier
	require.NoError(
		t,
		err,
	) // NewClient itself should not error with this config
	require.NotNil(t, client)

	_, err = client.fetchManifest(context.Background())
	assert.Error(t, err)

	if err != nil { // Guard for err being non-nil before calling Contains
		assert.Contains(t, err.Error(), "received status 404")
	}
}

// ... (The TestFetchArtifactContentByID_Success and TestNewClient_Validation
//      would also need similar corrections: remove package qualifiers for
//      Manifest, Artifact, THEAServiceConfig, and NewClient) ...
