#!/bin/sh
# Installs only a checksum-verified release, without root. No network at runtime.
set -eu
repo=${KEYWI_REPO:-gaurv-m/keywi}
bindir=${KEYWI_INSTALL_DIR:-"${HOME:?HOME is required}/.local/bin"}
version=${KEYWI_VERSION:-latest}
fail() { printf 'keywi: %s\n' "$*" >&2; exit 1; }
case "$bindir" in /*) ;; *) fail 'KEYWI_INSTALL_DIR must be an absolute path' ;; esac
case ${1:-install} in
  uninstall)
    target="$bindir/keywi"
    if [ -L "$target" ]; then fail 'refusing to uninstall a symbolic link'; fi
    if [ ! -f "$target" ]; then fail "no installation at $target"; fi
    # Validate before removal; never recursively remove user data or directories.
    "$target" --version 2>/dev/null | awk '/^keywi / { found=1 } END { exit !found }' || fail 'target is not a Keywi binary'
    rm -- "$target"
    printf 'Removed %s. Preferences and history were preserved.\n' "$target"
    exit 0 ;;
  install|upgrade) ;;
  *) fail 'usage: sh install.sh [install|upgrade|uninstall]' ;;
esac
case "$repo" in OWNER/*) fail 'No public repository is configured. Set KEYWI_REPO=owner/keywi after releases are published.' ;; esac
# Disallow shell metacharacters, path traversal, URL credentials and query strings.
case "$repo" in *[!a-zA-Z0-9._/-]*|*..*|/*|*/|*/*/*) fail 'invalid KEYWI_REPO (expected owner/repository)' ;; esac
case "$repo" in */*) ;; *) fail 'invalid KEYWI_REPO (expected owner/repository)' ;; esac
case "$version" in latest) release='latest/download' ;; v[0-9]*)
 case "$version" in *[!a-zA-Z0-9._-]*) fail 'invalid KEYWI_VERSION' ;; esac
 release="download/$version" ;; *) fail 'KEYWI_VERSION must be latest or a v-prefixed release tag' ;; esac
case $(uname -s) in Darwin) platform=darwin ;; Linux) platform=linux ;; *) fail 'supported systems: macOS and Linux' ;; esac
case $(uname -m) in arm64|aarch64) arch=arm64 ;; x86_64|amd64) arch=amd64 ;; *) fail 'supported architectures: ARM64 and AMD64' ;; esac
for tool in curl tar awk mktemp; do command -v "$tool" >/dev/null 2>&1 || fail "required tool missing: $tool"; done
if command -v sha256sum >/dev/null 2>&1; then hash=sha256sum
elif command -v shasum >/dev/null 2>&1; then hash=shasum
else fail 'sha256sum or shasum is required'; fi
umask 077
tmp=$(mktemp -d "${TMPDIR:-/tmp}/keywi.XXXXXXXX") || fail 'cannot create temporary directory'
staged=''
cleanup() { rm -rf -- "$tmp"; if [ -n "$staged" ]; then rm -f -- "$staged"; fi; }
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
asset="keywi_${platform}_${arch}.tar.gz"
base="https://github.com/$repo/releases/$release"
fetch() { curl --fail --silent --show-error --location --proto '=https' --proto-redir '=https' --tlsv1.2 --connect-timeout 15 --max-time 180 --retry 2 "$1" -o "$2" || fail "download failed: $1"; }
fetch "$base/$asset" "$tmp/$asset"
fetch "$base/checksums.txt" "$tmp/checksums.txt"
expected=$(awk -v name="$asset" '$2==name {print $1; n++} END {if(n!=1) exit 1}' "$tmp/checksums.txt") || fail 'missing or ambiguous release checksum'
case "$expected" in *[!0-9a-fA-F]*|'') fail 'invalid release checksum' ;; esac
[ ${#expected} -eq 64 ] || fail 'invalid SHA-256 length'
if [ "$hash" = sha256sum ]; then actual=$(sha256sum "$tmp/$asset" | awk '{print $1}')
else actual=$(shasum -a 256 "$tmp/$asset" | awk '{print $1}'); fi
[ "$actual" = "$expected" ] || fail 'checksum mismatch; installation aborted'
# Extract only the single expected member, never archive-controlled paths.
tar -xzf "$tmp/$asset" -C "$tmp" keywi || fail 'invalid release archive'
[ -f "$tmp/keywi" ] && [ ! -L "$tmp/keywi" ] || fail 'archive did not contain a regular binary'
chmod 755 "$tmp/keywi"
"$tmp/keywi" --version | awk '/^keywi / {found=1} END {exit !found}' || fail 'downloaded binary cannot run on this system'
mkdir -p -- "$bindir" || fail "cannot create $bindir"
[ ! -L "$bindir/keywi" ] || fail 'refusing to replace a symbolic link'
# Rename within destination filesystem, so interrupted upgrades keep the old file.
staged=$(mktemp "$bindir/.keywi.XXXXXXXX") || fail 'cannot stage installation'
cp "$tmp/keywi" "$staged"
chmod 755 "$staged"
mv -f -- "$staged" "$bindir/keywi"
staged=''
printf 'Installed %s\n' "$bindir/keywi"
case ":${PATH:-}:" in *":$bindir:"*) exit 0 ;; esac
# Default installs add the command to the user's interactive shell once.
profile=''
if [ "$bindir" = "$HOME/.local/bin" ]; then
  case ${SHELL:-} in
    */zsh|zsh) profile="${ZDOTDIR:-$HOME}/.zshrc" ;;
    */bash|bash) profile="$HOME/.bashrc" ;;
  esac
fi
pathline='export PATH="$HOME/.local/bin:$PATH"'
if [ -n "$profile" ] && [ ! -L "$profile" ]; then
  if [ ! -f "$profile" ] || ! awk -v line="$pathline" '$0==line {found=1} END {exit !found}' "$profile"; then
    printf '\n%s\n' "$pathline" >> "$profile" || fail "cannot update $profile"
  fi
  printf 'Open a new terminal, then run: keywi\n'
else
  printf 'Add to your shell profile:\n  export PATH="%s:$PATH"\n' "$bindir"
fi
