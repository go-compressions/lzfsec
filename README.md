<p align="center"><img src="https://raw.githubusercontent.com/go-compressions/brand/main/social/go-compressions-lzfsec.png" alt="go-compressions/lzfsec" width="720"></p>

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
lzfsec [-v] compress   [-i input] [-o output]
lzfsec [-v] decompress [-i input] [-o output]
```

- `-i` / `--input` defaults to stdin.
- `-o` / `--output` defaults to stdout.
- `-v` / `--verbose` (persistent) prints a summary line to stderr with
  byte counts, compression ratio (for `compress`), and elapsed time:

  ```text
  compressed 65536 → 12345 bytes (18.8%) in 4.231ms
  decompressed 12345 → 65536 bytes in 1.872ms
  ```

  Without `-v` both sub-commands stay silent so the binary output on
  stdout is safe to pipe.

## Examples

```sh
# Compress a file to disk.
lzfsec compress -i big.bin -o big.bin.lzfse

# Round-trip through a pipe.
cat big.bin | lzfsec compress | lzfsec decompress > restored.bin

# Show timing + ratio.
lzfsec -v compress -i big.bin -o big.bin.lzfse
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

## License

[BSD 3-Clause](LICENSE).

## Test coverage

`task test` reports **100 % statement coverage** ([`cover.out`](cover.out))
across the four sub-packages:

| Package                              | Role                                       |
| ------------------------------------ | ------------------------------------------ |
| `cmd/lzfsec`                         | `main` and cobra root command              |
| `cmd/lzfsec/compress`                | `compress` sub-command                     |
| `cmd/lzfsec/decompress`              | `decompress` sub-command                   |
| `cmd/lzfsec/internal/cmdio`          | shared stdin/stdout/file IO helpers        |
