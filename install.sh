#!/usr/bin/env bash
# n0man Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/noeltz/n0man/main/install.sh | bash
#        or with checksum verification:
#        CHECKSUM=<sha256sum> curl -sSL https://raw.githubusercontent.com/noeltz/n0man/main/install.sh | bash
#
# This script installs n0man to $HOME/.local/bin/n0man
# It requires Go, git, and curl/wget to be installed

set -euo pipefail

# ============================================================================
# Bash Version Check
# ============================================================================

check_bash_version() {
    # Check for bash 4.0+
    if [ -z "${BASH_VERSINFO[0]:-}" ]; then
        log_error "This script must be run with bash"
        exit $EXIT_MISSING_BASH
    fi
    
    local bash_major="${BASH_VERSINFO[0]}"
    if [ "${bash_major}" -lt 4 ]; then
        log_error "Bash version 4.0 or higher is required (found ${BASH_VERSION})"
        exit $EXIT_MISSING_BASH
    fi
    
    log_info "Bash version: ${BASH_VERSION}"
}

# ============================================================================
# Checksum Verification
# ============================================================================

verify_checksum() {
    local expected_checksum="$1"
    
    if [ -z "${expected_checksum}" ]; then
        log_warn "No checksum provided, skipping verification"
        return 0
    fi
    
    log_info "Verifying script integrity..."
    
    # Calculate SHA256 checksum of this script
    local actual_checksum
    actual_checksum=$(sha256sum "$0" 2>/dev/null | awk '{print $1}')
    
    if [ -z "${actual_checksum}" ]; then
        log_error "Failed to calculate checksum"
        exit $EXIT_VERIFICATION_FAILED
    fi
    
    if [ "${actual_checksum}" != "${expected_checksum}" ]; then
        log_error "Checksum verification failed!"
        log_error "Expected: ${expected_checksum}"
        log_error "Actual:   ${actual_checksum}"
        log_error "The script may have been tampered with or corrupted"
        exit $EXIT_VERIFICATION_FAILED
    fi
    
    log_success "Checksum verified successfully"
}

# ============================================================================
# Configuration and Constants
# ============================================================================

# Exit codes
EXIT_SUCCESS=0
EXIT_MISSING_BASH=2
EXIT_MISSING_DOWNLOAD_TOOL=3
EXIT_MISSING_GIT=4
EXIT_MISSING_GO=5
EXIT_GO_VERSION_TOO_OLD=6
EXIT_LOCK_FAILED=7
EXIT_CLONE_FAILED=8
EXIT_BUILD_FAILED=9
EXIT_PERMISSION_DENIED=10
EXIT_VERIFICATION_FAILED=11

INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="n0man"
INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}"

REPO_URL="https://github.com/noeltz/n0man.git"
REPO_BRANCH="main"

# Temporary directory for build
TEMP_DIR=$(mktemp -d)
BUILD_DIR="${TEMP_DIR}/n0man"

# Lock file to prevent concurrent installations
# Use XDG_RUNTIME_DIR if available, otherwise use /tmp
if [ -n "${XDG_RUNTIME_DIR:-}" ]; then
    LOCK_FILE="${XDG_RUNTIME_DIR}/n0man-install.lock"
else
    LOCK_FILE="/tmp/n0man-install.lock"
fi

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[0;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# ============================================================================
# Utility Functions
# ============================================================================

log_info() {
    echo -e "${BLUE}ℹ${NC} $*"
}

log_success() {
    echo -e "${GREEN}✓${NC} $*"
}

log_error() {
    echo -e "${RED}✗${NC} $*" >&2
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $*"
}

cleanup() {
    log_info "Cleaning up temporary files..."
    rm -rf "${TEMP_DIR}"
    rm -f "${LOCK_FILE}"
}

# Set trap for cleanup on exit
trap cleanup EXIT

# ============================================================================
# Prerequisite Checks
# ============================================================================

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check for bash
    if ! command -v bash &> /dev/null; then
        log_error "bash is required but not installed"
        exit $EXIT_MISSING_BASH
    fi
    
    # Check for curl or wget
    if ! command -v curl &> /dev/null && ! command -v wget &> /dev/null; then
        log_error "curl or wget is required to download files"
        exit $EXIT_MISSING_DOWNLOAD_TOOL
    fi
    
    # Check for git
    if ! command -v git &> /dev/null; then
        log_error "git is required but not installed"
        log_info "Install git: sudo apt install git (Ubuntu/Debian)"
        log_info "              brew install git (macOS)"
        exit $EXIT_MISSING_GIT
    fi
    
    # Check for Go
    if ! command -v go &> /dev/null; then
        log_error "Go is required but not installed"
        log_info "Install Go: https://go.dev/dl/"
        exit $EXIT_MISSING_GO
    fi
    
    # Verify Go version (requires 1.22+)
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Found Go version: ${GO_VERSION}"
    
    # Check Go version (basic check for 1.22+)
    GO_MAJOR=$(echo "${GO_VERSION}" | cut -d. -f1)
    GO_MINOR=$(echo "${GO_VERSION}" | cut -d. -f2)
    if [ "${GO_MAJOR}" -lt 1 ] || ([ "${GO_MAJOR}" -eq 1 ] && [ "${GO_MINOR}" -lt 22 ]); then
        log_error "Go version 1.22 or higher is required (found ${GO_VERSION})"
        exit $EXIT_GO_VERSION_TOO_OLD
    fi
    
    log_success "All prerequisites satisfied"
}

