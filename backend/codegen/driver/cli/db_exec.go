package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	irbuilderapp "goadmin/codegen/application/irbuilder"
	"goadmin/core/config"
	apperrors "goadmin/core/errors"
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

func (req DatabaseExecutionRequest) Validate() error {
	if strings.TrimSpace(req.Driver) == "" {
		return apperrors.New(apperrors.CodeBadRequest, "database driver is required")
	}
	if strings.TrimSpace(req.DSN) == "" {
		return apperrors.New(apperrors.CodeBadRequest, "database dsn is required")
	}
	if strings.TrimSpace(req.Database) == "" {
		return apperrors.New(apperrors.CodeBadRequest, "database name is required")
	}
	return nil
}

// ExecuteDatabaseDocument inspects the configured database, converts it through
// IR/DSL/planner, and optionally writes files via the existing generator flow.
func ExecuteDatabaseDocument(root string, builder *irbuilderapp.Service, req DatabaseExecutionRequest, dryRun bool) (DatabasePreviewReport, error) {
	if builder == nil {
		builder = irbuilderapp.NewService(irbuilderapp.Dependencies{})
	}
	if strings.TrimSpace(root) == "" {
		return DatabasePreviewReport{}, apperrors.New(apperrors.CodeBadRequest, "project root is required")
	}
	if err := req.Validate(); err != nil {
		return DatabasePreviewReport{}, err
	}
	conn, err := infradb.Open(config.DatabaseConfig{Driver: req.Driver, DSN: req.DSN})
	if err != nil {
		return DatabasePreviewReport{}, apperrors.New(apperrors.CodeInternal, "open database connection failed")
	}
	opts := irbuilderapp.DatabaseBuildOptions{
		Tables:           append([]string(nil), req.Tables...),
		Force:            req.Force,
		GenerateFrontend: req.GenerateFrontend,
		GeneratePolicy:   req.GeneratePolicy,
	}
	irDoc, err := builder.BuildFromDatabaseWithOptions(conn, req.Database, req.Schema, opts)
	if err != nil {
		return DatabasePreviewReport{}, err
	}
	schemaDoc := irbuilderapp.ConvertIRDocumentToSchemaDocumentWithOptions(irDoc, opts)
	plan, err := builder.PlanSchemaDocumentWithOptions(schemaDoc, opts)
	if err != nil {
		return DatabasePreviewReport{}, err
	}
	report, err := BuildDatabasePreviewReport(root, req, irDoc, schemaDoc, plan, dryRun)
	if err != nil {
		return DatabasePreviewReport{}, err
	}
	if dryRun {
		report.Messages = append([]string{"database preview: dry-run; no files will be written"}, report.Messages...)
		return report, nil
	}
	resources, err := schemaDoc.ResolveResources()
	if err != nil {
		return DatabasePreviewReport{}, err
	}
	if err := ExecuteDSLResources(root, resources, req.Force); err != nil {
		return DatabasePreviewReport{}, err
	}
	report.DryRun = false
	report.Messages = append(report.Messages, fmt.Sprintf("generated %d resource(s)", len(report.Resources)))
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
	if err := request.Validate(); err != nil {
		return err
	}
	switch mode {
	case "preview":
		report, err := ExecuteDatabaseDocument(root, nil, request, true)
		if err != nil {
			return err
		}
		return previewDatabaseReport(report)
	case "generate":
		_, err := ExecuteDatabaseDocument(root, nil, request, false)
		return err
	default:
		return fmt.Errorf("unknown generate db subcommand %q", args[0])
	}
}

func previewDatabaseReport(report DatabasePreviewReport) error {
	_, err := fmt.Fprint(os.Stdout, previewDatabaseReportText(report))
	return err
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
