#!/usr/bin/env bash
# ─────────────────────────────────────────────────────────────
# CURSE — Zero-Touch Installer (Linux / macOS / WSL)
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
echo "  ║  Zero-Touch Installer                        ║"
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

if [ -f "${CURSE_HOME}/releases/curse-${OS}-${ARCH}" ]; then
  cp "${CURSE_HOME}/releases/curse-${OS}-${ARCH}" "${BIN_DIR}/curse"
elif command -v go &>/dev/null; then
  echo -e "  ${CYAN}•${NC} Building from source..."
  cd "${CURSE_HOME}"
  CGO_ENABLED=0 go build -o "${BIN_DIR}/curse" ./cmd/dashboard/
else
  echo -e "  ${CYAN}•${NC} Building with Go (will install Go first)..."
  if ! command -v go &>/dev/null; then
    case "${OS}" in
      linux)
        curl -fsSL https://go.dev/dl/go1.23.4.${OS}-${ARCH}.tar.gz | sudo tar -C /usr/local -xzf -
        export PATH="/usr/local/go/bin:${PATH}" ;;
      darwin)
        brew install go ;;
    esac
  fi
  cd "${CURSE_HOME}"
  CGO_ENABLED=0 go build -o "${BIN_DIR}/curse" ./cmd/dashboard/
fi

chmod +x "${BIN_DIR}/curse"

# ── Register PATH ──────────────────────────────────────────
SHELL_CONFIG="${HOME}/.bashrc"
if [ -f "${HOME}/.zshrc" ]; then
  SHELL_CONFIG="${HOME}/.zshrc"
fi
if ! grep -q '\.local/bin' "${SHELL_CONFIG}" 2>/dev/null; then
  echo 'export PATH="${HOME}/.local/bin:${PATH}"' >> "${SHELL_CONFIG}"
  echo -e "  ${CYAN}•${NC} Added ~/.local/bin to PATH in ${SHELL_CONFIG}"
fi

# ── Bootstrap .env ──────────────────────────────────────────
if [ ! -f "${CURSE_HOME}/.env" ]; then
  cp "${CURSE_HOME}/.env.example" "${CURSE_HOME}/.env"
  echo -e "  ${CYAN}•${NC} Created ${CURSE_HOME}/.env (edit with your API keys)"
fi

# ── GitHub Auth handshake ──────────────────────────────────
if ! command -v gh &>/dev/null; then
  echo -e "  ${CYAN}•${NC} Installing GitHub CLI..."
  case "${OS}" in
    linux)
      curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg
      echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null
      sudo apt-get update && sudo apt-get install -y gh ;;
    darwin) brew install gh ;;
  esac
fi

if ! gh auth status &>/dev/null; then
  echo -e "\n  ${CYAN}•${NC} Authenticating with GitHub..."
  gh auth login --web || true
fi

# ── Done ────────────────────────────────────────────────────
echo -e "\n${GREEN}${BOLD}  ✔ CURSE installed successfully${NC}"
echo -e "\n  ${CYAN}•${NC} Binary: ${BIN_DIR}/curse"
echo -e "  ${CYAN}•${NC} Source: ${CURSE_HOME}"
echo -e "  ${CYAN}•${NC} Config: ${CURSE_HOME}/.env"
echo -e "\n  ${BOLD}Run:${NC}  curse"
echo -e "  ${BOLD}Or:${NC}   source ${SHELL_CONFIG} && curse\n"
