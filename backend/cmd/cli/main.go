package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"goadmin/cli/generate"
)

func main() {
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	gen := generate.New(root)

	switch os.Args[1] {
	case "generate":
		if err := runGenerate(gen, os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func runGenerate(gen *generate.Generator, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("generate requires a subcommand: module, crud, plugin")
	}
	switch args[0] {
	case "module":
		return runGenerateModule(gen, args[1:])
	case "crud":
		return runGenerateCRUD(gen, args[1:])
	case "plugin":
		return runGeneratePlugin(gen, args[1:])
	default:
		return fmt.Errorf("unknown generate subcommand %q", args[0])
	}
}

func runGenerateModule(gen *generate.Generator, args []string) error {
	fs := flag.NewFlagSet("generate module", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite existing files")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("generate module requires a module name")
	}
	return gen.GenerateModule(generate.ModuleOptions{Name: fs.Arg(0), Force: *force})
}

func runGenerateCRUD(gen *generate.Generator, args []string) error {
	fs := flag.NewFlagSet("generate crud", flag.ContinueOnError)
	fields := fs.String("fields", "", "comma separated field definitions like name:string,status:string")
	primary := fs.String("primary", "", "comma separated primary key fields")
	indexes := fs.String("index", "", "comma separated indexed fields")
	uniques := fs.String("unique", "", "comma separated unique fields")
	frontend := fs.Bool("frontend", true, "generate frontend scaffolding")
	policy := fs.Bool("policy", true, "append Casbin policy lines")
	force := fs.Bool("force", false, "overwrite existing files")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("generate crud requires an entity name")
	}
	parsedFields, err := generate.ParseFields(*fields, *primary, *indexes, *uniques)
	if err != nil {
		return err
	}
	return gen.GenerateCRUD(generate.CRUDOptions{
		Name:             fs.Arg(0),
		Fields:           parsedFields,
		GenerateFrontend: *frontend,
		GeneratePolicy:   *policy,
		Force:            *force,
	})
}

func runGeneratePlugin(gen *generate.Generator, args []string) error {
	fs := flag.NewFlagSet("generate plugin", flag.ContinueOnError)
	force := fs.Bool("force", false, "overwrite existing files")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("generate plugin requires a plugin name")
	}
	return gen.GeneratePlugin(generate.PluginOptions{Name: fs.Arg(0), Force: *force})
}

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("detect cwd: %w", err)
	}
	current := cwd
	for {
		if fileExists(filepath.Join(current, "go.work")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return cwd, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func usage() {
	fmt.Fprintf(os.Stderr, strings.TrimSpace(`
Usage:
  goadmin-cli generate module <name> [--force]
  goadmin-cli generate crud <name> [--fields name:string,status:string] [--primary id] [--index name] [--unique code] [--frontend] [--policy] [--force]
  goadmin-cli generate plugin <name> [--force]

Examples:
  goadmin-cli generate module user
  goadmin-cli generate crud order --fields id:string,name:string,status:string --policy
  goadmin-cli generate plugin demo
`)+"\n")
}
