#!/bin/bash

set -ex

DIR=`realpath $(dirname "${0}")`
PRIME_DIRECTORY="${CRAFT_PRIME:-${DIR}/../../.prime}"

BIN="${1}"
ARCH_TRIPLET="${2}"

# Note(ben): patchelf messes up the segments when patching the binary
# due to https://github.com/NixOS/patchelf/issues/446.
# Instead, we use a custom script based on LIEF to manually patch the rpath and interpreter.
#
# RPATH and interpreter use absolute in-snap paths (no /snap/<base>/current/...
# prefix). Under strict confinement the base snap is mounted at /, so these
# resolve to the base's libs/loader; the explicit /snap/core*/... form is
# rejected by AppArmor and was the source of the pre-refresh-hook DENIED logs.
echo "Patching ELF file: $BIN with LIEF"

# Install Python deps for LIEF patcher
# Due to LIEF only being available as binary wheels for
# amd64 and arm64, only support these architectures.
pip3 install -r $DIR/requirements.txt

if [ "$ARCH_TRIPLET" = "x86_64-linux-gnu" ]; then
    python3 $DIR/patchelf.py "$BIN" \
        --set-rpath /lib/x86_64-linux-gnu/ \
        --set-interpreter /lib64/ld-linux-x86-64.so.2
elif [ "$ARCH_TRIPLET" = "aarch64-linux-gnu" ]; then
    python3 $DIR/patchelf.py "$BIN" \
        --set-rpath /lib/aarch64-linux-gnu/ \
        --set-interpreter /lib/ld-linux-aarch64.so.1
else
    echo "Unsupported Architecture: $ARCH_TRIPLET"
    exit 1
fi

echo "==> LIEF patching $BIN complete"
mkdir -p "$PRIME_DIRECTORY/bin"
cp "$BIN" "$PRIME_DIRECTORY/bin/"
