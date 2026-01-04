package quality

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/contextvibes/cli/internal/globals"
	"github.com/contextvibes/cli/internal/ui"
	"github.com/spf13/cobra"
)

const (
	defaultPort    = 8080
	defaultAddress = "localhost"
)

//nolint:gochecknoglobals // Cobra flags require package-level variables.
var (
	port    int
	address string
)

// serveCmd represents the serve command
//
//nolint:exhaustruct,gochecknoglobals // Cobra commands are defined with partial structs and globals by design.
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Runs the quality pipeline as a server, exposing it via an HTTP API.",
	Long: `This command starts a lightweight HTTP server to provide access to the quality
pipeline via a REST-ful API. It is intended for editor integrations and other
programmatic uses.

The primary endpoint is /quality, which accepts a 'mode' query parameter
(e.g., /quality?mode=strict) and returns the results as a JSON object.`,
	RunE: runServe,
}

func runServe(cmd *cobra.Command, args []string) error {
	presenter := ui.NewPresenter(cmd.OutOrStdout(), cmd.ErrOrStderr())
	ctx := cmd.Context()

	listenAddr := fmt.Sprintf("%s:%d", address, port)
	presenter.Success("Starting quality server on http://%s", listenAddr)
	presenter.Info("Access the API at http://%s/quality?mode=<mode>", listenAddr)

	mux := http.NewServeMux()
	mux.HandleFunc("/quality", handleQualityRequest(ctx, presenter, args))

	//nolint:gosec // This is a local development server, not a production service.
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func handleQualityRequest(
	ctx context.Context,
	presenter *ui.Presenter,
	args []string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("mode")
		if mode == "" {
			mode = "essential"
		}

		globals.AppLogger.Info("Handling quality request", "mode", mode)
		presenter.Info("Running quality check (mode: %s)...", mode)

		// Updated call to exported function
		results, err := RunQualityChecks(ctx, presenter, mode, args)
		if err != nil {
			presenter.Error("Failed to run quality checks: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}

		presenter.Success("Request completed.")

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(results); err != nil {
			presenter.Error("Failed to encode results to JSON: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
	}
}

//nolint:gochecknoinits // Cobra requires init() for command registration.
func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", defaultPort, "Port to serve on")
	serveCmd.Flags().StringVarP(&address, "address", "a", defaultAddress, "Address to bind to")
}
