package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/yuta-masano/goviz/dotwriter"
	"github.com/yuta-masano/goviz/goimport"
	"github.com/yuta-masano/goviz/metrics"
)

type options struct {
	InputDir   string `short:"i" long:"input" required:"true" description:"input project name"`
	OutputFile string `short:"o" long:"output" default:"STDOUT" description:"output file"`
	Depth      int    `short:"d" long:"depth" default:"128" description:"max plot depth of the dependency tree"`
	Reversed   string `short:"f" long:"focus" description:"focus on the specific module"`
	SeekPath   string `short:"s" long:"search" default:"" description:"top directory of searching"`
	PlotLeaf   bool   `short:"l" long:"leaf" description:"whether leaf nodes are plotted"`
	UseMetrics bool   `short:"m" long:"metrics" description:"display module metrics"`
	FilesShown int    `short:"e" long:"files-shown" default:"2147483647" description:"limit filenames displayed in a package"`
}

func getOptions() (*options, error) {
	options := new(options)
	_, err := flags.Parse(options)
	if err != nil {
		return nil, err
	}
	return options, nil

}
func main() {
	res := process()
	os.Exit(res)
}

func errorf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
}

func process() int {
	options, err := getOptions()
	if err != nil {
		return 1
	}
	factory := goimport.ParseRelation(
		options.InputDir,
		options.SeekPath,
		options.PlotLeaf,
	)
	if factory == nil {
		errorf("inputdir does not exist.\n go get %s", options.InputDir)
		return 1
	}
	root := factory.GetRoot()
	if !root.HasFiles() {
		errorf("%s has no .go files\n", root.ImportPath)
		return 1
	}
	if 0 > options.Depth {
		errorf("-d or --depth should have positive int\n")
		return 1
	}
	if 0 > options.FilesShown {
		errorf("-e or --files-shown should have positive int\n")
		return 1
	}
	output := getOutputWriter(options.OutputFile)
	if options.UseMetrics {
		metrics_writer := metrics.New(output)
		metrics_writer.Plot(pathToNode(factory.GetAll()))
		return 0
	}

	dotwriter.FilesShown = options.FilesShown
	writer := dotwriter.New(output)
	writer.MaxDepth = options.Depth
	if options.Reversed == "" {
		writer.PlotGraph(root)
		return 0
	}
	writer.Reversed = true

	rroot := factory.Get(options.Reversed)
	if rroot == nil {
		errorf("-r %s does not exist.\n ", options.Reversed)
		return 1
	}
	if !rroot.HasFiles() {
		errorf("-r %s has no go files.\n ", options.Reversed)
		return 1
	}

	writer.PlotGraph(rroot)
	return 0
}

func pathToNode(pathes []*goimport.ImportPath) []dotwriter.IDotNode {
	r := make([]dotwriter.IDotNode, len(pathes))

	for i, _ := range pathes {
		r[i] = pathes[i]
	}
	return r
}
func getOutputWriter(name string) *os.File {
	if name == "STDOUT" {
		return os.Stdout
	}
	if name == "STDERR" {
		return os.Stderr
	}
	f, _ := os.Create(name)
	return f
}
