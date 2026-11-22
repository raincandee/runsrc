# runsrc

Compile and run source files without having to remember compiler commands.

`runsrc` detects the language from the file extension, compiles the source, and optionally runs the resulting program. It also supports Cargo and Make projects.

By default, it only compiles. Pass `--run` to execute the result.

```sh
runsrc main.cpp
runsrc --run main.cpp 1 2
```

The first command compiles `main.cpp` and prints the command to run the resulting binary. The second does the same, but also runs it with `1` and `2` as arguments.

## Installation

The easiest way to install runsrc is through Go:

```sh
go install github.com/raincandee/runsrc@latest
```

You'll need Go 1.21 or newer. Make sure `$(go env GOPATH)/bin` is in your `PATH`.

If you don't have Go installed, grab it from [go.dev/dl](https://go.dev/dl/) or your package manager (for example, `sudo pacman -S go` on Arch Linux).

Alternatively, you can build it yourself:

```sh
git clone https://github.com/raincandee/runsrc
cd runsrc
go build -o runsrc .
```

runsrc doesn't include any compilers. It uses whatever tools are installed on your system, so you'll need the appropriate compiler for the language you're working with.

## Usage

```text
runsrc [flags] [lang] <file> [args...]
runsrc [flags] cargo [dir] [args...]
runsrc [flags] make [dir] [targets...]
runsrc [flags] <dir>
```

Flags go before the file or project path.

When using `--run`, arguments following the file or project are passed to the program. If an argument starts with `-`, use `--` to separate it from runsrc's arguments:

```sh
runsrc --run cargo ~/app -- --port 8080
```

### Flags

| Flag | Description |
| --- | --- |
| `-r`, `--run` | Run the program after building |
| `-h`, `--help` | Show usage and supported languages |
| `--version` | Print the version |

### Language selection

Language detection is based on the file extension. You can also specify a language manually:

```sh
runsrc rs scratch.txt
```

This is useful when working with files that have unusual extensions or none at all.

`clang` and `clangpp` must always be specified manually, since they don't have dedicated file extensions.

## Supported languages

| Name | Compiler | Extensions |
| --- | --- | --- |
| `c` | gcc | `.c` |
| `clang` | clang | (explicit) |
| `cpp` | g++ (`-std=c++23`) | `.cpp`, `.cc`, `.cxx` |
| `clangpp` | clang++ (`-std=c++23`) | (explicit) |
| `rs` | rustc | `.rs` |
| `go` | go build | `.go` |
| `zig` | zig build-exe | `.zig` |
| `java` | javac | `.java` |
| `kt` | kotlinc | `.kt`, `.kts` |
| `swift` | swiftc | `.swift` |
| `hs` | ghc | `.hs` |
| `f90` | gfortran | `.f90`, `.f95`, `.f03`, `.f08`, `.f` |
| `pas` | fpc | `.pas`, `.pp` |
| `nasm` | nasm + ld | `.asm` |
| `nim` | nim c | `.nim` |
| `d` | ldc2 | `.d` |
| `py` | Nuitka | `.py` |

### Notes

- **Java:** The filename needs to match the public class name.
- **Python:** Uses Nuitka to compile Python into standalone binaries. Install it for the interpreter selected by `RUNSRC_PYTHON` (`pip install nuitka`). Initial compilation may take a while.
- **NASM:** Currently limited to x86-64 Linux. The entry point must be `_start`.
- **Make:** Targets are executed immediately, so `--run` has no effect.

## Build cache

Compiled binaries are kept in a cache instead of being placed alongside the source files.

| Platform | Location |
| --- | --- |
| Linux | `$XDG_CACHE_HOME/runsrc` (usually `~/.cache/runsrc`) |
| macOS | `~/Library/Caches/runsrc` |

Cache entries are identified by the source filename, not its full path. This means two files with the same basename will share an entry, and building one will overwrite the other.

## Exit codes

| Code | Meaning |
| --- | --- |
| `0` | Build or execution succeeded |
| `1` | Build failed, required tool missing, or invalid arguments |
| `2` | No arguments provided |

When `--run` is used, runsrc returns the exit code of the executed program.

Without `--run`, it prints the command needed to execute the compiled binary to stderr.

## Environment variables

| Variable | Default | Description |
| --- | --- | --- |
| `RUNSRC_CFLAGS` | `-Wall -Wextra -g` | Compiler flags for `c`, `clang`, `cpp`, and `clangpp` |
| `RUNSRC_PYTHON` | `python3` | Python interpreter used for Nuitka |

## Adding a language

Language definitions are in `languages.go`.

Each language needs:
- A name.
- A list of supported extensions.
- A function that generates its compile command.
- A function that generates its run command.

Adding another language should only require a small entry in the definitions.

## License

MIT. See [LICENSE](LICENSE).
