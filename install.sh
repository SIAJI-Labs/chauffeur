#!/usr/bin/env bash
# Avoid 'unbound variable' errors with arrays by temporarily disabling nounset
set -o pipefail

# Try to get SCRIPT_DIR, but handle cases where BASH_SOURCE is undefined
if [[ -n "${BASH_SOURCE:-}" ]] && [[ ${#BASH_SOURCE[@]} -gt 0 ]]; then
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
else
  # If BASH_SOURCE is not available (e.g., piped via curl), use /tmp as default
  SCRIPT_DIR="/tmp"
fi

# Now we can safely enable nounset for the rest of the script
set -u



WORKSPACE="${HOME}/.chauffeur"
BIN_DIR="${WORKSPACE}/bin"
SHIMS_DIR="${BIN_DIR}/shims"
PATH_LINE='export PATH="$HOME/.chauffeur/bin:$PATH"'
SHELL_NAME="$(basename "${SHELL:-}")"
# For curl installs, SCRIPT_DIR will be /tmp, so LOCAL_SRC_DIR won't exist
if [[ -n "${SCRIPT_DIR}" && "${SCRIPT_DIR}" != "/tmp" ]]; then
  LOCAL_SRC_DIR="${SCRIPT_DIR}/cli"
else
  LOCAL_SRC_DIR=""
fi
REPO_URL="${CHAUF_REPO_URL:-https://github.com/SIAJI-Labs/chauffeur.git}"

section() {
  printf '\n[ INSTALL ] %s\n' "$(echo "$1" | tr '[:lower:]' '[:upper:]')"
}

info() {
  printf '    - %s\n' "$*"
}

is_curl_install() {
  # Check if we're being piped from curl (stdin not a terminal)
  [[ ! -t 0 ]] || [[ "${CHAUF_CURL_INSTALL:-}" == "1" ]]
}

check_existing_chauf() {
  # First check if workspace is initialized
  if [[ ! -f "${WORKSPACE}/config/chauffeur.yaml" ]]; then
    echo "Chauffeur workspace is not initialized."
    echo "Running 'chauf init' to create default configuration..."
    echo ""
    
    # Try to initialize workspace if we have a binary
    local workspace_chauf="${BIN_DIR}/chauf"
    if [[ -x "${workspace_chauf}" ]]; then
      if "${workspace_chauf}" init; then
        echo "Workspace initialized successfully."
        echo ""
      else
        warn "Failed to initialize workspace with existing binary."
        echo "Will create basic configuration..."
        create_basic_config
      fi
    else
      info "No chauf binary available yet, creating basic configuration..."
      create_basic_config
    fi
    # Continue with installation even if init failed
  fi

  # Now check for existing chauf binary (after potential workspace init)
  local existing_chauf=""
  local existing_version=""
  
  # First check the workspace binary location (preferred for current workspace)
  workspace_chauf="${BIN_DIR}/chauf"
  if [[ -x "${workspace_chauf}" ]]; then
    existing_chauf="${workspace_chauf}"
    existing_version="$("${workspace_chauf}" --version 2>/dev/null || echo "unknown")"
  # Then check PATH for other installation
  elif command -v chauf >/dev/null 2>&1; then
    existing_chauf="$(which chauf)"
    existing_version="$(chauf --version 2>/dev/null || echo "unknown")"
  fi
  
  if [[ -n "${existing_chauf}" ]]; then
    echo "Chauffeur is already installed:"
    echo "  - Binary location: ${existing_chauf}"
    echo "  - Version: ${existing_version}"
    echo ""
    
    # Check if existing binary matches current workspace
    if [[ "${existing_chauf}" == "${workspace_chauf}" ]]; then
      echo "Workspace is already set up and binary is current."
      echo "To reinstall/rebuild:"
      echo "  - Recompile manually: go build -o ${workspace_chauf} ./cli"
      echo "  - Or remove existing: rm ${workspace_chauf} && ./install.sh"
      echo ""
      echo "To completely remove Chauffeur:"
      echo "  chauf uninstall --purge"
      return 2  # Special return code for "already installed"
    else
      echo "Chauffeur is installed but not from this workspace."
      echo "Workspace binary will be placed at: ${workspace_chauf}"
      echo "This will coexist with your existing installation."
      echo ""
      return 0  # Continue with installation
    fi
  fi
  
  return 0  # No existing installation, continue
}

check_go_requirements() {
  if ! command -v go >/dev/null 2>&1; then
    warn "Go is required to build Chauffeur CLI."
    warn "Please install Go 1.22 or newer:"
    warn "  - On Arch Linux: sudo pacman -S go"
    warn "  - On Ubuntu/Debian: sudo apt install golang-go"
    warn "  - On macOS (with Homebrew): brew install go"
    warn "  - Visit: https://golang.org/dl/"
    return 1
  fi

  local go_version
  go_version=$(go version 2>/dev/null | sed -E 's/go version go([0-9]+\.[0-9]+).*/\1/')
  
  # Check if Go version is 1.22 or higher (comparing major.minor versions)
  if [[ -n "${go_version}" ]]; then
    local major="${go_version%%.*}"  # Get major version (1)
    local minor="${go_version#*.}"  # Get minor version (22 or higher)
    local min_major=1
    local min_minor=22
    
    if [[ "${major}" -lt "${min_major}" ]] || [[ "${major}" -eq "${min_major}" && "${minor}" -lt "${min_minor}" ]]; then
      warn "Found Go ${go_version}. Chauffeur requires Go 1.22 or newer."
      warn "Please upgrade your Go installation."
      return 1
    else
      success "Found Go ${go_version} ✓"
      return 0
    fi
  else
    warn "Unable to determine Go version."
    return 1
  fi
}

success() {
  printf '    [OK] %s\n' "$*"
}

warn() {
  printf '    [WARN] %s\n' "$*" >&2
}

error() {
  printf '    [ERR] %s\n' "$*" >&2
}

ensure_directories() {
  local -a dirs=(
    "${BIN_DIR}"
    "${SHIMS_DIR}"
    "${WORKSPACE}/config"
    "${WORKSPACE}/projects"
    "${WORKSPACE}/php"
    "${WORKSPACE}/nginx/bin"
    "${WORKSPACE}/nginx/etc"
    "${WORKSPACE}/nginx/sites-available"
    "${WORKSPACE}/nginx/sites-enabled"
    "${WORKSPACE}/nginx/conf.d"
  )

  for dir in "${dirs[@]}"; do
    mkdir -p "${dir}"
  done
}

trim_trailing_whitespace_and_empty_lines() {
  local rc_file="$1"
  
  # Remove trailing whitespace and clean up trailing empty lines
  if [[ -f "${rc_file}" ]]; then
    local temp_file
    temp_file="$(mktemp)"
    
    # Remove trailing whitespace from each line
    sed 's/[[:space:]]*$//' "${rc_file}" > "${temp_file}"
    
    # Remove trailing empty lines using a simple approach
    # Keep removing empty lines from the end as long as they exist
    while [[ -s "${temp_file}" ]]; do
      last_line=$(tail -n 1 "${temp_file}")
      if [[ -n "${last_line}" ]]; then
        break  # Last line is not empty, we're done
      fi
      # Remove the last empty line
      head -n -1 "${temp_file}" > "${temp_file}.tmp"
      mv "${temp_file}.tmp" "${temp_file}"
    done
    
    mv "${temp_file}" "${rc_file}"
  fi
}

ensure_path_export() {
  local rc_file=""

  case "${SHELL_NAME}" in
    bash)
      rc_file="${HOME}/.bashrc"
      ;;
    zsh)
      rc_file="${HOME}/.zshrc"
      ;;
    "")
      warn "Detected empty SHELL variable; unable to configure PATH automatically."
      warn "Add ${BIN_DIR} to your PATH manually."
      return
      ;;
    *)
      warn "Shell '${SHELL_NAME}' is not managed automatically."
      warn "Add ${BIN_DIR} to your PATH manually."
      return
      ;;
  esac

  touch "${rc_file}"
  if grep -qxF "${PATH_LINE}" "${rc_file}"; then
    info "PATH already contains ${BIN_DIR}; skipping ${rc_file} update"
  else
    # Clean up trailing whitespace to prevent accumulation
    trim_trailing_whitespace_and_empty_lines "${rc_file}"
    
    # Check if the file ends with a newline and if the last line is empty
    local last_line_nl last_line_content
    if [[ -f "${rc_file}" && -s "${rc_file}" ]]; then
      last_line_nl=$(tail -n 1 "${rc_file}")
      # Check if last line is empty (contains only whitespace or nothing)
      if [[ "${last_line_nl//[[:space:]]/}" == "" ]]; then
        # Last line is empty/whitespace, just add PATH without extra newline
        printf '%s\n' "${PATH_LINE}" >>"${rc_file}"
      else
        # Last line has content, add PATH with preceding newline
        printf '\n%s\n' "${PATH_LINE}" >>"${rc_file}"
      fi
    else
      # File is empty or doesn't exist, just add PATH line
      printf '%s\n' "${PATH_LINE}" >>"${rc_file}"
    fi
    info "Added ${BIN_DIR} to PATH via ${rc_file}"
  fi
}

