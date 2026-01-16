FROM debian:12-slim

ARG BINUTILS_VERSION=2.43
ARG GCC_VERSION=15.2.0

ENV DEBIAN_FRONTEND=noninteractive
ENV PREFIX=/opt/cross
ENV TARGET=x86_64-elf
ENV PATH=$PREFIX/bin:$PATH

# 1. Install base dependencies for building binutils/gcc and OSDev tooling
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      build-essential \
      bison \
      flex \
      wget \
      curl \
      xz-utils \
      ca-certificates \
      libgmp3-dev \
      libmpfr-dev \
      libmpc-dev \
      texinfo \
      grub-pc-bin \
      xorriso \
      mtools \
      qemu-system-x86 && \
    rm -rf /var/lib/apt/lists/*

# 2. Build binutils cross toolchain (x86_64-elf-*)
RUN mkdir -p /src && cd /src && \
    wget https://ftp.gnu.org/gnu/binutils/binutils-${BINUTILS_VERSION}.tar.xz && \
    tar xf binutils-${BINUTILS_VERSION}.tar.xz && \
    mkdir build-binutils && cd build-binutils && \
    ../binutils-${BINUTILS_VERSION}/configure \
        --target=${TARGET} \
        --prefix="${PREFIX}" \
        --with-sysroot \
        --disable-nls \
        --disable-werror && \
    make -j"$(nproc)" && \
    make install && \
    cd / && rm -rf /src

# 3. Build GCC cross toolchain (x86_64-elf-gcc and x86_64-elf-gccgo)
#    We enable C, C++ and Go frontends, but use it in freestanding mode (no headers).
RUN mkdir -p /src && cd /src && \
    wget https://ftp.gnu.org/gnu/gcc/gcc-${GCC_VERSION}/gcc-${GCC_VERSION}.tar.xz && \
    tar xf gcc-${GCC_VERSION}.tar.xz && \
    mkdir build-gcc && cd build-gcc && \
    ../gcc-${GCC_VERSION}/configure \
        --target=${TARGET} \
        --prefix="${PREFIX}" \
        --disable-nls \
        --enable-languages=c,c++,go \
        --without-headers \
        --disable-libsanitizer && \
    make all-gcc -j"$(nproc)" && \
    make all-target-libgcc -j"$(nproc)" && \
    make install-gcc && \
    make install-target-libgcc && \
    cd / && rm -rf /src

# Ensure login shells keep the cross toolchain in PATH.
RUN printf '%s\n' 'export PATH=/opt/cross/bin:$PATH' > /etc/profile.d/cross.sh

# 4. Default working directory for the OS project
WORKDIR /work

# 5. Default command: just drop into a shell
CMD ["bash"]
