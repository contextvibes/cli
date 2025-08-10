#!/bin/bash
#
# This script proactively fetches the correct Nix source hash for the current
# git commit, preventing the need for the "fail-and-fix" cycle.

set -e

# --- Get the latest commit hash ---
COMMIT_HASH=$(git rev-parse HEAD)
if [ -z "$COMMIT_HASH" ]; then
    echo "❌ Error: Could not get git commit hash. Are you in a git repository?"
    exit 1
fi
echo "==> Found latest local commit hash: $COMMIT_HASH"
echo "==> IMPORTANT: Ensure this commit has been pushed to the remote repository!"
echo

# --- Use nix-shell to run the prefetch command ---
echo "==> Prefetching source hash from GitHub in a pure environment..."
# CRITICAL FIX: Add 'nix' to the -p flag to ensure 'nix-build' is available
# inside the pure shell for the prefetch script to use.
PREFETCH_OUTPUT=$(nix-shell -p nix-prefetch-github nix --pure --run "nix-prefetch-github --rev $COMMIT_HASH contextvibes cli")

# --- Extract the hash from the JSON output using jq. ---
SOURCE_HASH=$(echo "$PREFETCH_OUTPUT" | jq -r '.hash')

if [ -z "$SOURCE_HASH" ] || [ "$SOURCE_HASH" == "null" ]; then
    echo "❌ Error: Failed to prefetch hash. The commit might not be pushed, or the repo is private."
    echo "Raw output:"
    echo "$PREFETCH_OUTPUT"
    exit 1
fi

echo
echo "✅ Success! Here is the information for your .nix file:"
echo "--------------------------------------------------------"
echo "rev = \"$COMMIT_HASH\";"
echo "hash = \"$SOURCE_HASH\";"
echo "--------------------------------------------------------"
echo
echo "NOTE: This script provides the 'hash' for the source code."
echo "You will still need to discover the 'vendorHash' for Go modules via the one-time build failure after updating these values."

