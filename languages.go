package main

// buildCtx holds the paths and settings for compiling one source file.
type buildCtx struct {
	Src    string   // source file, as given on the command line
	Name   string   // file name without its extension
	Cache  string   // runsrc's cache directory
	Bin    string   // output path inside Cache
	CFlags []string // compiler flags from RUNSRC_CFLAGS
	Python string   // interpreter used to run nuitka
}

// Language describes how to turn a source file into something runnable.
// Compile returns the commands to run in order; each must succeed.
// Exec returns the command that runs the result.
type Language struct {
	Name    string
	Exts    []string
	Compile func(c buildCtx) [][]string
	Exec    func(c buildCtx) []string
}

func binaryExec(c buildCtx) []string { return []string{c.Bin} }

// cc builds a C-family compile command with the user's CFLAGS included.
func cc(tool string, c buildCtx, rest ...string) [][]string {
	argv := append([]string{tool}, c.CFlags...)
	argv = append(argv, rest...)
	return [][]string{argv}
}

var languages = []*Language{
	{
		Name: "c", Exts: []string{"c"},
		Compile: func(c buildCtx) [][]string { return cc("gcc", c, c.Src, "-o", c.Bin) },
		Exec:    binaryExec,
	},
	{
		Name:    "clang",
		Compile: func(c buildCtx) [][]string { return cc("clang", c, c.Src, "-o", c.Bin) },
		Exec:    binaryExec,
	},
	{
		Name: "cpp", Exts: []string{"cpp", "cc", "cxx"},
		Compile: func(c buildCtx) [][]string { return cc("g++", c, "-std=c++23", c.Src, "-o", c.Bin) },
		Exec:    binaryExec,
	},
	{
		Name:    "clangpp",
		Compile: func(c buildCtx) [][]string { return cc("clang++", c, "-std=c++23", c.Src, "-o", c.Bin) },
		Exec:    binaryExec,
	},
	{
		Name: "rs", Exts: []string{"rs"},
		Compile: func(c buildCtx) [][]string { return [][]string{{"rustc", c.Src, "-o", c.Bin}} },
		Exec:    binaryExec,
	},
	{
		Name: "go", Exts: []string{"go"},
		Compile: func(c buildCtx) [][]string { return [][]string{{"go", "build", "-o", c.Bin, c.Src}} },
		Exec:    binaryExec,
	},
	{
		Name: "zig", Exts: []string{"zig"},
		Compile: func(c buildCtx) [][]string {
			return [][]string{{"zig", "build-exe", c.Src, "-femit-bin=" + c.Bin}}
		},
		Exec: binaryExec,
	},
	{
		Name: "java", Exts: []string{"java"},
		Compile: func(c buildCtx) [][]string {
			return [][]string{{"javac", "-d", c.Bin + ".classes", c.Src}}
		},
		Exec: func(c buildCtx) []string {
			return []string{"java", "-cp", c.Bin + ".classes", c.Name}
		},
	},
	{
		Name: "kt", Exts: []string{"kt", "kts"},
		Compile: func(c buildCtx) [][]string {
			return [][]string{{"kotlinc", c.Src, "-include-runtime", "-d", c.Bin + ".jar"}}
		},
		Exec: func(c buildCtx) []string { return []string{"java", "-jar", c.Bin + ".jar"} },
	},
	{
		Name: "swift", Exts: []string{"swift"},
		Compile: func(c buildCtx) [][]string { return [][]string{{"swiftc", c.Src, "-o", c.Bin}} },
		Exec:    binaryExec,
	},
	{
		Name: "hs", Exts: []string{"hs"},
		Compile: func(c buildCtx) [][]string {
			return [][]string{{"ghc", "-O0", "-outputdir", c.Cache + "/ghc", "-o", c.Bin, c.Src}}
		},
		Exec: binaryExec,
	},
	{
		Name: "f90", Exts: []string{"f90", "f95", "f03", "f08", "f"},
		Compile: func(c buildCtx) [][]string { return [][]string{{"gfortran", c.Src, "-o", c.Bin}} },
		Exec:    binaryExec,
	},
	{
		Name: "pas", Exts: []string{"pas", "pp"},
		Compile: func(c buildCtx) [][]string {
			return [][]string{{"fpc", "-FU" + c.Cache, "-o" + c.Bin, c.Src}}
		},
		Exec: binaryExec,
	},
	{
		// Requires x86-64 Linux. The program entry point must be _start.
		Name: "nasm", Exts: []string{"asm"},
		Compile: func(c buildCtx) [][]string {
			return [][]string{
				{"nasm", "-f", "elf64", "-o", c.Bin + ".o", c.Src},
				{"ld", "-o", c.Bin, c.Bin + ".o"},
			}
		},
		Exec: binaryExec,
	},
	{
		Name: "nim", Exts: []string{"nim"},
		Compile: func(c buildCtx) [][]string {
			return [][]string{{"nim", "c", "-o:" + c.Bin, "--nimcache:" + c.Cache + "/nimcache", c.Src}}
		},
		Exec: binaryExec,
	},
	{
		Name: "d", Exts: []string{"d"},
		Compile: func(c buildCtx) [][]string {
			return [][]string{{"ldc2", "-od=" + c.Cache, "-of=" + c.Bin, c.Src}}
		},
		Exec: binaryExec,
	},
	{
		Name: "py", Exts: []string{"py"},
		Compile: func(c buildCtx) [][]string {
			return [][]string{{c.Python, "-m", "nuitka", "--onefile", "--remove-output", "-o", c.Bin, c.Src}}
		},
		Exec: binaryExec,
	},
}
