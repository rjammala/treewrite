package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

var (
	flagHelp = flag.Bool("help", false, "If true print a help message and exit.")
	flagEdit = flag.Bool("edit", false, "If true, edit files in place.")
	flagFile = flag.String("apply", "",
		"If non-empty, pattern and replacement are read from the specified file.  The pattern comes first and is separated from the replacement by a line that consists entirely of dashes (at least three dashes are required).")
)

func usage(dst io.Writer) {
	fmt.Fprint(dst, `Usage

treewrite _pattern_ _replacement_ files...
    Replace occurrences of _pattern_ with _replacement_ in supplied files.
    If no files are supplied, read standard input.

treewrite -apply _filename_ files...
    Read pattern and replacement from _filename_ and apply to the specified
    files (or standard input if no input files are specified).

    The contents of _filename_ should be pattern followed by replacement,
    separated by a line containing entirely of dashes (at least three
    dashes are required).

treewrite -edit ...
    Read from each of the specified files (there must be at least one), apply
    the replacement, and write the result back to source file.

`)
}

func main() {
	flag.Parse()
	if *flagHelp {
		usage(os.Stdout)
		os.Exit(0)
	}
	args := flag.Args()
	var pattern, replacement []byte
	if *flagFile != "" {
		pattern, replacement = splitFile(*flagFile)
	} else {
		if len(args) < 2 {
			usage(os.Stderr)
			os.Exit(1)
		}
		pattern = []byte(args[0])
		replacement = []byte(args[1])
		args = args[2:]
	}

	pat := parse(pattern)
	rep := parse(replacement)

	if !*flagEdit {
		if len(args) == 0 {
			data, err := ioutil.ReadAll(os.Stdin)
			reportError(err)
			input := parse(data)
			replace(input, pat, rep)
			os.Stdout.Write(input.serialize())
			return
		}
		for _, fname := range args {
			data, err := ioutil.ReadFile(fname)
			reportError(err)
			input := parse(data)
			replace(input, pat, rep)
			os.Stdout.Write(input.serialize())
		}
		return
	}

	if len(args) == 0 {
		reportError(errors.New("Must specify at least one file with -edit flag."))
	}
	for _, fname := range args {
		data, err := ioutil.ReadFile(fname)
		reportError(err)
		input := parse(data)
		replace(input, pat, rep)
		reportError(saveFile(fname, input.serialize()))
	}
}

// splitFile extracts pattern and replacement text from named file.
func splitFile(file string) ([]byte, []byte) {
	data, err := ioutil.ReadFile(file)
	reportError(err)
	re := regexp.MustCompile("(?m)^---+\n")
	m := re.FindIndex(data)
	if m == nil {
		reportError(errors.New("no separator line in " + file))
		os.Exit(1)
	}
	return data[:m[0]], data[m[1]:]
}

// saveFile saves data to fname by writing to a temporary file and renaming.
func saveFile(fname string, data []byte) error {
	tmp, err := ioutil.TempFile(filepath.Dir(fname), filepath.Base(fname)+"-tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	_, err = tmp.Write(data)
	if err != nil {
		tmp.Close()
		return err
	}
	err = tmp.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmpName, fname)
}

func reportError(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
