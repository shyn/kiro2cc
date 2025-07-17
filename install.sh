#!/bin/bash
set -e
#
# This script handles downloading and installing the latest kiro2cc binary.
#
# Usage:
#   curl -sSL https://raw.githubusercontent.com/deepwind/kiro2cc/main/install.sh | bash
#

# GitHub repository
REPO="deepwind/kiro2cc"

# Function to detect OS and architecture
detect_platform() {
    local os arch
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    case "$os" in
        linux)
            os="linux"
            ;;
        darwin)
            os="macos"
            ;;
        *)
            echo "Unsupported OS: $os" >&2
            exit 1
            ;;
    esac

    case "$arch" in
        x86_64)
            arch="amd64"
            ;;
        arm64 | aarch64)
            arch="arm64"
            ;;
        *)
            echo "Unsupported architecture: $arch" >&2
            exit 1
            ;;
    esac
    echo "$os-$arch"
}

# Main installation logic
main() {
    local platform
    platform=$(detect_platform)

    echo "Platform detected: $platform"

    # Get the latest release version from GitHub API
    echo "Fetching the latest release version..."
    local latest_release_tag
    latest_release_tag=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$latest_release_tag" ]; then
        echo "Error: Could not fetch the latest release tag." >&2
        exit 1
    fi

    local version
    version=${latest_release_tag#v} # Remove 'v' prefix
    echo "Latest version is $version"

    # Construct the asset name and download URL
    local asset_name="kiro2cc-${version}-${platform}"
    local download_url="https://github.com/$REPO/releases/download/$latest_release_tag/$asset_name"

    # Download the binary to a temporary file
    local tmp_file
    tmp_file=$(mktemp)
    echo "Downloading from: $download_url"
    curl -L -o "$tmp_file" "$download_url"

    if ! [ -s "$tmp_file" ]; then
        echo "Error: Download failed or the file is empty." >&2
        rm "$tmp_file"
        exit 1
    fi

    # Make it executable and install it
    local install_dir="$HOME/.local/bin"
    local install_path="$install_dir/kiro2cc"

    # Create the installation directory if it doesn't exist
    mkdir -p "$install_dir"

    echo "Installing to $install_path..."
    chmod +x "$tmp_file"
    mv "$tmp_file" "$install_path"

    echo "kiro2cc version $version has been installed successfully!"
    echo ""
    echo "Please make sure '$install_dir' is in your PATH."
    echo "You can add it to your shell configuration file (e.g., ~/.bashrc, ~/.zshrc) with:"
    echo "export PATH=\"
$HOME/.local/bin:$PATH\"
"
    echo ""
    echo "You can now run 'kiro2cc' from your terminal (you may need to restart it)."
}

# Run the script
main