build_local_chauf() {
  local output_path="$1"

  if [[ ! -d "${LOCAL_SRC_DIR}" ]]; then
    return 1
  fi

  if ! command -v go >/dev/null 2>&1; then
    warn "Go toolchain not found; cannot build chauf from source."
    return 4
  fi

  local tmp_binary
  tmp_binary="$(mktemp)"
  local build_ts
  build_ts="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  local commit_sha
  commit_sha="$(cd "${SCRIPT_DIR}" && git rev-parse HEAD 2>/dev/null || printf 'unknown')"

  local ldflags="-X main.buildTimestamp=${build_ts}"
  if [[ -n "${commit_sha}" ]]; then
    ldflags+=" -X main.buildCommit=${commit_sha}"
  fi

  if (cd "${SCRIPT_DIR}" && GO111MODULE=on go build -ldflags "${ldflags}" -o "${tmp_binary}" ./cli); then
    install -m 0755 "${tmp_binary}" "${output_path}"
    rm -f "${tmp_binary}"
    success "Built chauf binary from Go sources"
    return 0
  else
    rm -f "${tmp_binary}"
    warn "Go build failed; run 'go build ./cli' manually for details."
    return 5
  fi
}

ensure_chauf_binary() {
  local binary_path="${BIN_DIR}/chauf"
  local downloaded=0
  local download_error=0
  local build_status=0

  if [[ -x "${binary_path}" ]]; then
    info "Found existing chauf binary at ${binary_path}"
    return 0
  fi

  if [[ -n "${CHAUF_RELEASE_URL:-}" ]]; then
    if ! command -v curl >/dev/null 2>&1; then
      warn "curl is required to download chauf from CHAUF_RELEASE_URL."
      warn "Install curl or place your built chauf binary at ${binary_path}."
      download_error=3
    else
      local tmp_file
      tmp_file="$(mktemp)"
      if curl -fL "${CHAUF_RELEASE_URL}" -o "${tmp_file}"; then
        mv "${tmp_file}" "${binary_path}"
        chmod +x "${binary_path}"
        success "Downloaded chauf binary"
        downloaded=1
        return 0
      else
        rm -f "${tmp_file}"
        warn "Failed to download chauf from ${CHAUF_RELEASE_URL}"
        download_error=3
      fi
    fi
  fi

  if build_local_chauf "${binary_path}"; then
    return 0
  else
    build_status=$?
  fi

  if [[ -x "${binary_path}" ]]; then
    if [[ "${downloaded}" -eq 1 ]]; then
      info "chauf binary downloaded"
    else
      info "Make sure ${binary_path} remains executable."
    fi
    return 0
  fi

  if [[ "${download_error}" -ne 0 ]]; then
    return "${download_error}"
  fi

  if [[ "${build_status}" -eq 5 ]]; then
    return "${build_status}"
  fi

  warn "Place your built chauf binary at ${binary_path} or set CHAUF_RELEASE_URL to download."
  return 2
}

