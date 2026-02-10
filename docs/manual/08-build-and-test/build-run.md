# Build and Run

## Reference files

- `Makefile`
- `Dockerfile`
- `iso/grub/grub.cfg`

## Docker build (recommended)

```bash
docker build --platform=linux/amd64 -t go-dav-os-toolchain .
docker run --rm --platform=linux/amd64 -v "$PWD":/work -w /work go-dav-os-toolchain make
qemu-system-x86_64 -cdrom build/dav-go-os.iso
```

## Native build

Requires cross toolchain `x86_64-elf-*` and GRUB/QEMU tools.

```bash
make
qemu-system-x86_64 -cdrom build/dav-go-os.iso
```

## Main artifacts

- `build/kernel.elf`
- `build/dav-go-os.iso`
- `disk.img` (when using `make run`)

## Useful Make targets

- `make` -> full ISO build
- `make run` -> boot QEMU with ISO + `disk.img` (20 MB)
- `make clean` -> remove build output
