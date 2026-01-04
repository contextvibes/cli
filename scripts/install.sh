#!/bin/bash
set -e

REPO="contextvibes/cli"
LATEST_URL=$(curl -s https://api.github.com/repos/$REPO/releases/latest | grep "browser_download_url" | grep "$(uname -s)_$(uname -m)" | cut -d '"' -f 4)

if [ -z "$LATEST_URL" ]; then
  echo "Could not find a release for your OS/Arch."
  exit 1
fi

echo "Downloading $LATEST_URL..."
curl -L -o contextvibes.tar.gz "$LATEST_URL"

echo "Installing..."
tar -xzf contextvibes.tar.gz contextvibes
chmod +x contextvibes
mv contextvibes $HOME/go/bin/ # Or /usr/local/bin

echo "Done! Run 'contextvibes version' to verify."
rm contextvibes.tar.gz