create_basic_config() {
  local config_dir="${WORKSPACE}/config"
  local config_file="${config_dir}/chauffeur.yaml"
  
  # Ensure config directory exists
  mkdir -p "${config_dir}"
  
  # Create basic configuration with user-space ports
  cat > "${config_file}" << 'EOF'
# Chauffeur Configuration File
# This file controls global Chauffeur settings and port management

# Configuration version (do not modify)
version: 1

# Enable/disable telemetry data collection
telemetry: false

# Workspace directory where Chauffeur stores its data
workspace_dir: ~/.chauffeur

# Nginx web server configuration  
nginx:
  enable: true
  # Custom ports to avoid conflicts with system services
  http_port: 8080     # HTTP port (user-space)
  https_port: 8443    # HTTPS port (user-space)

# PHP runtime configuration
php:
  default: "8.3"

# Port management settings
ports:
  # Port range for automatic port allocation
  start_range: 8080
  end_range: 8099
  
  # How to handle port conflicts:
  # - "prompt": Ask user to select alternative ports (default)
  # - "auto":  Automatically select available ports
  # - "fail":  Fail if ports are in use
  conflict_resolution: "prompt"
  
  # Fallback ports for each service
  nginx_http_fallback: 8080
  nginx_https_fallback: 8443
  php_fpm_fallback: 9000

# Directory where Chauffeur stores project configurations
projects_dir: ~/.chauffeur/projects
EOF
  
  info "Created basic configuration with user-space ports"
  success "Configuration file created: ${config_file}"
  info "Default ports: Nginx HTTP 8080, Nginx HTTPS 8443"
}

