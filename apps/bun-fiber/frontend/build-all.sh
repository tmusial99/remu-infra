#!/bin/sh
set -e

SITES_DIR="./src/sites"

if [ ! -d "$SITES_DIR" ]; then
  echo "Directory $SITES_DIR does not exist!"
  exit 1
fi

for site in "$SITES_DIR"/*; do
  if [ -d "$site" ]; then
    site_name=$(basename "$site")
    echo "Building for SITE=$site_name"
    SITE=$site_name bun astro build
  fi
done

echo "All builds completed."