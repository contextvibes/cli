// internal/thea/client.go
package thea

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Manifest represents the structure of the thea-manifest.json file.
type Manifest struct {
	ManifestSchemaVersion       string     `json:"manifestSchemaVersion"`
	THEAFrameworkReleaseVersion string     `json:"theaFrameworkReleaseVersion"`
	LastUpdated                 string     `json:"lastUpdated"` // ISO 8601 timestamp
	Artifacts                   []Artifact `json:"artifacts"`
}

// Artifact represents a single artifact entry in the manifest.
type Artifact struct {
	ID                string   `json:"id"`
	FileExtension     string   `json:"fileExtension"`
	Title             string   `json:"title"`
	ArtifactVersion   string   `json:"artifactVersion"`
	Summary           string   `json:"summary"`
	UsageGuidance     []string `json:"usageGuidance,omitempty"`
	Owner             string   `json:"owner,omitempty"`
	CreatedDate       string   `json:"createdDate,omitempty"`
	LastModifiedDate  string   `json:"lastModifiedDate,omitempty"`
	DefaultTargetPath string   `json:"defaultTargetPath,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

// Client provides methods to interact with the THEA framework (e.g., fetching manifests and artifacts).
type Client struct {
	logger     *slog.Logger
	config     *THEAServiceConfig // A dedicated config substruct for this client
	httpClient *http.Client       // For making HTTP requests
}

// THEAServiceConfig contains configuration specific to the THEA client.
// This would be part of your main AppConfig.LoadedAppConfig.THEA.ServiceConfig or similar.
type THEAServiceConfig struct {
	ManifestURL        string        // Full URL to the thea-manifest.json
	DefaultArtifactRef string        // e.g., "main" or a specific release tag like "v0.7.0"
	RawContentBaseURL  string        // e.g., "https://raw.githubusercontent.com/contextvibes/THEA" (without ref)
	CacheDir           string        // Directory for caching manifests/artifacts (e.g., ~/.contextvibes/cache/thea)
	CacheTTL           time.Duration // Time-to-live for cached items
	RequestTimeout     time.Duration // Timeout for HTTP requests
}

// NewClient creates a new THEA client.
// The context here is primarily for future use if any initial setup needs it;
// individual methods like FetchArtifactContent will also take a context.
func NewClient(_ context.Context, cfg *THEAServiceConfig, logger *slog.Logger) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("THEA service config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if cfg.ManifestURL == "" {
		return nil, fmt.Errorf("THEA manifest URL is not configured")
	}
	if cfg.RawContentBaseURL == "" {
		return nil, fmt.Errorf("THEA raw content base URL is not configured")
	}
	if cfg.DefaultArtifactRef == "" {
		// Could default to "main" if not set, or error out
		logger.Warn("THEA default artifact ref is not configured, consider setting it. Defaulting to 'main'.")
		cfg.DefaultArtifactRef = "main" // Or handle as error
	}

	// Default timeout for HTTP client if not specified in config
	timeout := cfg.RequestTimeout
	if timeout == 0 {
		timeout = 30 * time.Second // Default to 30 seconds
	}

	return &Client{
		logger: logger.With(slog.String("service", "thea")), // Add service context to logger
		config: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// --- Manifest Methods ---

// fetchManifest fetches the manifest from the configured URL.
func (c *Client) fetchManifest(ctx context.Context) (*Manifest, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.ManifestURL, nil)
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to create manifest request", slog.String("url", c.config.ManifestURL), slog.String("error", err.Error()))
		return nil, fmt.Errorf("creating manifest request: %w", err)
	}

	c.logger.InfoContext(ctx, "Fetching THEA manifest", slog.String("url", c.config.ManifestURL))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to fetch manifest", slog.String("url", c.config.ManifestURL), slog.String("error", err.Error()))
		return nil, fmt.Errorf("fetching manifest from %s: %w", c.config.ManifestURL, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.WarnContext(ctx, "Failed to close manifest response body", slog.String("error", closeErr.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		c.logger.ErrorContext(ctx, "Failed to fetch manifest, unexpected status", slog.String("url", c.config.ManifestURL), slog.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("fetching manifest: received status %d from %s", resp.StatusCode, c.config.ManifestURL)
	}

	var manifest Manifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		c.logger.ErrorContext(ctx, "Failed to decode manifest JSON", slog.String("url", c.config.ManifestURL), slog.String("error", err.Error()))
		return nil, fmt.Errorf("decoding manifest JSON from %s: %w", c.config.ManifestURL, err)
	}
	// TODO: Add manifest caching logic here (save to c.config.CacheDir)
	return &manifest, nil
}

// LoadManifest retrieves the THEA manifest, using a cache if available and valid.
// For MVP, it might just always fetch.
func (c *Client) LoadManifest(ctx context.Context) (*Manifest, error) {
	// TODO: Implement caching logic:
	// 1. Check if manifest exists in c.config.CacheDir.
	// 2. If yes, check if it's within c.config.CacheTTL.
	// 3. If yes and valid, load and return from cache.
	// 4. Otherwise, fetch, save to cache, and return.

	// For MVP - always fetch:
	return c.fetchManifest(ctx)
}

// GetArtifactMetadata finds an artifact in the loaded manifest by ID.
// It does not consider version yet for simplicity in this example, but should.
func (m *Manifest) GetArtifactByID(id string) (*Artifact, error) {
	// Corrected loop: iterate over the slice m.Artifacts
	for i := range m.Artifacts { // Iterate by index to get a pointer to the original element
		if m.Artifacts[i].ID == id {
			return &m.Artifacts[i], nil // Return a pointer to the artifact in the slice
		}
	}
	return nil, fmt.Errorf("artifact with ID '%s' not found in manifest", id)
}

// --- Artifact Content Fetching ---

// FetchArtifactContentByID fetches the content of a specific artifact version.
// If version is empty, it might fetch the one specified in the manifest or use DefaultArtifactRef.
func (c *Client) FetchArtifactContentByID(ctx context.Context, id string, artifactVersionHint string) (string, error) {
	manifest, err := c.LoadManifest(ctx) // Ensure manifest is loaded (potentially from cache)
	if err != nil {
		return "", fmt.Errorf("loading manifest: %w", err)
	}

	artifact, err := manifest.GetArtifactByID(id)
	if err != nil {
		return "", err // Artifact ID not found
	}

	// Determine the Git ref to use for fetching this artifact's content
	// For now, let's assume the artifactVersionHint IS the Git ref (tag/branch)
	// or we use the config's DefaultArtifactRef.
	// A more complex strategy would involve resolving artifact.ArtifactVersion to a Git ref.
	gitRef := c.config.DefaultArtifactRef
	if artifactVersionHint != "" {
		gitRef = artifactVersionHint
	} else if artifact.ArtifactVersion != "" {
		// This assumes artifact.ArtifactVersion can be directly used as a Git ref (e.g., "v1.2.3")
		// This part needs careful thought: does artifactVersion in manifest mean "this is the version at DefaultArtifactRef"
		// or is artifactVersion itself a tag? For now, let's assume direct use or fallback.
		c.logger.DebugContext(ctx, "Using artifactVersion from manifest as Git ref hint", slog.String("id", id), slog.String("artifact_version_ref", artifact.ArtifactVersion))
		// Potentially, we'd have a mapping or convention: artifact v1.2.3 is found on git tag thea-artifact-<id>-v1.2.3 or release tag vX.Y.Z
		// For simplicity, if artifactVersionHint is not given, we'll try artifact.ArtifactVersion IF it looks like a version tag.
		// This logic will need refinement based on THEA repo's tagging strategy.
		// For now: if artifactVersionHint is empty, use default ref.
	}

	// Construct the actual source path in the repo
	var effectiveSourcePathInRepo string
	if artifact.FileExtension != "" {
		effectiveSourcePathInRepo = artifact.ID + "." + artifact.FileExtension
	} else {
		effectiveSourcePathInRepo = artifact.ID // For files like .editorconfig where ID is full name
	}
	effectiveSourcePathInRepo = strings.TrimPrefix(effectiveSourcePathInRepo, "/") // Ensure no leading slash for JoinPath

	// Construct the full raw download URL
	// Example: https://raw.githubusercontent.com/contextvibes/THEA/main/docs/templates/contributing-guide.md
	fullURL, err := url.JoinPath(c.config.RawContentBaseURL, gitRef, effectiveSourcePathInRepo)
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to construct artifact download URL",
			slog.String("base", c.config.RawContentBaseURL),
			slog.String("ref", gitRef),
			slog.String("path_in_repo", effectiveSourcePathInRepo),
			slog.String("error", err.Error()))
		return "", fmt.Errorf("constructing artifact download URL: %w", err)
	}

	c.logger.InfoContext(ctx, "Fetching THEA artifact content",
		slog.String("id", id),
		slog.String("version_ref_used", gitRef),
		slog.String("url", fullURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating artifact content request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching artifact content from %s: %w", fullURL, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.WarnContext(ctx, "Failed to close artifact content response body", slog.String("url", fullURL), slog.String("error", closeErr.Error()))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024)) // Read a bit of the body for error context
		c.logger.ErrorContext(ctx, "Failed to fetch artifact content, unexpected status",
			slog.String("url", fullURL),
			slog.Int("status", resp.StatusCode),
			slog.String("response_snippet", string(bodyBytes)))
		return "", fmt.Errorf("fetching artifact content: received status %d from %s", resp.StatusCode, fullURL)
	}

	contentBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading artifact content from %s: %w", fullURL, err)
	}

	return string(contentBytes), nil
}
