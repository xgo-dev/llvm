# Go bindings to system LLVM

[![Build Status](https://github.com/xgo-dev/llvm/actions/workflows/go.yml/badge.svg)](https://github.com/xgo-dev/llvm/actions/workflows/go.yml)
[![GoDoc](https://pkg.go.dev/badge/github.com/xgo-dev/llvm.svg)](https://pkg.go.dev/github.com/xgo-dev/llvm)
<!--
[![GitHub release](https://img.shields.io/github/v/tag/goplus/llvm.svg?label=release)](https://github.com/xgo-dev/llvm/releases)
[![Coverage Status](https://codecov.io/gh/goplus/llvm/branch/main/graph/badge.svg)](https://codecov.io/gh/goplus/llvm)
-->

This library provides bindings to a system-installed LLVM.

Currently supported:

  * LLVM 22, 21, 20, 19, 18, 17, 16, 15 and 14 from [apt.llvm.org](http://apt.llvm.org/) on Debian/Ubuntu.
  * LLVM 22, 21, 20, 19, 18, 17, 16, 15 and 14 from Homebrew on macOS.
  * LLVM 22, 21, 20, 19, 18, 17, 16, 15 and 14 in an MSYS2 MINGW64 environment on Windows (see the setup below).
  * Any of the above versions with a manually built LLVM through the `byollvm` build tag. You need to set up `CFLAGS`/`LDFLAGS` etc yourself in this case.

You can select the LLVM version using a build tag, for example `-tags=llvm17`
to use LLVM 17.

## Usage

If you have a supported LLVM installation, you should be able to do a simple `go get`:

    go get github.com/xgo-dev/llvm

You can use build tags to select a LLVM version. For example, use `-tags=llvm15` to select LLVM 15. Setting a build tag for a LLVM version that is not supported will be ignored.

### Windows (MSYS2 MINGW64)

The Windows bindings expect `pkg-config` metadata named after the selected LLVM
major version, such as `llvm-19`. After installing a MinGW-compatible LLVM,
`mingw-w64-x86_64-gcc`, and `mingw-w64-x86_64-pkgconf` in a MINGW64
environment, generate that file from `llvm-config`:

    pc_dir="$PWD/.llvm-pkgconfig"
    mkdir -p "$pc_dir"
    llvm_version="$(llvm-config --version)"
    llvm_major="${llvm_version%%.*}"
    case "$llvm_major" in
      14|15|16|17|18|19|20|21|22) ;;
      *) echo "unsupported LLVM version: $llvm_version" >&2; exit 1 ;;
    esac
    cflags="$(llvm-config --cflags | tr '\r\n' '  ')"
    # Normalize the release build machine's zstd path and link winpthread.
    ldflags="$(llvm-config --ldflags --libs all --system-libs | tr '\r\n' '  ' | sed -E 's#[^[:space:]]*/mingw64/lib/libzstd\.dll\.a#-lzstd#g') -lwinpthread"
    printf '%s\n' \
      "Name: LLVM $llvm_major" \
      "Description: LLVM $llvm_major host compiler and linker flags" \
      "Version: $llvm_version" \
      "Cflags: $cflags" \
      "Libs: $ldflags" \
      > "$pc_dir/llvm-$llvm_major.pc"
    export PKG_CONFIG_PATH="$(cygpath -m "$pc_dir")"
    export CC=gcc CXX=g++ CGO_ENABLED=1
    go test -tags="llvm$llvm_major"

The Windows CI job in [`.github/workflows/go.yml`](.github/workflows/go.yml)
shows the exact LLVM versions and SHA-256 checksums used by this project.

## License

These LLVM bindings for Go originally come from LLVM, but they have since been [removed](https://discourse.llvm.org/t/rfc-remove-the-go-bindings/65725). Still, they remain under the same license as they were originally, which is the [Apache License 2.0 (with LLVM exceptions)](http://releases.llvm.org/9.0.0/LICENSE.TXT). Check upstream LLVM for detailed copyright information.

This README, the backports\* files, and the Makefile are separate from LLVM but are licensed under the same license.