is_git_repo() {
  # If SCRIPT_DIR is empty, we're definitely not in a git repo
  if [[ -z "${SCRIPT_DIR}" ]]; then
    return 1
  fi
  
  if [[ -d "${SCRIPT_DIR}/.git" ]]; then
    return 0
  fi
  if command -v git >/dev/null 2>&1 && git -C "${SCRIPT_DIR}" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    return 0
  fi
  return 1
}

bootstrap_from_remote() {
  section "Bootstrap"
  
  # Check Go requirements before cloning
  if ! check_go_requirements; then
    section "Failed"
    error "Go is required to build Chauffeur CLI from source."
    error "Please install Go 1.22 or newer and try again."
    exit 1
  fi
  
  if ! command -v git >/dev/null 2>&1; then
    error "git is required to clone ${REPO_URL}."
    exit 1
  fi

  local tmp_dir
  tmp_dir="$(mktemp -d)"
  info "Cloning Chauffeur repository"
  if ! git clone --depth 1 --quiet "${REPO_URL}" "${tmp_dir}/chauffeur"; then
    rm -rf "${tmp_dir}"
    error "Failed to clone ${REPO_URL}."
    exit 1
  fi

  info "Running installer from cloned repository"
  if CHAUF_BOOTSTRAP=1 "${tmp_dir}/chauffeur/install.sh" "$@"; then
    rm -rf "${tmp_dir}"
    # After successful bootstrap, check if workspace exists and suggest init
    if [[ -d "${WORKSPACE}/config" && ! -f "${WORKSPACE}/config/chauffeur.yaml" ]]; then
      echo ""
      section "Workspace Ready"
      info "Chauffeur workspace is ready but not initialized."
      info "Run 'chauf init' to create default configuration:"
      echo "  chauf init"
    fi
    exit 0
  else
    local status=$?
    rm -rf "${tmp_dir}"
    exit "${status}"
  fi
}

run_local_install() {
  # Ensure workspace directory exists first (needed before check_existing_chauf)
  if [[ ! -d "${WORKSPACE}" ]]; then
    info "Creating workspace directory: ${WORKSPACE}"
    mkdir -p "${WORKSPACE}"
  fi

  # Check if chauf is already installed (skip for bootstrap since we checked before cloning)
  if [[ "${CHAUF_BOOTSTRAP:-}" != "1" ]]; then
    section "Installation Check"
    check_existing_chauf
    local check_result=$?
    if [[ ${check_result} -eq 2 ]]; then
      # Function returned 2 - chauf already installed in this workspace
      exit 0
    fi
  fi

  # For local installation, check Go requirements first
  if [[ "${CHAUF_BOOTSTRAP:-}" != "1" ]]; then
    section "Requirements"
    if ! check_go_requirements; then
      section "Failed"
      error "Go is required to build Chauffeur CLI from source."
      error "Please install Go 1.22 or newer and try again."
      exit 1
    fi
  fi

  # If we're in bootstrap mode, re-evaluate SCRIPT_DIR since we're now in the cloned repo
  if [[ "${CHAUF_BOOTSTRAP:-}" == "1" ]]; then
    if [[ -n "${BASH_SOURCE:-}" ]] && [[ ${#BASH_SOURCE[@]} -gt 0 ]]; then
      SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
      # Update LOCAL_SRC_DIR to point to the cloned repo's cli directory
      LOCAL_SRC_DIR="${SCRIPT_DIR}/cli"
      info "Updated SCRIPT_DIR to cloned repo: ${SCRIPT_DIR}"
    fi
  fi

  section "Workspace"
  info "Creating workspace directories"
  ensure_directories
  info "Ensuring PATH exports"
  ensure_path_export

  section "Binary"
  local binary_ready=0
  if ensure_chauf_binary; then
    binary_ready=1
  else
    case "$?" in
      2)
        binary_ready=0
        ;;
      *)
        exit "$?"
        ;;
    esac
  fi

  section "Complete"
  success "Workspace ready at ${WORKSPACE}"
  info "Bin directory: ${BIN_DIR}"
  if [[ "${binary_ready}" -eq 0 ]]; then
    warn "chauf binary is not installed; see guidance above."
  fi
  info "Reload your shell (e.g. 'source ~/.zshrc' or open a new terminal) so 'chauf' is available."
}

main() {
  # Check if we're being called via curl (not in git repo and no SCRIPT_DIR context)
  if ! is_git_repo; then
    if is_curl_install; then
      info "Detected curl installation - downloading repository..."
      export CHAUF_CURL_INSTALL=1
    fi
    bootstrap_from_remote "$@"
    return
  fi

  run_local_install "$@"
}

main "$@"
