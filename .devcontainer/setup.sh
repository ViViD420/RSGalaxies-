#!/usr/bin/env bash
set -euo pipefail

GODOT_VERSION="4.3"
GODOT_RELEASE="4.3-stable"
GODOT_DIR="/tmp/rs-galaxies-godot"
TEMPLATE_DIR="$HOME/.local/share/godot/export_templates/${GODOT_VERSION}.stable"

mkdir -p "$GODOT_DIR" "$TEMPLATE_DIR"
curl -fsSL "https://github.com/godotengine/godot/releases/download/${GODOT_RELEASE}/Godot_v${GODOT_RELEASE}_linux.x86_64.zip" -o "$GODOT_DIR/godot.zip"
unzip -oq "$GODOT_DIR/godot.zip" -d "$GODOT_DIR"
sudo install -m 0755 "$GODOT_DIR/Godot_v${GODOT_RELEASE}_linux.x86_64" /usr/local/bin/godot

curl -fsSL "https://github.com/godotengine/godot/releases/download/${GODOT_RELEASE}/Godot_v${GODOT_RELEASE}_export_templates.tpz" -o "$GODOT_DIR/templates.tpz"
mkdir -p "$GODOT_DIR/templates"
unzip -oq "$GODOT_DIR/templates.tpz" -d "$GODOT_DIR/templates"
cp -a "$GODOT_DIR/templates/templates/." "$TEMPLATE_DIR/"

cd server
go mod download
go test ./...
go vet ./...
cd ../client
godot --headless --editor --path . --quit
