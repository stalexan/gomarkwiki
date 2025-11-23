#!/bin/bash

# Exit on error
set -euo pipefail

SOURCE_DIR="$(cd "$(dirname "$0")/" && pwd)"
TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
DEST_FILE="$HOME/tmp/for-claude/gmw-${TIMESTAMP}.txt"
COMMON_PATH=$SOURCE_DIR

# Copy files
"$HOME/bin/sc-copy-files-for-claude.py" "$SOURCE_DIR" "$DEST_FILE" "$COMMON_PATH"

# Remove past versions of DEST_FILE (excluding the one we just created)
find "$HOME/tmp/for-claude" -name "gmw-*.txt" -not -name "$(basename $DEST_FILE)" -delete 2>/dev/null || true

echo "Files copied to $DEST_FILE"
