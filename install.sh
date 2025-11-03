#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE="${HOME}/.chauffeur"
BIN_DIR="${WORKSPACE}/bin"
SHIMS_DIR="${BIN_DIR}/shims"
PATH_LINE='export PATH="$HOME/.chauffeur/bin:$PATH"'
SHELL_NAME="$(basename "${SHELL:-}")"
LOCAL_SRC_DIR="${SCRIPT_DIR}/cli"
REPO_URL="${CHAUF_REPO_URL:-https://github.com/SIAJI-Labs/chauffeur.git}"

section() {
  printf '\n[ INSTALL ] %s\n' "$(echo "$1" | tr '[:lower:]' '[:upper:]')"
}

info() {
  printf '    - %s\n' "$*"
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
    "${WORKSPACE}/caddy/bin"
  )

  for dir in "${dirs[@]}"; do
    mkdir -p "${dir}"
  done
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
    printf '\n%s\n' "${PATH_LINE}" >>"${rc_file}"
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
  if (cd "${SCRIPT_DIR}" && GO111MODULE=on go build -o "${tmp_binary}" ./cli); then
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

is_git_repo() {
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
    exit 0
  else
    local status=$?
    rm -rf "${tmp_dir}"
    exit "${status}"
  fi
}

run_local_install() {
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
  if ! is_git_repo; then
    bootstrap_from_remote "$@"
    return
  fi

  run_local_install "$@"
}

main "$@"
