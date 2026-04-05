# Examples and Demo Scripts

This directory contains shell scripts to help you quickly build, run, and explore **go-dav-os**.

## 🚀 Building and Running

### 1. Build and Run locally
If you have the `x86_64-elf` toolchain and `qemu` installed:
```bash
./examples/build_and_run.sh
```

### 2. Run inside Docker
If you want to use the provided Docker toolchain:
```bash
./examples/run_docker.sh
```

## 💾 FAT16 File System Demo

Once the OS has booted and you see the `dav-go-os> ` prompt, you can try these commands to explore the FAT16 implementation:

1. **Initialize the disk:**
   ```bash
   fatinit
   ```
2. **Format the disk:**
   ```bash
   fatformat
   ```
3. **Create a file:**
   ```bash
   fatcreate welcome.txt Hello_from_the_other_side
   ```
4. **List files:**
   ```bash
   fatls
   ```
5. **Read the file:**
   ```bash
   fatread welcome.txt
   ```
6. **Show system info:**
   ```bash
   version
   uptime
   mem
   ```

## 🛠️ Script Details

- `build_and_run.sh`: Performs a clean build and launches QEMU.
- `run_docker.sh`: Uses the Docker toolchain to build and run the OS.
- `test_all.sh`: Runs all unit tests.
