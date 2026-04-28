package cli

import (
	"fmt"
	"strings"

	legacygenerate "goadmin/cli/generate"
	"goadmin/codegen/planner"
	"goadmin/codegen/schema"
)

type DSLPreviewResource struct {
	Index   int      `json:"index"`
	Kind    string   `json:"kind"`
	Name    string   `json:"name"`
	Force   bool     `json:"force"`
	Actions []string `json:"actions"`
}

type DSLExecutionReport struct {
	DryRun   bool                 `json:"dry_run"`
	Messages []string             `json:"messages,omitempty"`
	Items    []DSLPreviewResource `json:"items,omitempty"`
}

func ParseDSLResources(data []byte) (schema.Document, []schema.Resource, error) {
	doc, err := schema.ParseYAML(data)
	if err != nil {
		return schema.Document{}, nil, err
	}
	resources, err := doc.ResolveResources()
	if err != nil {
		return schema.Document{}, nil, err
	}
	return doc, resources, nil
}

func ParseDSLResourcesFromFile(path string) (schema.Document, []schema.Resource, error) {
	doc, err := schema.ParseYAMLFile(path)
	if err != nil {
		return schema.Document{}, nil, err
	}
	resources, err := doc.ResolveResources()
	if err != nil {
		return schema.Document{}, nil, err
	}
	return doc, resources, nil
}

func BuildDSLExecutionReport(resources []schema.Resource, force bool, dryRun bool) DSLExecutionReport {
	normalized := normalizeDSLResources(resources, force)
	report := DSLExecutionReport{
		DryRun:   dryRun,
		Messages: make([]string, 0, 2),
		Items:    make([]DSLPreviewResource, 0, len(normalized)),
	}
	if dryRun {
		report.Messages = append(report.Messages, "dsl dry-run: no files will be written")
	}
	if len(normalized) == 0 {
		report.Messages = append(report.Messages, "dsl contains no resources")
		return report
	}
	for i, resource := range normalized {
		report.Items = append(report.Items, DSLPreviewResource{
			Index:   i,
			Kind:    string(resource.Kind),
			Name:    resource.Name,
			Force:   resource.Force,
			Actions: describeSchemaResourceActions(resource),
		})
	}
	return report
}

func ExecuteDSLResources(root string, resources []schema.Resource, force bool) error {
	gen := legacygenerate.New(root)
	for _, resource := range resources {
		if err := generateFromSchemaResource(gen, resource, force); err != nil {
			return err
		}
	}
	return nil
}

func ExecuteDSLDocument(root string, data []byte, force bool, dryRun bool) (DSLExecutionReport, error) {
	doc, resources, err := ParseDSLResources(data)
	if err != nil {
		return DSLExecutionReport{}, err
	}
	if _, err := planner.New().Plan(doc); err != nil {
		return DSLExecutionReport{}, err
	}
	report := BuildDSLExecutionReport(resources, force, dryRun)
	if dryRun {
		return report, nil
	}
	if err := ExecuteDSLResources(root, normalizeDSLResources(resources, force), force); err != nil {
		return DSLExecutionReport{}, err
	}
	report.DryRun = false
	report.Messages = append(report.Messages, fmt.Sprintf("generated %d resource(s)", len(report.Items)))
	return report, nil
}

func normalizeDSLResources(resources []schema.Resource, force bool) []schema.Resource {
	if len(resources) == 0 {
		return nil
	}
	normalized := make([]schema.Resource, len(resources))
	for i, resource := range resources {
		normalized[i] = resource
		if force {
			normalized[i].Force = true
		}
	}
	return normalized
}

func joinDSLMessages(messages []string) string {
	parts := make([]string, 0, len(messages))
	for _, message := range messages {
		if strings.TrimSpace(message) == "" {
			continue
		}
		parts = append(parts, strings.TrimSpace(message))
	}
	return strings.Join(parts, "; ")
}
