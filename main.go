package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var version = "dev"

func init() {
	flag.CommandLine.Init("add", flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	// flag.Usage is a no-op; help output is handled explicitly in the ErrHelp branch.
	// This prevents double-printing since flag.ContinueOnError calls Usage internally
	// before returning ErrHelp.
	flag.Usage = func() {}
}

func main() {
	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "show version number")
	flag.BoolVar(&showVersion, "version", false, "show version number")

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			fmt.Print(helpText)
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if showVersion {
		fmt.Printf("add %s\n", version)
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) < 2 {
		printError("missing arguments, need two integers")
		os.Exit(1)
	}
	if len(args) > 2 {
		printError("too many arguments, only two integers accepted")
		os.Exit(1)
	}

	a, err := parseArg(args[0])
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}
	b, err := parseArg(args[1])
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	result, err := add(a, b)
	if err != nil {
		printError(err.Error())
		os.Exit(1)
	}

	fmt.Println(result)
}
