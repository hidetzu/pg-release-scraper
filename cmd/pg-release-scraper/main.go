package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/hidetzu/pg-release-scraper/internal/excel"
	"github.com/hidetzu/pg-release-scraper/internal/markdown"
	"github.com/hidetzu/pg-release-scraper/internal/scraper"
)

func main() {
	start := flag.String("start", "", "Start version (e.g. 14.0)")
	end := flag.String("end", "", "End version (e.g. 15.6)")
	output := flag.String("output", "./output", "Output directory")
	format := flag.String("format", "both", "Output format: md, xlsx, or both")
	useStdout := flag.Bool("stdout", false, "Write Markdown to stdout instead of an .md file (xlsx still goes to file when --format includes xlsx)")
	quiet := flag.Bool("quiet", false, "Suppress non-error output")
	printVersion := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Scrape PostgreSQL release notes for a version range and export to Excel and/or Markdown.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s --start <version> --end <version> [--format md|xlsx|both] [--output <dir>] [--stdout] [--quiet]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *printVersion {
		fmt.Println(toolVersion())
		return
	}

	if *start == "" || *end == "" {
		flag.Usage()
		os.Exit(2)
	}

	wantMD, wantXLSX, err := resolveFormat(*format, *useStdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	releases, err := scraper.Fetch(*start, *end)
	if err != nil {
		fmt.Fprintln(os.Stderr, "scrape failed:", err)
		os.Exit(1)
	}
	if !*quiet {
		fmt.Fprintf(os.Stderr, "fetched %d release note items\n", len(releases))
	}
	if len(releases) == 0 {
		fmt.Fprintf(os.Stderr, "warning: no release notes found for range %s..%s\n", *start, *end)
		return
	}

	if wantMD {
		if *useStdout {
			if err := markdown.Render(os.Stdout, releases, *start, *end); err != nil {
				fmt.Fprintln(os.Stderr, "render md failed:", err)
				os.Exit(1)
			}
		} else {
			path, err := markdown.Write(releases, *start, *end, *output)
			if err != nil {
				fmt.Fprintln(os.Stderr, "write md failed:", err)
				os.Exit(1)
			}
			if !*quiet {
				fmt.Fprintln(os.Stderr, "wrote", path)
			}
		}
	}

	if wantXLSX {
		path, err := excel.Write(releases, *output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "write xlsx failed:", err)
			os.Exit(1)
		}
		if !*quiet {
			fmt.Fprintln(os.Stderr, "wrote", path)
		}
	}
}

func resolveFormat(format string, useStdout bool) (md, xlsx bool, err error) {
	switch format {
	case "md":
		md = true
	case "xlsx":
		xlsx = true
	case "both":
		md = true
		xlsx = true
	default:
		return false, false, fmt.Errorf("invalid --format value %q (must be md, xlsx, or both)", format)
	}
	if useStdout && !md {
		return false, false, fmt.Errorf("--stdout requires --format md or both")
	}
	return md, xlsx, nil
}

func toolVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}
