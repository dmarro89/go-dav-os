# ATA PIO Driver

## What it does

Reads and writes 512-byte sectors using ATA PIO mode and port I/O.

Key files:

- `drivers/ata/ata.go`
- assembly I/O stubs in `boot/stubs_amd64.s`

## Main I/O ports

- `0x1F0` data
- `0x1F2` sector count
- `0x1F3..0x1F5` LBA low/mid/high
- `0x1F6` drive/head
- `0x1F7` status/command

## Operations

- `ReadSector(lba, buf)`
- `WriteSector(lba, data)`

Both wait for ready state (`waitBusy`, `waitDRQ`) with timeout.

## Related shell commands

- `disk read <lba>`
- `disk write <lba> <text>`

Useful for low-level disk debugging before FAT16 operations.
