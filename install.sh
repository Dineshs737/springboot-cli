#!/usr/bin/env bash
set -e

# Colors for terminal output
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}Installing SpringCLI...${NC}"

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS_NAME=linux;;
    Darwin*)    OS_NAME=darwin;;
    *)          echo -e "${RED}Unsupported OS: ${OS}${NC}"; exit 1;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64)  ARCH_NAME=amd64;;
    arm64)   ARCH_NAME=arm64;;
    aarch64) ARCH_NAME=arm64;;
    *)       echo -e "${RED}Unsupported Architecture: ${ARCH}${NC}"; exit 1;;
esac

REPO="Dineshs737/springboot-cli"
BINARY_NAME="springcli-${OS_NAME}-${ARCH_NAME}"

# Fetch the latest release version from GitHub API
echo "Fetching latest version from GitHub..."
TMP_LATEST=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$TMP_LATEST" ]; then
    echo -e "${RED}Failed to fetch the latest release version. Maybe GitHub rate limit hit?${NC}"
    # Fallback to the known tag
    TMP_LATEST="v1.0.0" 
fi

echo "Downloading ${BINARY_NAME} version ${TMP_LATEST}..."

# Download URL
URL="https://github.com/${REPO}/releases/download/${TMP_LATEST}/${BINARY_NAME}"

# Download the binary to a temporary file
curl -L --fail --progress-bar -o springcli_tmp "$URL"

if [ $? -ne 0 ]; then
    echo -e "${RED}Failed to download the binary from ${URL}${NC}"
    rm -f springcli_tmp
    exit 1
fi

# Make it executable
chmod +x springcli_tmp

# Move to /usr/local/bin
echo "Moving to /usr/local/bin (may require sudo password)..."
sudo mv springcli_tmp /usr/local/bin/springcli

if [ $? -eq 0 ]; then
    echo -e "${GREEN}SpringCLI installed successfully!${NC}"
    echo "Run 'springcli' to get started."
else
    echo -e "${RED}Failed to move the binary. Do you have permissions?${NC}"
    rm -f springcli_tmp
    exit 1
fi
