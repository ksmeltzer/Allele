#!/bin/bash
echo "Compiling Completeness Arbitrage Strategy to WASM..."

# Create the hidden strategies directory if it doesn't exist
mkdir -p ../../../.allele/strategies

# Build the WASM file using standard Go compiler
# For highly optimized WASM, tinygo is recommended instead of standard go.
GOOS=wasip1 GOARCH=wasm go build -o ../../../.allele/strategies/completeness.wasm main.go

echo "Done! The WASM plugin is now in .allele/strategies/"
