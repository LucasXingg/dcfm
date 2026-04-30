#!/usr/bin/env bash
set -e

# Default settings
REPO="LucasXingg/dcfm"
BIN_NAME="dcfm"
INSTALL_DIR="/usr/local/bin"

# Determine OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     OS_NAME="Linux";;
    Darwin*)    OS_NAME="macOS";;
    CYGWIN*|MINGW*|MSYS*) OS_NAME="Windows";;
    *)          echo "Unsupported OS: ${OS}"; exit 1;;
esac

# Determine Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
    x86_64|amd64) ARCH_NAME="x86_64";;
    aarch64|arm64) ARCH_NAME="arm64";;
    i386|i686|i386) ARCH_NAME="i386";;
    *)            echo "Unsupported architecture: ${ARCH}"; exit 1;;
esac

echo "Fetching latest release for ${REPO}..."
RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
RELEASE_DATA=$(curl -s "$RELEASE_URL")

# Check if curl failed or returned an error from GitHub API
if echo "$RELEASE_DATA" | grep -q "API rate limit exceeded"; then
    echo "GitHub API rate limit exceeded. Please try again later or install manually."
    exit 1
fi

VERSION=$(echo "$RELEASE_DATA" | grep '"tag_name":' | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Error: Could not determine latest version."
    exit 1
fi

echo "Latest version is ${VERSION}"

# Set the correct extension
if [ "${OS_NAME}" = "Windows" ]; then
    EXT="zip"
else
    EXT="tar.gz"
fi

# Extract the browser_download_url
DOWNLOAD_URL=$(echo "$RELEASE_DATA" | grep '"browser_download_url":' | grep -i "${OS_NAME}" | grep -i "${ARCH_NAME}" | grep "${EXT}" | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$DOWNLOAD_URL" ]; then
    # Fallback check for Darwin
    if [ "${OS_NAME}" = "macOS" ]; then
        DOWNLOAD_URL=$(echo "$RELEASE_DATA" | grep '"browser_download_url":' | grep -i "Darwin" | grep -i "${ARCH_NAME}" | grep "${EXT}" | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/')
    fi
    
    if [ -z "$DOWNLOAD_URL" ]; then
        echo "Error: Could not find a release asset for OS: ${OS_NAME} and Arch: ${ARCH_NAME}"
        exit 1
    fi
fi

echo "Downloading from ${DOWNLOAD_URL}..."

TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"

if [ "${EXT}" = "tar.gz" ]; then
    curl -sL "$DOWNLOAD_URL" | tar xz
elif [ "${EXT}" = "zip" ]; then
    curl -sLo "${BIN_NAME}.zip" "$DOWNLOAD_URL"
    unzip -q "${BIN_NAME}.zip"
fi

echo "Installing ${BIN_NAME} to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
    mv "${BIN_NAME}" "${INSTALL_DIR}/"
else
    echo "Sudo permissions required to install to ${INSTALL_DIR}"
    sudo mv "${BIN_NAME}" "${INSTALL_DIR}/"
fi

cd - > /dev/null
rm -rf "$TMP_DIR"

echo "Installation complete! Run '${BIN_NAME} --help' to get started."
