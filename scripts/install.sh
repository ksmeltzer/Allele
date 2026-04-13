#!/usr/bin/env bash

# Setup /etc/hosts for allele
echo "Adding allele to /etc/hosts (requires sudo)..."
sudo sh -c 'grep -q "allele" /etc/hosts || echo "127.0.0.1 allele" >> /etc/hosts'

echo "Building watchdog..."
mkdir -p .allele
go build -o .allele/watchdog ./cmd/watchdog
if [ $? -ne 0 ]; then
    echo "Error: Failed to build watchdog daemon."
    exit 1
fi

echo "Building initial container image..."
podman-compose up -d --build

echo "Installing allele CLI..."
mkdir -p "$HOME/.local/bin"

# Copy and substitute PROJECT_ROOT in scripts/allele
sed "s|<PROJECT_ROOT>|$(pwd)|g" scripts/allele > "$HOME/.local/bin/allele"
chmod +x "$HOME/.local/bin/allele"

echo "Installation complete!"
echo ""
echo "Run 'allele' to launch the Arbitrage Dashboard."
