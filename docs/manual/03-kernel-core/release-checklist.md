# Release Checklist for v0.4.0

This checklist provides a quick way to validate that ring3 execution and memory isolation are working as expected before a release.

## 1. Build the OS

Ensure the ISO is built successfully:

```bash
make clean
make iso
```

## 2. Test Normal User Execution

Run QEMU with the newly built ISO:

```bash
qemu-system-x86_64 -cdrom build/dav-go-os.iso -serial stdio
```

**Steps:**
1. Wait for the `DavOS >` prompt.
2. Run the command: `run hello`
3. **Expected Output:**
   - `hello from userland`
   - `Process exited with status 0`
   - The shell prompt `> ` should return normally.

## 3. Test Isolation (Kernel Read Fault)

**Steps:**
1. In the DavOS shell, run: `run kread`
2. **Expected Output:**
   - A protection fault (`#PF`) should be triggered.
   - A `PF` marker should be visible in the debug log.
   - The QEMU instance should halt (or the kernel should trap and halt as implemented).

## 4. Test Isolation (Kernel Write Fault)

**Steps:**
1. Restart the OS and at the prompt, run: `run kwrite`
2. **Expected Output:**
   - A protection fault (`#PF`) should be triggered.
   - A `PF` marker should be visible in the debug log.
   - The QEMU instance should halt.

## 5. Automated Verification

For automated validation, run the boot test suite:

```bash
python3 scripts/test_boot.py
```

**Expected Output:**
- `All QEMU checks passed.`

If all the above tests pass, ring3 execution and kernel memory isolation are functioning correctly.
