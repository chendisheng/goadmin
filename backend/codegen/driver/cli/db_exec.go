package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	irbuilderapp "goadmin/codegen/application/irbuilder"
	"goadmin/core/config"
	infradb "goadmin/infrastructure/db"
)

// DatabaseExecutionRequest captures the shared database-driven codegen inputs.
type DatabaseExecutionRequest struct {
	Driver           string
	DSN              string
	Database         string
	Schema           string
	Tables           []string
	Force            bool
	GenerateFrontend *bool
	GeneratePolicy   *bool
}

// ExecuteDatabaseDocument inspects the configured database, converts it through
// IR/DSL/planner, and optionally writes files via the existing generator flow.
func ExecuteDatabaseDocument(root string, builder *irbuilderapp.Service, req DatabaseExecutionRequest, dryRun bool) (DSLExecutionReport, error) {
	if builder == nil {
		builder = irbuilderapp.NewService(irbuilderapp.Dependencies{})
	}
	if strings.TrimSpace(root) == "" {
		return DSLExecutionReport{}, fmt.Errorf("project root is required")
	}
	if strings.TrimSpace(req.Driver) == "" {
		return DSLExecutionReport{}, fmt.Errorf("database driver is required")
	}
	if strings.TrimSpace(req.DSN) == "" {
		return DSLExecutionReport{}, fmt.Errorf("database dsn is required")
	}
	if strings.TrimSpace(req.Database) == "" {
		return DSLExecutionReport{}, fmt.Errorf("database name is required")
	}
	conn, err := infradb.Open(config.DatabaseConfig{Driver: req.Driver, DSN: req.DSN})
	if err != nil {
		return DSLExecutionReport{}, err
	}
	opts := irbuilderapp.DatabaseBuildOptions{
		Tables:           append([]string(nil), req.Tables...),
		Force:            req.Force,
		GenerateFrontend: req.GenerateFrontend,
		GeneratePolicy:   req.GeneratePolicy,
	}
	doc, err := builder.BuildSchemaDocumentWithOptions(conn, req.Database, req.Schema, opts)
	if err != nil {
		return DSLExecutionReport{}, err
	}
	if _, err := builder.PlanSchemaDocumentWithOptions(doc, opts); err != nil {
		return DSLExecutionReport{}, err
	}
	resources, err := doc.ResolveResources()
	if err != nil {
		return DSLExecutionReport{}, err
	}
	report := BuildDSLExecutionReport(resources, req.Force, dryRun)
	if dryRun {
		return report, nil
	}
	if err := ExecuteDSLResources(root, resources, req.Force); err != nil {
		return DSLExecutionReport{}, err
	}
	report.DryRun = false
	report.Messages = append(report.Messages, fmt.Sprintf("generated %d resource(s)", len(report.Items)))
	return report, nil
}

func runGenerateDB(root string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("generate db requires a subcommand: preview, generate")
	}
	return runGenerateDBCommand(root, args)
}

func runGenerateDBCommand(root string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("generate db requires a subcommand: preview, generate")
	}
	mode := strings.ToLower(strings.TrimSpace(args[0]))
	fs := flag.NewFlagSet("generate db "+mode, flag.ContinueOnError)
	driver := fs.String("driver", "", "database driver name")
	dsn := fs.String("dsn", "", "database dsn")
	database := fs.String("database", "", "database name")
	schemaName := fs.String("schema", "", "database schema")
	force := fs.Bool("force", false, "overwrite existing files")
	frontend := fs.Bool("generate_frontend", true, "generate frontend scaffolding")
	policy := fs.Bool("generate_policy", true, "append Casbin policy lines")
	var tables stringListFlag
	fs.Var(&tables, "table", "database table to include; repeat for multiple tables")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	if strings.TrimSpace(*driver) == "" {
		return fmt.Errorf("generate db requires --driver")
	}
	if strings.TrimSpace(*dsn) == "" {
		return fmt.Errorf("generate db requires --dsn")
	}
	if strings.TrimSpace(*database) == "" {
		return fmt.Errorf("generate db requires --database")
	}
	request := DatabaseExecutionRequest{
		Driver:           *driver,
		DSN:              *dsn,
		Database:         *database,
		Schema:           *schemaName,
		Tables:           tables.Values(),
		Force:            *force,
		GenerateFrontend: frontend,
		GeneratePolicy:   policy,
	}
	switch mode {
	case "preview":
		report, err := ExecuteDatabaseDocument(root, nil, request, true)
		if err != nil {
			return err
		}
		return previewExecutionReport(report)
	case "generate":
		_, err := ExecuteDatabaseDocument(root, nil, request, false)
		return err
	default:
		return fmt.Errorf("unknown generate db subcommand %q", args[0])
	}
}

func previewExecutionReport(report DSLExecutionReport) error {
	for _, message := range report.Messages {
		if _, err := fmt.Fprintln(os.Stdout, message); err != nil {
			return err
		}
	}
	for _, item := range report.Items {
		if _, err := fmt.Fprintf(os.Stdout, "resource[%d] kind=%s name=%s\n", item.Index, item.Kind, item.Name); err != nil {
			return err
		}
		for _, action := range item.Actions {
			if _, err := fmt.Fprintf(os.Stdout, "  - %s\n", action); err != nil {
				return err
			}
		}
	}
	return nil
}

type stringListFlag []string

func (s *stringListFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringListFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	*s = append(*s, value)
	return nil
}

func (s *stringListFlag) Values() []string {
	if len(*s) == 0 {
		return nil
	}
	return append([]string(nil), *s...)
}
