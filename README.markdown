# Go bindings to system LLVM

[![Build Status](https://github.com/xgo-dev/llvm/actions/workflows/go.yml/badge.svg)](https://github.com/xgo-dev/llvm/actions/workflows/go.yml)
[![GoDoc](https://pkg.go.dev/badge/github.com/xgo-dev/llvm.svg)](https://pkg.go.dev/github.com/xgo-dev/llvm)
<!--
[![GitHub release](https://img.shields.io/github/v/tag/goplus/llvm.svg?label=release)](https://github.com/xgo-dev/llvm/releases)
[![Coverage Status](https://codecov.io/gh/goplus/llvm/branch/main/graph/badge.svg)](https://codecov.io/gh/goplus/llvm)
-->

This library provides bindings to a system-installed LLVM.

Currently supported:

  * LLVM 20, 19, 18, 17, 16, 15 and 14 from [apt.llvm.org](http://apt.llvm.org/) on Debian/Ubuntu.
  * LLVM 20, 19, 18, 17, 16, 15 and 14 from Homebrew on macOS.
  * LLVM 19 in an MSYS2 CLANG64 environment on Windows (see the setup below).
  * Any of the above versions with a manually built LLVM through the `byollvm` build tag. You need to set up `CFLAGS`/`LDFLAGS` etc yourself in this case.

You can select the LLVM version using a build tag, for example `-tags=llvm17`
to use LLVM 17.

## Usage

If you have a supported LLVM installation, you should be able to do a simple `go get`:

    go get github.com/xgo-dev/llvm

You can use build tags to select a LLVM version. For example, use `-tags=llvm15` to select LLVM 15. Setting a build tag for a LLVM version that is not supported will be ignored.

### Windows (MSYS2 CLANG64)

The LLVM 19 binding expects `pkg-config` metadata named `llvm-19`. MSYS2 does
not provide that versioned file, so after installing LLVM 19 and `pkgconf` in a
CLANG64 environment, generate it from `llvm-config`:

    pc_dir="$PWD/.llvm-pkgconfig"
    mkdir -p "$pc_dir"
    llvm_version="$(llvm-config --version)"
    test "$llvm_version" = 19.1.7
    cflags="$(llvm-config --cflags | tr '\r\n' '  ')"
    ldflags="$(llvm-config --ldflags --libs all --system-libs | tr '\r\n' '  ')"
    printf '%s\n' \
      'Name: LLVM 19' \
      'Description: LLVM 19 host compiler and linker flags' \
      "Version: $llvm_version" \
      "Cflags: $cflags" \
      "Libs: $ldflags" \
      > "$pc_dir/llvm-19.pc"
    export PKG_CONFIG_PATH="$(cygpath -m "$pc_dir")"
    export CC=clang CXX=clang++ CGO_ENABLED=1

The Windows CI job in [`.github/workflows/go.yml`](.github/workflows/go.yml)
shows the exact pinned MSYS2 packages used by this project.

## License

These LLVM bindings for Go originally come from LLVM, but they have since been [removed](https://discourse.llvm.org/t/rfc-remove-the-go-bindings/65725). Still, they remain under the same license as they were originally, which is the [Apache License 2.0 (with LLVM exceptions)](http://releases.llvm.org/9.0.0/LICENSE.TXT). Check upstream LLVM for detailed copyright information.

This README, the backports\* files, and the Makefile are separate from LLVM but are licensed under the same license.
