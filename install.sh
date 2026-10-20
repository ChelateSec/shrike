#!/bin/bash

# Shrike Installation Script
# Part of Chelate Security - https://chelate.tech

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Chelate banner
print_banner() {
    echo -e "${CYAN}"
    cat << "EOF"
   ________          __      __
  / ____/ /_  ___  / /___ _/ /____
 / /   / __ \/ _ \/ / __ `/ __/ _ \
/ /___/ / / /  __/ / /_/ / /_/  __/
\____/_/ /_/\___/_/\__,_/\__/\___/

     Security for Agentic CI/CD
EOF
    echo -e "${NC}"
    echo -e "${BOLD}Installing Shrike - AI Agent Security Scanner${NC}"
    echo ""
}

# Print step message
print_step() {
    echo -e "${BLUE}==>${NC} ${BOLD}$1${NC}"
}

# Print success message
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Print error message
print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Print warning message
print_warning() {
    echo -e "${YELLOW}!${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        msys*|mingw*|cygwin*)
            OS="windows"
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    print_success "Detected platform: $PLATFORM"
}

# Check if running as root (for installation to /usr/local/bin)
check_sudo() {
    if [ "$EUID" -eq 0 ]; then
        SUDO=""
    else
        if command -v sudo >/dev/null 2>&1; then
            SUDO="sudo"
        else
            print_warning "sudo not found. Will attempt to install to ~/.local/bin instead"
            INSTALL_DIR="$HOME/.local/bin"
            mkdir -p "$INSTALL_DIR"
            return
        fi
    fi
    INSTALL_DIR="/usr/local/bin"
}

# Check for required commands
check_dependencies() {
    print_step "Checking dependencies..."

    local missing_deps=()

    # Check for curl or wget
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        missing_deps+=("curl or wget")
    fi

    # Check for tar (needed to extract if we package as tar.gz in future)
    if ! command -v tar >/dev/null 2>&1; then
        print_warning "tar not found (not critical)"
    fi

    if [ ${#missing_deps[@]} -eq 0 ]; then
        print_success "All required dependencies are installed"
    else
        print_error "Missing dependencies: ${missing_deps[*]}"
        echo ""
        echo "Please install the missing dependencies and try again:"
        echo ""
        echo "  Debian/Ubuntu: sudo apt-get install curl tar"
        echo "  RHEL/CentOS:   sudo yum install curl tar"
        echo "  macOS:         brew install curl"
        echo ""
        exit 1
    fi
}

# Download Shrike binary
download_shrike() {
    print_step "Downloading Shrike..."

    VERSION="latest"
    BINARY_NAME="shrike"
    if [ "$OS" = "windows" ]; then
        BINARY_NAME="shrike.exe"
    fi

    DOWNLOAD_URL="https://github.com/w3shinew/shrike/releases/${VERSION}/download/shrike-${PLATFORM}"
    if [ "$OS" = "windows" ]; then
        DOWNLOAD_URL="${DOWNLOAD_URL}.exe"
    fi

    TMP_DIR=$(mktemp -d)
    TMP_FILE="$TMP_DIR/$BINARY_NAME"

    # Try curl first, then wget
    if command -v curl >/dev/null 2>&1; then
        if curl -fsSL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
            print_success "Downloaded Shrike"
        else
            print_error "Failed to download Shrike from $DOWNLOAD_URL"
            rm -rf "$TMP_DIR"
            exit 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if wget -q "$DOWNLOAD_URL" -O "$TMP_FILE"; then
            print_success "Downloaded Shrike"
        else
            print_error "Failed to download Shrike from $DOWNLOAD_URL"
            rm -rf "$TMP_DIR"
            exit 1
        fi
    fi

    # Make executable
    chmod +x "$TMP_FILE"

    # Move to install directory
    print_step "Installing to $INSTALL_DIR..."
    if [ -n "$SUDO" ]; then
        $SUDO mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    else
        mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Cleanup
    rm -rf "$TMP_DIR"

    print_success "Shrike installed to $INSTALL_DIR/$BINARY_NAME"
}

# Verify installation
verify_installation() {
    print_step "Verifying installation..."

    if command -v shrike >/dev/null 2>&1; then
        print_success "Shrike is installed and available in PATH"
        echo ""
        shrike --version 2>/dev/null || echo "Shrike CLI ready"
    else
        print_warning "Shrike installed but not in PATH"
        echo ""
        echo "Add $INSTALL_DIR to your PATH by adding this to your shell config:"
        echo ""
        echo "  export PATH=\"$INSTALL_DIR:\$PATH\""
        echo ""

        if [ -f "$HOME/.bashrc" ]; then
            echo "For bash, run:"
            echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.bashrc"
            echo "  source ~/.bashrc"
        fi

        if [ -f "$HOME/.zshrc" ]; then
            echo "For zsh, run:"
            echo "  echo 'export PATH=\"$INSTALL_DIR:\$PATH\"' >> ~/.zshrc"
            echo "  source ~/.zshrc"
        fi
    fi
}

# Print next steps
print_next_steps() {
    echo ""
    echo -e "${BOLD}${GREEN}Installation complete!${NC}"
    echo ""
    echo -e "${BOLD}Quick Start:${NC}"
    echo ""
    echo -e "  ${CYAN}shrike scan .${NC}                    # Scan current directory"
    echo -e "  ${CYAN}shrike scan --repo owner/name${NC}   # Scan GitHub repository"
    echo -e "  ${CYAN}shrike scan --github-org org${NC}    # Scan entire organization"
    echo -e "  ${CYAN}shrike --help${NC}                   # Show all options"
    echo ""
    echo -e "${BOLD}Resources:${NC}"
    echo ""
    echo "  Documentation: https://github.com/w3shinew/shrike"
    echo "  Website:       https://chelate.tech"
    echo "  Issues:        https://github.com/w3shinew/shrike/issues"
    echo ""
    echo -e "${BOLD}Need help?${NC} Contact: security@chelate.tech"
    echo ""
}

# Main installation flow
main() {
    print_banner
    detect_platform
    check_sudo
    check_dependencies
    download_shrike
    verify_installation
    print_next_steps
}

# Run main
main
