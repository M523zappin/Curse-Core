#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────
# CURSE — One-Line Installer (Linux / macOS / Windows WSL)
# Just copy-paste this ONE command:
#   curl -fsSL https://curse.sh/install | bash
# ─────────────────────────────────────────────────────────────
set -euo pipefail

# Colors
CYAN='\033[0;36m'
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${CYAN}${BOLD}"
echo "  ╔══════════════════════════════════════════╗"
echo "  ║         C U R S E  Installer            ║"
echo "  ║    Zero API Keys • 100% Offline Ready   ║"
echo "  ╚══════════════════════════════════════════╝"
echo -e "${NC}"

# Detect OS
OS="$(uname -s)"
case "${OS}" in
  Linux*)   OS="linux" ;;
  Darwin*)  OS="darwin" ;;
  *)        echo -e "${RED}Unsupported OS: ${OS}${NC}"; exit 1 ;;
esac

ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64*)  ARCH="amd64" ;;
  aarch64*|arm64*) ARCH="arm64" ;;
  *)        ARCH="amd64" ;;
esac

echo -e "  ${CYAN}•${NC} Platform: ${OS}/${ARCH}"

# Install location
BIN_DIR="${HOME}/.local/bin"
if [ ! -d "$BIN_DIR" ]; then
  mkdir -p "$BIN_DIR"
fi

# Check for curl
if ! command -v curl &> /dev/null; then
  echo -e "${RED}Error: curl is required but not installed.${NC}"
  exit 1
fi

# Download latest release
REPO="M523zappin/Curse-Core"
VERSION=$(curl -s https://api.github.com/repos/${REPO}/releases/latest 2>/dev/null | grep '"tag_name"' | cut -d'"' -f4 || echo "latest")
FILENAME="curse-${OS}-${ARCH}"

echo -e "  ${CYAN}•${NC} Downloading CURSE..."
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

# Try direct download first
if curl -fSL --progress-bar "$DOWNLOAD_URL" -o "${BIN_DIR}/curse" 2>/dev/null; then
  chmod +x "${BIN_DIR}/curse"
  echo -e "  ${GREEN}✓${NC} Installed to ${BIN_DIR}/curse"
else
  # Fallback to curl script
  echo -e "  ${YELLOW}!${NC} Binary not found, building from source..."
  curl -fsSL "https://raw.githubusercontent.com/${REPO}/main/scripts/install-source.sh" 2>/dev/null | bash || {
    echo -e "${RED}Installation failed. Try: go install github.com/${REPO}/cmd/dashboard@latest${NC}"
    exit 1
  }
fi

# Add to PATH if needed
SHELL_RC="${HOME}/.bashrc"
if [ -f "${HOME}/.zshrc" ]; then SHELL_RC="${HOME}/.zshrc"; fi
if [ -f "${HOME}/.config/fish/config.fish" ]; then SHELL_RC=""; fi

if [[ ":$PATH:" != *":${BIN_DIR}:"* ]]; then
  echo "" >> "$SHELL_RC" 2>/dev/null || true
  echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$SHELL_RC" 2>/dev/null || true
  echo -e "  ${YELLOW}!${NC} Added ${BIN_DIR} to PATH in ${SHELL_RC}"
  echo -e "  ${CYAN}→${NC} Run: source ${SHELL_RC}  or restart your terminal"
fi

echo ""
echo -e "${GREEN}${BOLD}✓ CURSE installed successfully!${NC}"
echo ""
echo -e "  ${CYAN}Quick Start:${NC}"
echo -e "    ${BOLD}curse${NC}                  # Start CURSE"
echo -e "    ${BOLD}curse --help${NC}            # Show help"
echo ""
echo -e "  ${CYAN}Examples:${NC}"
echo -e "    >>> create a REST API in Go"
echo -e "    >>> add authentication middleware"
echo -e "    >>> write unit tests"
echo ""
