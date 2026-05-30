#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────
# CURSE — Autonomous Installer (Linux / macOS / WSL)
# No API keys needed. No forced cloud auth. Just run.
# Usage:  curl -fsSL https://raw.githubusercontent.com/M523zappin/Curse-Core/master/install.sh | bash
# ─────────────────────────────────────────────────────────────
set -euo pipefail

REPO="M523zappin/Curse-Core"
BRANCH="master"
INSTALL_DIR="${HOME}/.curse-install"
BIN_DIR="${HOME}/.local/bin"
CURSE_HOME="${HOME}/curse"

CYAN='\033[0;36m'
GREEN='\033[0;32m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${CYAN}${BOLD}"
echo "  ╔══════════════════════════════════════════════╗"
echo "  ║              C U R S E                       ║"
echo "  ║  Autonomous Installer — Zero API Keys        ║"
echo "  ╚══════════════════════════════════════════════╝"
echo -e "${NC}"

# ── Detect OS ──────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"
case "${OS}" in
  Linux*)   OS="linux" ;;
  Darwin*)  OS="darwin" ;;
  *)        echo -e "${RED}Unsupported OS: ${OS}${NC}"; exit 1 ;;
esac
echo -e "  ${CYAN}•${NC} Platform: ${OS}/${ARCH}"

# ── Dependency check ───────────────────────────────────────
check_dep() {
  if ! command -v "$1" &>/dev/null; then
    echo -e "  ${CYAN}•${NC} Installing $1..."
    case "${OS}" in
      linux)
        if command -v apt-get &>/dev/null; then
          sudo apt-get install -y "$1" >/dev/null 2>&1
        elif command -v yum &>/dev/null; then
          sudo yum install -y "$1" >/dev/null 2>&1
        elif command -v pacman &>/dev/null; then
          sudo pacman -S --noconfirm "$1" >/dev/null 2>&1
        fi ;;
      darwin) brew install "$1" >/dev/null 2>&1 ;;
    esac
  fi
  echo -e "  ${GREEN}✔${NC} $1"
}

echo -e "\n  ${BOLD}Dependencies${NC}"
check_dep git
check_dep curl

# ── Clone repository ───────────────────────────────────────
if [ -d "${CURSE_HOME}/.git" ]; then
  echo -e "\n  ${CYAN}•${NC} Updating existing installation at ${CURSE_HOME}"
  cd "${CURSE_HOME}" && git pull --ff-only origin "${BRANCH}"
else
  echo -e "\n  ${CYAN}•${NC} Cloning ${REPO} → ${CURSE_HOME}"
  rm -rf "${CURSE_HOME}" 2>/dev/null || true
  git clone --depth 1 --branch "${BRANCH}" "https://github.com/${REPO}.git" "${CURSE_HOME}"
fi

# ── Install binary ─────────────────────────────────────────
mkdir -p "${BIN_DIR}"

install_via_prebuilt() {
  local binary="${CURSE_HOME}/releases/curse-${OS}-${ARCH}"
  if [ -f "$binary" ]; then
    cp "$binary" "${BIN_DIR}/curse"
    chmod +x "${BIN_DIR}/curse"
    echo -e "  ${GREEN}✔${NC} Pre-built binary deployed"
    return 0
  fi
  return 1
}

install_via_go() {
  if ! command -v go &>/dev/null; then
    echo -e "  ${CYAN}•${NC} Installing Go..."
    case "${OS}" in
      linux)
        curl -fsSL https://go.dev/dl/go1.26.0.${OS}-${ARCH}.tar.gz | sudo tar -C /usr/local -xzf -
        export PATH="/usr/local/go/bin:${PATH}" ;;
      darwin)
        if ! command -v brew &>/dev/null; then
          /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
        fi
        brew install go ;;
    esac
  fi
  echo -e "  ${CYAN}•${NC} Building from source..."
  cd "${CURSE_HOME}"
  CGO_ENABLED=0 go build -o "${BIN_DIR}/curse" ./cmd/dashboard/
  chmod +x "${BIN_DIR}/curse"
  echo -e "  ${GREEN}✔${NC} Binary built from source"
}

# Try pre-built first, then Go build
if ! install_via_prebuilt; then
  install_via_go
fi

# ── Register PATH ──────────────────────────────────────────
SHELL_CONFIG="${HOME}/.bashrc"
if [ -f "${HOME}/.zshrc" ]; then
  SHELL_CONFIG="${HOME}/.zshrc"
fi
if ! grep -q '\.local/bin' "${SHELL_CONFIG}" 2>/dev/null; then
  echo 'export PATH="${HOME}/.local/bin:${PATH}"' >> "${SHELL_CONFIG}"
  echo -e "  ${CYAN}•${NC} Added ~/.local/bin to PATH in ${SHELL_CONFIG}"
fi

# ── Done ────────────────────────────────────────────────────
echo -e "\n${GREEN}${BOLD}  ✔ CURSE installed successfully${NC}"
echo -e "\n  ${CYAN}•${NC} Binary: ${BIN_DIR}/curse"
echo -e "  ${CYAN}•${NC} Source: ${CURSE_HOME}"
echo -e "\n  ${BOLD}No API keys needed.${NC}"
echo -e "  ${BOLD}Run:${NC}  curse"
echo -e "  ${BOLD}Or:${NC}   source ${SHELL_CONFIG} && curse\n"
