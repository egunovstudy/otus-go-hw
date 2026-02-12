package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	from, to      string
	limit, offset int64
)

func init() {
	flag.StringVar(&from, "from", "", "file to read from")
	flag.StringVar(&to, "to", "", "file to write to")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy")
	flag.Int64Var(&offset, "offset", 0, "offset in input file")
}

func main() {
	flag.Parse()

	if from == "" || to == "" {
		_, _ = fmt.Fprintln(os.Stderr, "both -from and -to must be specified")
		os.Exit(1)
	}
	if offset < 0 || limit < 0 {
		_, _ = fmt.Fprintln(os.Stderr, "-offset and -limit must be non-negative")
		os.Exit(1)
	}

	if err := Copy(from, to, offset, limit); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
