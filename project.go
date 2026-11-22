package main

import "path/filepath"

// splitDir consumes a leading directory argument, defaulting to the current one.
func splitDir(args []string) (string, []string) {
	if len(args) > 0 && isDir(args[0]) {
		return args[0], args[1:]
	}
	return ".", args
}

func projectCargo(args []string, opts options) int {
	dir, rest := splitDir(args)
	if len(rest) > 0 && rest[0] == "--" {
		rest = rest[1:]
	}

	if err := requireTool("cargo"); err != nil {
		return fail("%v", err)
	}

	argv := []string{"cargo", "build", "-q"}
	if opts.run {
		argv = append([]string{"cargo", "run", "-q", "--"}, rest...)
	}

	code, err := runCommand(dir, argv)
	if err != nil {
		return fail("%v", err)
	}
	return code
}

func projectMake(args []string) int {
	dir, targets := splitDir(args)

	if err := requireTool("make"); err != nil {
		return fail("%v", err)
	}

	code, err := runCommand(dir, append([]string{"make"}, targets...))
	if err != nil {
		return fail("%v", err)
	}
	return code
}

func projectAuto(dir string, rest []string, opts options) int {
	withDir := append([]string{dir}, rest...)

	if exists(filepath.Join(dir, "Cargo.toml")) {
		return projectCargo(withDir, opts)
	}
	for _, name := range []string{"Makefile", "makefile", "GNUmakefile"} {
		if exists(filepath.Join(dir, name)) {
			return projectMake(withDir)
		}
	}
	return fail("'%s' has no Cargo.toml or Makefile", dir)
}