# ============================================================================
# Installation Logic
# ============================================================================

acquire_lock() {
    if [ -f "${LOCK_FILE}" ]; then
        # Check lock file age
        if [[ "$OSTYPE" == "darwin"* ]]; then
            # macOS
            lock_age=$(($(date +%s) - $(stat -f %m "${LOCK_FILE}" 2>/dev/null || echo 0)))
        else
            # Linux
            lock_age=$(($(date +%s) - $(stat -c %Y "${LOCK_FILE}" 2>/dev/null || echo 0)))
        fi
        
        if [ "${lock_age}" -lt 3600 ]; then
            # Lock is less than 1 hour old
            log_error "Installation already in progress (lock file exists)"
            log_info "If this is incorrect, remove: ${LOCK_FILE}"
            exit $EXIT_LOCK_FAILED
        fi
        log_warn "Removing stale lock file..."
        rm -f "${LOCK_FILE}"
    fi
    
    echo $$ > "${LOCK_FILE}"
}

create_install_directory() {
    log_info "Creating installation directory: ${INSTALL_DIR}"
    
    if [ ! -d "${INSTALL_DIR}" ]; then
        mkdir -p "${INSTALL_DIR}"
        log_success "Created ${INSTALL_DIR}"
    else
        log_info "Directory already exists: ${INSTALL_DIR}"
    fi
}

clone_repository() {
    log_info "Cloning repository from ${REPO_URL}..."
    
    if ! git clone --depth 1 --branch "${REPO_BRANCH}" "${REPO_URL}" "${BUILD_DIR}"; then
        log_error "Failed to clone repository"
        exit $EXIT_CLONE_FAILED
    fi
    
    log_success "Repository cloned successfully"
    
    # Verify repository integrity by checking commit signature if available
    log_info "Verifying repository integrity..."
    if ! git -C "${BUILD_DIR}" verify-commit HEAD 2>/dev/null; then
        log_warn "Commit signature verification failed or no signature found"
        log_warn "Proceeding anyway, but be aware the repository may not be signed"
    else
        log_success "Repository integrity verified"
    fi
}

build_binary() {
    log_info "Building n0man binary..."
    
    cd "${BUILD_DIR}"
    
    # Build with version information
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    if ! go build \
        -ldflags="-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}" \
        -o "${INSTALL_PATH}" \
        ./cmd/n0man; then
        log_error "Failed to build binary"
        exit $EXIT_BUILD_FAILED
    fi
    
    log_success "Binary built successfully"
}

set_permissions() {
    log_info "Setting executable permissions..."
    
    if ! chmod 0755 "${INSTALL_PATH}"; then
        log_error "Failed to set executable permissions"
        exit $EXIT_PERMISSION_DENIED
    fi
    
    log_success "Permissions set successfully"
}

verify_installation() {
    log_info "Verifying installation..."
    
    if [ ! -f "${INSTALL_PATH}" ]; then
        log_error "Binary not found at ${INSTALL_PATH}"
        exit $EXIT_VERIFICATION_FAILED
    fi
    
    if ! "${INSTALL_PATH}" --help &> /dev/null; then
        log_error "Binary exists but is not functional"
        exit $EXIT_VERIFICATION_FAILED
    fi
    
    log_success "Installation verified successfully"
}

check_path() {
    if [[ ":${PATH}:" != *":${INSTALL_DIR}:"* ]]; then
        log_warn "${INSTALL_DIR} is not in your PATH"
        log_info "Add the following to your shell configuration:"
        echo ""
        echo "  export PATH=\"\$PATH:\$HOME/.local/bin\""
        echo ""
        log_info "Then restart your shell or run: source ~/.bashrc"
    else
        log_success "${INSTALL_DIR} is already in PATH"
    fi
}

# ============================================================================
# Main Installation Flow
# ============================================================================

main() {
    echo "🚀 Installing n0man..."
    echo ""
    
    # Verify checksum if provided
    verify_checksum "${CHECKSUM:-}"
    
    check_bash_version
    acquire_lock
    check_prerequisites
    create_install_directory
    clone_repository
    build_binary
    set_permissions
    verify_installation
    check_path
    
    echo ""
    log_success "n0man installed successfully!"
    echo ""
    log_info "Binary location: ${INSTALL_PATH}"
    log_info "Run 'n0man --help' to get started"
    echo ""
}

main "$@"
