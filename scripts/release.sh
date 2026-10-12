#!/usr/bin/env bash
set -euo pipefail

# scripts/release.sh - automated version bumping, changelog generation, and tagging
# usage: ./scripts/release.sh [patch|minor|major|vX.Y.Z] [--dry-run]

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$REPO_ROOT"

DRY_RUN=false
BUMP_TYPE="patch"

for arg in "$@"; do
  case "$arg" in
    --dry-run)
      DRY_RUN=true
      ;;
    patch|minor|major|v*|[0-9]*)
      BUMP_TYPE="$arg"
      ;;
    -h|--help)
      echo "usage: $0 [patch|minor|major|vX.Y.Z] [--dry-run]"
      exit 0
      ;;
    *)
      echo "error: unknown argument '$arg'"
      exit 1
      ;;
  esac
done

# 1. verify clean working tree
if [ "$DRY_RUN" = false ]; then
  if [ -n "$(git status --porcelain 2>/dev/null || true)" ]; then
    echo "error: working tree is dirty. commit or stash changes before releasing."
    exit 1
  fi
fi

# 2. get latest tag (fallback to v0.0.0 if none found)
LATEST_TAG=$(git tag -l "v*" --sort=-v:refname 2>/dev/null | head -n 1 || echo "")
if [ -z "$LATEST_TAG" ]; then
  LATEST_TAG="v0.0.0"
fi
CLEAN_LATEST="${LATEST_TAG#v}"

IFS='.' read -r MAJOR MINOR PATCH <<< "$CLEAN_LATEST"
MAJOR="${MAJOR:-0}"
MINOR="${MINOR:-0}"
PATCH="${PATCH:-0}"

# 3. determine next version
case "$BUMP_TYPE" in
  patch)
    if [ "$LATEST_TAG" = "v0.0.0" ]; then
      NEXT_VERSION="v0.1.0"
    else
      NEXT_VERSION="v${MAJOR}.${MINOR}.$((PATCH + 1))"
    fi
    ;;
  minor)
    if [ "$LATEST_TAG" = "v0.0.0" ]; then
      NEXT_VERSION="v0.1.0"
    else
      NEXT_VERSION="v${MAJOR}.$((MINOR + 1)).0"
    fi
    ;;
  major)
    if [ "$LATEST_TAG" = "v0.0.0" ]; then
      NEXT_VERSION="v1.0.0"
    else
      NEXT_VERSION="v$((MAJOR + 1)).0.0"
    fi
    ;;
  v*)
    NEXT_VERSION="$BUMP_TYPE"
    ;;
  *)
    NEXT_VERSION="v$BUMP_TYPE"
    ;;
esac

CLEAN_NEXT="${NEXT_VERSION#v}"
RELEASE_DATE=$(date +%Y-%m-%d)

# verify tag doesn't already exist
if git rev-parse -q --verify "refs/tags/${NEXT_VERSION}" >/dev/null 2>&1; then
  echo "error: tag '${NEXT_VERSION}' already exists in git."
  exit 1
fi

echo "========================================"
echo " readable-gcl release automation"
echo "========================================"
echo "current version: ${LATEST_TAG}"
echo "next version:    ${NEXT_VERSION}"
echo "release date:    ${RELEASE_DATE}"
echo "mode:            $([ "$DRY_RUN" = true ] && echo "dry run" || echo "live release")"
echo "========================================"

# 4. verify clean working tree before generating release
if [ "$DRY_RUN" = false ] && [ -n "$(git status --porcelain 2>/dev/null || true)" ]; then
  echo "error: working tree is dirty. commit or stash changes before releasing."
  exit 1
fi

# 5. collect commit logs since latest tag
COMMITS_LOG=""
if git rev-parse --verify HEAD >/dev/null 2>&1; then
  if [ "$LATEST_TAG" = "v0.0.0" ] || ! git rev-parse -q --verify "refs/tags/${LATEST_TAG}" >/dev/null 2>&1; then
    COMMITS_LOG=$(git log --pretty=format:"%s (%h)" 2>/dev/null || echo "")
  else
    COMMITS_LOG=$(git log "${LATEST_TAG}..HEAD" --pretty=format:"%s (%h)" 2>/dev/null || echo "")
  fi
fi

ADDED_ITEMS=""
FIXED_ITEMS=""
CHANGED_ITEMS=""
OTHER_ITEMS=""

if [ -n "$COMMITS_LOG" ]; then
  while IFS= read -r line; do
    [ -z "$line" ] && continue
    case "$line" in
      feat*|Feat*)
        clean_msg=$(echo "$line" | sed -E 's/^[fF]eat(\([^)]+\))?:[[:space:]]*//')
        ADDED_ITEMS="${ADDED_ITEMS}- ${clean_msg}
