# lzfsec

[![ci](https://github.com/go-compressions/lzfsec/actions/workflows/ci.yml/badge.svg)](https://github.com/go-compressions/lzfsec/actions/workflows/ci.yml)
![coverage](https://img.shields.io/badge/coverage-100%25-brightgreen)

CLI wrapper around [`github.com/go-compressions/lzfse`](../lzfse) — Apple's
LZFSE / LZVN compression formats, in pure Go.

## Module

```text
github.com/go-compressions/lzfsec
```

## Commands

```sh
lzfsec compress   [-i input] [-o output]
lzfsec decompress [-i input] [-o output]
```

`-i` / `--input` defaults to stdin, `-o` / `--output` defaults to stdout.
When writing to a file (`-o`) a short summary line is printed to stderr
(byte counts and compression ratio for `compress`).

## Examples

```sh
# Compress a file to disk.
lzfsec compress -i big.bin -o big.bin.lzfse

# Round-trip through a pipe.
cat big.bin | lzfsec compress | lzfsec decompress > restored.bin
```

## Build

```sh
go build ./cmd/lzfsec
```

Or via Taskfile:

```sh
task build
```

## Development

The package ships a [Taskfile](https://taskfile.dev) for the common
build, test, and lint targets used by both local development and the
GitHub Actions workflow at [.github/workflows/ci.yml](.github/workflows/ci.yml).

```sh
task lint    # go vet
task build   # go build
task test    # go test -race + combined coverage across sub-packages
task ci      # lint + build + test, what CI runs
```

Dependency updates are handled by Renovate ([renovate.json](renovate.json));
patch and minor `gomod` updates auto-merge.

## Test coverage

`task test` reports **100 % statement coverage** ([`cover.out`](cover.out))
across the four sub-packages:

| Package                              | Role                                       |
| ------------------------------------ | ------------------------------------------ |
| `cmd/lzfsec`                         | `main` and cobra root command              |
| `cmd/lzfsec/compress`                | `compress` sub-command                     |
| `cmd/lzfsec/decompress`              | `decompress` sub-command                   |
| `cmd/lzfsec/internal/cmdio`          | shared stdin/stdout/file IO helpers        |
