package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestVersionCmd(t *testing.T) {
	// AppVersion is initialized in root.go's init() function.
	// For isolated testing, ensure it reflects the expected version if necessary,
	// but typically package-level init() should handle this.
	// If there were issues with init order across files (unlikely for same package),
	// one might set AppVersion = "0.1.0" here for explicit test control.

	testCases := []struct {
		name           string
		appVersion     string // To explicitly set for the test run
		expectedOutput string
	}{
		{
			name:           "Check Version Output",
			appVersion:     "0.1.0", // Ensure this matches what the test expects
			expectedOutput: "SUMMARY:\n  Context Vibes CLI Version: 0.1.0\n\n",
		},
		{
			name:           "Check Different Version Output",
			appVersion:     "v0.0.2-beta",
			expectedOutput: "SUMMARY:\n  Context Vibes CLI Version: v0.0.2-beta\n\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set the global AppVersion for this specific test case run
			originalAppVersion := AppVersion // Store original to restore later
			AppVersion = tc.appVersion
			defer func() { AppVersion = originalAppVersion }() // Restore original AppVersion

			// Create a new root command for each test case to ensure isolation
			// and to properly wire up our versionCmd.
			// The actual rootCmd from root.go is not used directly to avoid init complexities
			// with logging or other persistent flags in a minimal test setup.
			testRootCmd := &cobra.Command{Use: "contextvibes"}

			// The versionCmd's init() function (which adds it to the global rootCmd)
			// will have run. To test it in isolation, we create a new root
			// and add the versionCmd to *it*.
			// Re-create versionCmd or ensure it's clean if it has state.
			// For this simple version command, just adding the package-level versionCmd is fine.
			testRootCmd.AddCommand(versionCmd)

			out := new(bytes.Buffer)
			testRootCmd.SetOut(out) // Capture standard output
			testRootCmd.SetErr(out) // Optionally capture stderr if needed, though version writes to stdout

			testRootCmd.SetArgs([]string{"version"})

			err := testRootCmd.Execute()
			if err != nil {
				t.Fatalf("Execute() failed unexpectedly: %v", err)
			}
			assert.Equal(t, tc.expectedOutput, out.String())
		})
	}
}
