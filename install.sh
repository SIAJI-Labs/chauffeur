#!/usr/bin/env bash
# Improved Chauffeur Installation Script
# Logic: Check if current dir is chauf git repo -> Build offline OR Clone and build

set -euo pipefail

# Configuration
WORKSPACE="${HOME}/.chauffeur"
BIN_DIR="${WORKSPACE}/bin"
BINARY_PATH="${BIN_DIR}/chauf"
REPO_URL="${CHAUF_REPO_URL:-https://github.com/SIAJI-Labs/chauffeur.git}"
REPO_BRANCH="${CHAUF_REPO_BRANCH:-release}"
PATH_LINE='export PATH="$HOME/.chauffeur/bin:$PATH"'

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Utility functions
log() {
    echo -e "${BLUE}[INSTALL]${NC} $*"
}
success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}
warn() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}
error() {
    echo -e "${RED}[ERROR]${NC} $*"
    exit 1
}

# Check if current directory is a Chauffeur git repository
is_chauf_git_repo() {
    # Check if we're in a git repository
    if ! git rev-parse --git-dir >/dev/null 2>&1; then
        return 1
    fi

    # Check if this looks like a Chauffeur repository
    if [[ -f "cli/main.go" ]] && [[ -f "go.mod" ]]; then
        # Check if go.mod contains the chauffeur module
        if grep -q "module.*chauffeur" go.mod 2>/dev/null; then
            return 0
        fi
    fi

    return 1
}

# Clone Chauffeur repository to temporary directory
clone_chauf_repo() {
    local temp_dir
    temp_dir="$(mktemp -d)"

    if ! git clone --branch "${REPO_BRANCH}" "${REPO_URL}" "${temp_dir}"; then
        error "Failed to clone Chauffeur repository from ${REPO_URL} (branch: ${REPO_BRANCH})"
    fi

    log "Cloned Chauffeur repository to ${temp_dir}"
    echo "${temp_dir}"
}

# Build Chauffeur binary from source directory
build_chauf_binary() {
    local src_dir="$1"

    log "Building Chauffeur binary from source"

    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        error "Go is required to build Chauffeur. Please install Go 1.22 or later."
    fi

    # Check Go version
    local go_version
    go_version="$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | head -1)"
    if ! go version | grep -qE 'go1\.(2[2-9]|[3-9][0-9]|[0-9]{3,})'; then
        error "Go 1.22 or later is required. Found: ${go_version}"
    fi

    # Create build metadata
    local build_ts
    build_ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    local commit_sha
    commit_sha="$(cd "${src_dir}" && git rev-parse HEAD 2>/dev/null || echo "unknown")"

    # Set build flags
    local ldflags="-X main.buildTimestamp=${build_ts}"
    if [[ "${commit_sha}" != "unknown" ]]; then
        ldflags+=" -X main.buildCommit=${commit_sha}"
    fi

    # Build the binary
    local temp_binary
    temp_binary="$(mktemp)"

    if (cd "${src_dir}" && GO111MODULE=on go build -ldflags "${ldflags}" -o "${temp_binary}" ./cli); then
        mkdir -p "${BIN_DIR}"
        mv "${temp_binary}" "${BINARY_PATH}"
        chmod +x "${BINARY_PATH}"
        success "Built Chauffeur binary successfully"
        return 0
    else
        rm -f "${temp_binary}"
        error "Failed to build Chauffeur binary. Please check the Go build output above."
    fi
}

# Add Chauffeur to PATH in shell configuration
add_to_path() {
    local shell_name="${SHELL##*/}"
    local rc_file="${HOME}/.${shell_name}rc"

    # Handle zsh specifically
    if [[ "${shell_name}" == "zsh" ]]; then
        rc_file="${HOME}/.zshrc"
    fi

    # Check if PATH is already configured
    if grep -q "${PATH_LINE}" "${rc_file}" 2>/dev/null; then
        log "Chauffeur already in PATH in ${rc_file}"
        return 0
    fi

    # Add to shell configuration
    echo "" >> "${rc_file}"
    echo "# Chauffeur" >> "${rc_file}"
    echo "${PATH_LINE}" >> "${rc_file}"

    success "Added Chauffeur to PATH in ${rc_file}"
    log "Run 'source ${rc_file}' or restart your shell to use Chauffeur"
}

# Main installation logic
main() {
    log "Starting Chauffeur installation"

    local src_dir=""

    # Check if we're already in a Chauffeur git repository
    if is_chauf_git_repo; then
        # Get current branch information
        local current_branch
        current_branch="$(git branch --show-current 2>/dev/null || echo "unknown")"

        success "Detected Chauffeur git repository in current directory"

        # Warn if not on release branch
        if [[ "${current_branch}" != "release" ]] && [[ "${current_branch}" != "unknown" ]]; then
            warn "You're installing from branch '${current_branch}', not from stable release branch"
            warn "This may contain unstable or unreleased features"
            read -p "Are you sure you want to continue? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                error "Installation cancelled by user"
            fi
        fi

        src_dir="$(pwd)"
        log "Building from local source (branch: ${current_branch})"
    else
        # Clone the repository
        echo "Cloning Chauffeur repository from ${REPO_URL} (branch: ${REPO_BRANCH})" >&2
        src_dir="$(mktemp -d)"
        if git clone --branch "${REPO_BRANCH}" "${REPO_URL}" "${src_dir}"; then
            log "Building from cloned repository (branch: ${REPO_BRANCH})"
        else
            error "Failed to clone Chauffeur repository from ${REPO_URL} (branch: ${REPO_BRANCH})"
        fi
    fi

    # Build the binary
    build_chauf_binary "${src_dir}"

    # Add to PATH
    add_to_path

    # Cleanup temporary directory if we cloned
    if [[ "${src_dir}" != "$(pwd)" ]]; then
        rm -rf "${src_dir}"
        log "Cleaned up temporary directory"
    fi

    # Success message
    success "Chauffeur installation completed successfully!"
    log "Binary installed at: ${BINARY_PATH}"
    log "Run 'source ~/.${SHELL##*/}rc' or restart your shell"

    echo ""
    echo -e "${GREEN}🎉 Chauffeur is now installed!${NC}"
    echo -e "${BLUE}Next steps:${NC}"
    echo "  1. Reload your shell: source ~/.${SHELL##*/}rc"
    echo "  2. Initialize workspace: chauf init"
    echo "  3. Install services: chauf install"
    echo "  4. Create your first project: chauf link"
    echo ""
}

# Run main function
main "$@"