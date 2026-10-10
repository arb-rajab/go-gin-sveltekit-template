#!/usr/bin/env bash
# Rename this template to your project name.
#
#   ./rename.sh <slug> [PascalName]
#
#   slug        lowercase letters and digits only, starting with a letter
#               (it becomes a Python package / Go module / Java package /
#               npm name, so no hyphens or underscores). Example: acmeportal
#   PascalName  optional; defaults to the slug with its first letter
#               upper-cased (acmeportal -> Acmeportal). Example: AcmePortal
#
# Replaces the placeholder tokens `starterapp`, `StarterApp` and `STARTERAPP`
# in every text file, then renames files and directories whose names contain
# `starterapp` or `StarterApp`. Set GO_OWNER=<github-user-or-org> to also
# replace `github.com/example/` in the Go module paths. Run it once, on a clean checkout, before your
# first commit; delete this script afterwards.
set -euo pipefail

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  echo "usage: $0 <slug> [PascalName]" >&2
  exit 2
fi

SLUG="$1"
if ! [[ "$SLUG" =~ ^[a-z][a-z0-9]*$ ]]; then
  echo "error: slug must match ^[a-z][a-z0-9]*$ (got '$SLUG')" >&2
  exit 2
fi
if [ -n "${2:-}" ]; then
  PASCAL="$2"
else
  PASCAL="$(printf '%s' "${SLUG:0:1}" | tr '[:lower:]' '[:upper:]')${SLUG:1}"
fi
if ! [[ "$PASCAL" =~ ^[A-Z][A-Za-z0-9]*$ ]]; then
  echo "error: PascalName must match ^[A-Z][A-Za-z0-9]*$ (got '$PASCAL')" >&2
  exit 2
fi
UPPER="$(printf '%s' "$SLUG" | tr '[:lower:]' '[:upper:]')"

# Tokens are assembled at run time so this script never rewrites itself.
T_SLUG="starter""app"
T_PASCAL="Starter""App"
T_UPPER="STARTER""APP"

cd "$(dirname "$0")"
SELF="$(basename "$0")"

EXCLUDES=(--exclude-dir=.git --exclude-dir=node_modules --exclude-dir=target
          --exclude-dir=.next --exclude-dir=.svelte-kit --exclude-dir=build --exclude-dir=dist --exclude="$SELF")

if [ -z "$(grep -rIl "${EXCLUDES[@]}" -e "$T_SLUG" -e "$T_PASCAL" -e "$T_UPPER" . || true)" ]; then
  echo "Nothing to rename: the placeholder tokens are not present." >&2
  exit 1
fi

# 1. Replace tokens inside text files (binary files are skipped by -I).
grep -rIl "${EXCLUDES[@]}" -e "$T_SLUG" -e "$T_PASCAL" -e "$T_UPPER" . |
while IFS= read -r file; do
  perl -pi -e "s/\Q$T_PASCAL\E/$PASCAL/g; s/\Q$T_UPPER\E/$UPPER/g; s/\Q$T_SLUG\E/$SLUG/g" "$file"
done

# 1b. Optional: point Go module paths at your own GitHub account or org.
if [ -n "${GO_OWNER:-}" ]; then
  grep -rIl "${EXCLUDES[@]}" -e "github.com/example/" . |
  while IFS= read -r file; do
    perl -pi -e "s{github\.com/example/}{github.com/$GO_OWNER/}g" "$file"
  done
fi

# 2. Rename files and directories, deepest paths first.
find . -depth \( -name "*$T_SLUG*" -o -name "*$T_PASCAL*" \) \
     -not -path './.git/*' -not -path '*/node_modules/*' |
while IFS= read -r path; do
  dir="$(dirname "$path")"
  base="$(basename "$path")"
  new="${base//$T_PASCAL/$PASCAL}"
  new="${new//$T_SLUG/$SLUG}"
  [ "$base" = "$new" ] || mv "$path" "$dir/$new"
done

echo "Renamed to '$SLUG' ($PASCAL). Review 'git diff', then delete $SELF."
