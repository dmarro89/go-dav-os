# Tests and Debugging

## Go unit tests

Examples included:

- `kernel/scheduler/scheduler_test.go`
- `fs/fs_test.go`

Run:

```bash
go test ./...
```

Note: these tests focus on local logic and do not validate full kernel boot.

## QEMU boot smoke test

Script:

- `scripts/test_boot.py`

What it checks:

1. boot and presence of "Welcome to DavOS"
2. `help` command output
3. `version` command output including `(64bit)` marker

Example:

```bash
python3 scripts/test_boot.py
```

## Low-level debug output

Terminal/stubs also write to debug port `0xE9`.
Use QEMU `-debugcon file:qemu.log` to capture early output.

This is useful when VGA output is unstable or when debugging very early boot stages.
