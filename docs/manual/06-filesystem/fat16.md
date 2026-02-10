# Persistent FAT16

## What it does

Implements a minimal FAT16 layer on top of ATA driver, with persistence on disk image.

Key file:

- `fs/fat16/fat16.go`

## Main functions

- `Format()` writes basic BPB/FAT/root structures
- `Init()` reads BPB and computes offsets
- `Info()` prints layout
- `ListDir()` lists root directory entries
- `CreateFile()` creates a root file (single cluster path)
- `ReadFile()` reads a root file

## Shell commands

- `fatformat`
- `fatinit`
- `fatinfo`
- `fatls`
- `fatcreate <filename> <content>`
- `fatread <filename>`

## Current limitations

- root directory only
- simplified 8.3 naming behavior
- practical file size path currently centered on single-sector read/write logic
- no full multi-cluster chain management in current command flow

## Suggested sequence

1. `fatformat`
2. `fatinit`
3. `fatcreate HELLO text`
4. `fatls`
5. `fatread HELLO`
