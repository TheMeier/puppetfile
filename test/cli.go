package main

import (
	"bufio"
	"log"
	"os"

	"github.com/TheMeier/puppetfile"

	"github.com/alecthomas/repr"
)

func main() {
	parser := puppetfile.New()
	content, err := os.Open("test/puppetfile")
	if err != nil {
		log.Fatalf("failed to read puppetfile: %v", err)
	}
	defer content.Close()

	expr, err := parser.Parse("", bufio.NewReader(content))
	if err != nil {
		log.Fatalf("failed to parse puppetfile: %v", err)
	}

	if err := expr.Validate(); err != nil {
		log.Fatalf("validation failed: %v", err)
	}

	repr.Println(expr)

}
