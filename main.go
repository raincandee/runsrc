package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const version = "0.1.0"

type options struct {
	run bool
}

const usageHead = `Usage: runsrc [flags] [lang] <file> [args...]   compile a source file
       runsrc [flags] cargo [dir] [args...]     cargo build in dir
       runsrc [flags] make  [dir] [targets...]  make in dir
       runsrc [flags] <dir>                     auto-detect Cargo.toml or Makefile

Flags (put them before the file or project):
  -r, --run      run the program after building it; args after the file go to it
  -h, --help     show this help
      --version  show the version

By default runsrc only compiles. It prints the command that runs the result and
does not execute anything unless --run is given. A leading "--" after the file
is dropped, so arguments that begin with a dash can still be passed through.

The language is inferred from the file extension. Name it explicitly
(e.g. "runsrc rs notes.txt") when the extension is missing or misleading.

Languages:
`

const usageTail = `
Environment:
  RUNSRC_CFLAGS  flags for the C and C++ compilers (default: -Wall -Wextra -g)
  RUNSRC_PYTHON  interpreter used to run nuitka for .py files (default: python3)

Examples:
  runsrc main.cpp               build only; prints how to run it
  runsrc --run main.cpp 1 2 3   build, then run with arguments 1 2 3
  runsrc rs scratch.txt         compile a Rust file that has no .rs extension
  runsrc --run cargo ~/app -- --port 8080
  runsrc make run               run the "run" target of the Makefile here
`

func main() {
	os.Exit(cli(os.Args[1:]))
}

func cli(args []string) int {
	var opts options

parse:
	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		flag := args[0]
		args = args[1:]
		switch flag {
		case "-h", "--help":
			fmt.Print(usage())
			return 0
		case "--version":
			fmt.Println("runsrc", version)
			return 0
		case "-r", "--run":
			opts.run = true
		case "--":
			break parse
		default:
			return fail("unknown flag %q (see runsrc --help)", flag)
		}
	}

	if len(args) == 0 {
		fmt.Fprint(os.Stderr, usage())
		return 2
	}

	switch args[0] {
	case "cargo":
		return projectCargo(args[1:], opts)
	case "make":
		return projectMake(args[1:])
	}

	var lang *Language
	if len(args) > 1 {
		if l := findLanguage(args[0]); l != nil {
			lang = l
			args = args[1:]
		}
	}

	if lang == nil && isDir(args[0]) {
		return projectAuto(args[0], args[1:], opts)
	}
	return runFile(lang, args, opts)
}

func runFile(lang *Language, args []string, opts options) int {
	file := args[0]
	progArgs := args[1:]
	if len(progArgs) > 0 && progArgs[0] == "--" {
		progArgs = progArgs[1:]
	}

	if isDir(file) {
		return fail("'%s' is a directory (use runsrc <dir>, runsrc cargo, or runsrc make)", file)
	}
	if !exists(file) {
		return fail("'%s': no such file (see runsrc --help)", file)
	}

	if lang == nil {
		lang = findByExt(file)
		if lang == nil {
			return fail("cannot infer language from '%s'; name it explicitly, e.g. runsrc rs %s",
				filepath.Ext(file), file)
		}
	}

	cache, err := cacheDir()
	if err != nil {
		return fail("cannot create cache directory: %v", err)
	}

	name := strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
	c := buildCtx{
		Src:    file,
		Name:   name,
		Cache:  cache,
		Bin:    filepath.Join(cache, name),
		CFlags: strings.Fields(cflags()),
		Python: pythonPath(),
	}

	steps := lang.Compile(c)
	for _, step := range steps {
		if err := requireTool(step[0]); err != nil {
			return fail("%v", err)
		}
	}
	for _, step := range steps {
		code, err := runCommand("", step)
		if err != nil {
			return fail("%v", err)
		}
		if code != 0 {
			fmt.Fprintf(os.Stderr, "runsrc: build failed (%s exited with status %d)\n", step[0], code)
			return 1
		}
	}

	execArgv := lang.Exec(c)
	if !opts.run {
		fmt.Fprintf(os.Stderr, "runsrc: built, run with: %s\n", strings.Join(execArgv, " "))
		return 0
	}

	if err := requireTool(execArgv[0]); err != nil {
		return fail("%v", err)
	}
	code, err := runCommand("", append(execArgv, progArgs...))
	if err != nil {
		return fail("%v", err)
	}
	if code != 0 {
		fmt.Fprintf(os.Stderr, "runsrc: program exited with status %d\n", code)
	}
	return code
}

func usage() string {
	var b strings.Builder
	b.WriteString(usageHead)
	for _, l := range languages {
		exts := "(name it explicitly)"
		if len(l.Exts) > 0 {
			exts = "." + strings.Join(l.Exts, " .")
		}
		fmt.Fprintf(&b, "  %-8s %s\n", l.Name, exts)
	}
	b.WriteString(usageTail)
	return b.String()
}

func fail(format string, a ...any) int {
	fmt.Fprintf(os.Stderr, "runsrc: "+format+"\n", a...)
	return 1
}

func isDir(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func findLanguage(name string) *Language {
	for _, l := range languages {
		if l.Name == name {
			return l
		}
	}
	return nil
}

func findByExt(file string) *Language {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file), "."))
	if ext == "" {
		return nil
	}
	for _, l := range languages {
		for _, e := range l.Exts {
			if e == ext {
				return l
			}
		}
	}
	return nil
}

func cacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "runsrc")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func cflags() string {
	if v := os.Getenv("RUNSRC_CFLAGS"); v != "" {
		return v
	}
	return "-Wall -Wextra -g"
}

func pythonPath() string {
	if p := os.Getenv("RUNSRC_PYTHON"); p != "" {
		return p
	}
	return "python3"
}
