#!/usr/bin/env bash
# Chauffeur installation script
# Usage: curl -fsSL https://raw.githubusercontent.com/siegg/chauffeur/main/scripts/install.sh | bash
#
# Environment variables:
#   CHAUFFEUR_VERSION   — specific version to install (default: latest release)
#   CHAUFFEUR_HOME      — workspace root (default: ~/.chauffeur)

set -euo pipefail

REPO="siegg/chauffeur"
BIN="chauf"
CHAUF_HOME="${CHAUFFEUR_HOME:-$HOME/.chauffeur}"
INSTALL_DIR="${CHAUF_HOME}/bin"

# ── output helpers ─────────────────────────────────────────────────────────────

bold=$'\033[1m'
green=$'\033[32m'
red=$'\033[31m'
gray=$'\033[90m'
reset=$'\033[0m'

info()    { printf "  %s\n" "$*"; }
success() { printf "  ${green}✓${reset}  %s\n" "$*"; }
fail()    { printf "  ${red}✗${reset}  %s\n" "$*" >&2; }
die()     { fail "$*"; exit 1; }

# ── platform detection ─────────────────────────────────────────────────────────

detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux)  os="linux" ;;
        Darwin) die "macOS is not yet supported. Build from source: go build ./cmd/chauf" ;;
        *)      die "Unsupported OS: $(uname -s)" ;;
    esac

    case "$(uname -m)" in
        x86_64)  arch="amd64" ;;
        aarch64) arch="arm64" ;;
        armv7l)  arch="arm"   ;;
        *)       die "Unsupported architecture: $(uname -m)" ;;
    esac

    echo "${os}_${arch}"
}

# ── repo detection ─────────────────────────────────────────────────────────────

in_chauffeur_repo() {
    [[ -f "go.mod" ]] && grep -q "chauffeur" go.mod 2>/dev/null
}

# ── build from source ──────────────────────────────────────────────────────────

build_from_source() {
    local srcdir="$1"

    command -v go &>/dev/null \
        || die "Go is required to build from source. Install from https://go.dev/dl/"

    local goversion
    goversion=$(go version | awk '{print $3}' | sed 's/go//')
    info "Go ${goversion} found"

    mkdir -p "$INSTALL_DIR"
    info "Building chauf..."
    (cd "$srcdir" && go build -o "${INSTALL_DIR}/${BIN}" ./cmd/chauf) \
        || die "Build failed"
}

# ── clone and build ────────────────────────────────────────────────────────────

clone_and_build() {
    command -v git &>/dev/null || die "git is required to clone the repository"
    command -v go  &>/dev/null || die "Go is required to build from source. Install from https://go.dev/dl/"

    local tmpdir
    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT

    info "Cloning  https://github.com/${REPO}"
    git clone --depth=1 "https://github.com/${REPO}" "${tmpdir}/chauffeur" 2>/dev/null \
        || die "Clone failed. Check your connection or try again."
    success "Repository cloned"

    build_from_source "${tmpdir}/chauffeur"
}

# ── download release binary ────────────────────────────────────────────────────

resolve_latest_version() {
    curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null \
        | grep '"tag_name"' \
        | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/'
}

install_release() {
    local version="$1" platform="$2"

    local filename="${BIN}_${version}_${platform}.tar.gz"
    local base_url="https://github.com/${REPO}/releases/download/${version}"
    local url="${base_url}/${filename}"
    local checksums_url="${base_url}/${BIN}_${version}_checksums.txt"

    local tmpdir
    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT

    info "Downloading  ${BIN} ${version} (${platform})"
    curl -fsSL "$url" -o "${tmpdir}/${filename}" \
        || die "Download failed: ${url}"

    # Verify checksum from combined checksums.txt
    if curl -fsSL "$checksums_url" -o "${tmpdir}/checksums.txt" 2>/dev/null; then
        local expected
        expected=$(grep " ${filename}$\| \*${filename}$" "${tmpdir}/checksums.txt" | awk '{print $1}')
        if [[ -n "$expected" ]]; then
            info "Verifying checksum"
            echo "${expected}  ${tmpdir}/${filename}" | sha256sum -c --quiet \
                || die "Checksum verification failed"
            success "Checksum verified"
        fi
    fi

    mkdir -p "$INSTALL_DIR"
    tar -xzf "${tmpdir}/${filename}" -C "$tmpdir"
    install -m755 "${tmpdir}/${BIN}" "${INSTALL_DIR}/${BIN}"
}

