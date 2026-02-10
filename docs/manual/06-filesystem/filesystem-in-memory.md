# In-Memory Filesystem

## What it does

Provides a tiny volatile filesystem in RAM for quick testing and debugging.

Key file:

- `fs/fs.go`

## Data model

- maximum 32 files
- maximum 16-byte file name
- each file uses exactly one 4096-byte page
- metadata table is static array-based

## Operations

- `Write` creates or overwrites a file
- `Lookup` finds a file by name
- `Entry` exposes metadata for listing
- `Remove` deletes a file and frees its page in PFA

## Shell commands

- `ls`
- `write <name> <text...>`
- `cat <name>`
- `rm <name>`
- `stat <name>`

## Important notes

- non-persistent: data is lost on reboot
- no real directories
- no journaling or permissions
