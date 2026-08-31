package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/epub"
)

func main() {
	out := flag.String("o", "", "output file (default stdout)")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: epub [-o out.txt] file.epub\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	text, err := epub.ExtractText(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	if *out == "" {
		fmt.Print(text)
		return
	}
	if err := os.WriteFile(*out, []byte(text), 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "wrote %d bytes to %s\n", len(text), *out)
}
