#!/bin/bash
# examples/build_and_run.sh
# Build the ISO and boot with QEMU using the local toolchain.

# Clean previous builds
echo "Cleaning old build artifacts..."
make clean

# Build the ISO
echo "Building the ISO..."
make iso

# Run with QEMU
echo "Launching QEMU..."
make run