# ── shell PATH injection ───────────────────────────────────────────────────────

SHELL_BLOCK_OPEN="# >>> chauffeur <<<"
SHELL_BLOCK_CLOSE="# <<< chauffeur <<<"
SHELL_PATH_LINE='export PATH="$HOME/.chauffeur/bin:$HOME/.chauffeur/bin/shims:$PATH"'

detect_shell_config() {
    local shell="${SHELL:-}"
    local home="$HOME"

    if [[ "$shell" == */zsh ]];  then echo "$home/.zshrc";  return; fi
    if [[ "$shell" == */bash ]]; then echo "$home/.bashrc"; return; fi

    for f in "$home/.zshrc" "$home/.bashrc" "$home/.profile"; do
        [[ -f "$f" ]] && echo "$f" && return
    done
}

inject_shell_path() {
    local cfg
    cfg=$(detect_shell_config)
    [[ -z "$cfg" ]] && return

    # Already configured — skip.
    grep -qF "$SHELL_BLOCK_OPEN" "$cfg" 2>/dev/null && return

    local block
    block="$SHELL_BLOCK_OPEN"$'\n'"$SHELL_PATH_LINE"$'\n'"$SHELL_BLOCK_CLOSE"

    # Insert after a bare "# Chauffeur" comment if present, else append.
    if grep -qxF "# Chauffeur" "$cfg" 2>/dev/null; then
        # Use awk to insert the block right after the "# Chauffeur" line.
        awk -v block="$block" '
            /^# Chauffeur$/ { print; print block; next }
            { print }
        ' "$cfg" > "${cfg}.tmp" && mv "${cfg}.tmp" "$cfg"
    else
        {
            echo ""
            echo "# Chauffeur"
            echo "$block"
        } >> "$cfg"
    fi

    local display="${cfg/$HOME/\~}"
    success "PATH added to ${display}  (run: source ${display})"
}

# ── main ───────────────────────────────────────────────────────────────────────

main() {
    echo
    printf "  ${bold}Chauffeur Installer${reset}\n"
    echo

    if in_chauffeur_repo; then
        # ── Running from inside the source repo ──
        info "Source repo detected  ($(pwd))"
        build_from_source "."

    else
        # ── Running externally (curl | bash or standalone) ──
        command -v curl &>/dev/null || die "curl is required"

        local platform version
        platform=$(detect_platform)
        local requested_version="${CHAUFFEUR_VERSION:-latest}"

        if [[ "$requested_version" != "latest" ]]; then
            install_release "$requested_version" "$platform"
        else
            version=$(resolve_latest_version)

            if [[ -n "$version" ]]; then
                install_release "$version" "$platform"
            else
                info "No releases found. Building from source..."
                echo
                clone_and_build
            fi
        fi
    fi

    success "${BIN} installed  →  ${INSTALL_DIR}/${BIN}"
    inject_shell_path

    # Run init directly via full path — no need for PATH to be active yet.
    echo
    "${INSTALL_DIR}/${BIN}" init

    # Tell the user how to activate the PATH in the current session.
    local cfg
    cfg=$(detect_shell_config)
    if [[ -n "$cfg" ]]; then
        local display="${cfg/$HOME/\~}"
        printf "  ${bold}Activate in this terminal:${reset}\n"
        printf "    source %s\n" "$display"
        echo
    fi
}

main "$@"