"
        ;;
      fix*|Fix*)
        clean_msg=$(echo "$line" | sed -E 's/^[fF]ix(\([^)]+\))?:[[:space:]]*//')
        FIXED_ITEMS="${FIXED_ITEMS}- ${clean_msg}
"
        ;;
      refactor*|perf*|docs*|chore*)
        clean_msg=$(echo "$line" | sed -E 's/^[a-zA-Z]+(\([^)]+\))?:[[:space:]]*//')
        CHANGED_ITEMS="${CHANGED_ITEMS}- ${clean_msg}
"
        ;;
      *)
        OTHER_ITEMS="${OTHER_ITEMS}- ${line}
"
        ;;
    esac
  done <<< "$COMMITS_LOG"
fi

# 6. build changelog snippet
NEW_CHANGELOG_ENTRY="## [${CLEAN_NEXT}] - ${RELEASE_DATE}
"

if [ -n "$ADDED_ITEMS" ]; then
  NEW_CHANGELOG_ENTRY="${NEW_CHANGELOG_ENTRY}
### Added

${ADDED_ITEMS}"
fi

if [ -n "$FIXED_ITEMS" ]; then
  NEW_CHANGELOG_ENTRY="${NEW_CHANGELOG_ENTRY}
### Fixed

${FIXED_ITEMS}"
fi

if [ -n "$CHANGED_ITEMS" ]; then
  NEW_CHANGELOG_ENTRY="${NEW_CHANGELOG_ENTRY}
### Changed

${CHANGED_ITEMS}"
fi

if [ -n "$OTHER_ITEMS" ] && [ -z "$ADDED_ITEMS" ] && [ -z "$FIXED_ITEMS" ] && [ -z "$CHANGED_ITEMS" ]; then
  NEW_CHANGELOG_ENTRY="${NEW_CHANGELOG_ENTRY}
### Changes

${OTHER_ITEMS}"
fi

if [ -z "$ADDED_ITEMS" ] && [ -z "$FIXED_ITEMS" ] && [ -z "$CHANGED_ITEMS" ] && [ -z "$OTHER_ITEMS" ]; then
  NEW_CHANGELOG_ENTRY="${NEW_CHANGELOG_ENTRY}
- Maintenance and internal improvements.
"
fi

echo "==> preview changelog entry:"
echo "----------------------------------------"
echo -e "$NEW_CHANGELOG_ENTRY"
echo "----------------------------------------"

if [ "$DRY_RUN" = true ]; then
  echo "dry-run complete. no files changed, no tags created."
  exit 0
fi

# 7. update CHANGELOG.md
entry_file=$(mktemp)
echo -e "$NEW_CHANGELOG_ENTRY" | awk '{lines[NR]=$0} END {while (NR>0 && lines[NR] ~ /^[[:space:]]*$/) NR--; for(i=1;i<=NR;i++) print lines[i]}' > "$entry_file"

if [ -f "CHANGELOG.md" ]; then
  temp_file=$(mktemp)
  awk -v ef="$entry_file" '
    /^## \[/ && !inserted {
      while ((getline line < ef) > 0) {
        print line
      }
      print ""
      inserted = 1
    }
    { print }
    END {
      if (!inserted) {
        while ((getline line < ef) > 0) {
          print line
        }
      }
    }
  ' CHANGELOG.md > "$temp_file"
  mv "$temp_file" CHANGELOG.md
else
  cat <<EOF > CHANGELOG.md
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

EOF
  cat "$entry_file" >> CHANGELOG.md
fi
rm -f "$entry_file"

# ensure exactly 1 empty line at the end of CHANGELOG.md
clean_file=$(mktemp)
awk '{lines[NR]=$0} END {while (NR>0 && lines[NR] ~ /^[[:space:]]*$/) NR--; for(i=1;i<=NR;i++) print lines[i]}' CHANGELOG.md > "$clean_file"
mv "$clean_file" CHANGELOG.md

# 8. commit changelog & create tag with lowercase messages
git add CHANGELOG.md
if ! git diff --cached --quiet; then
  git commit -m "chore: release ${NEXT_VERSION}"
fi
git tag -a "${NEXT_VERSION}" -m "release ${NEXT_VERSION}"

echo "==> release ${NEXT_VERSION} created successfully!"
echo ""
echo "to publish to remote and trigger release workflow:"
echo "  git push origin main && git push origin ${NEXT_VERSION}"
