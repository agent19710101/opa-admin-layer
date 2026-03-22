package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/agent19710101/opa-admin-layer/internal/admin"
	"github.com/agent19710101/opa-admin-layer/internal/httpapi"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}

	switch args[0] {
	case "render":
		return runRender(args[1:])
	case "validate":
		return runValidate(args[1:])
	case "serve":
		return runServe(args[1:])
	case "help", "-h", "--help":
		return usageError()
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usageText())
	}
}

func runRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	input := fs.String("input", "", "path to JSON admin spec")
	output := fs.String("output", "-", "output path or - for stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" {
		return errors.New("render requires -input")
	}

	spec, err := admin.LoadSpec(*input)
	if err != nil {
		return err
	}
	plan, err := admin.BuildPlan(spec)
	if err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	if *output == "-" {
		_, err = fmt.Println(string(encoded))
		return err
	}
	return os.WriteFile(*output, append(encoded, '\n'), 0o644)
}

func runValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	input := fs.String("input", "", "path to JSON admin spec")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *input == "" {
		return errors.New("validate requires -input")
	}

	spec, err := admin.LoadSpec(*input)
	if err != nil {
		return err
	}
	issues := admin.Validate(spec)
	if len(issues) > 0 {
		for _, issue := range issues {
			fmt.Fprintln(os.Stderr, issue)
		}
		return fmt.Errorf("validation failed with %d issue(s)", len(issues))
	}
	fmt.Println("validation passed")
	return nil
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", ":8080", "listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return httpapi.ListenAndServe(*addr)
}

func usageError() error {
	return errors.New(usageText())
}

func usageText() string {
	return `Usage:
  opa-admin-layer render -input spec.json [-output plan.json]
  opa-admin-layer validate -input spec.json
  opa-admin-layer serve -addr :8080`
}
