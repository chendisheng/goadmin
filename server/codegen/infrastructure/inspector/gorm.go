package inspector

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"goadmin/codegen/driver/db"
	codeschema "goadmin/codegen/schema"
	dbschema "goadmin/codegen/schema/database"

	"gorm.io/gorm"
)

type enumParseResult struct {
	Kind    string
	Mode    string
	Display string
	Values  []string
	Options []dbschema.EnumOption
	Source  string
	Ref     string
	OK      bool
}

// GormInspector is a read-only database inspector backed by gorm.DB.
type GormInspector struct {
	db       *gorm.DB
	driver   dbschema.DriverKind
	database string
	schema   string
}

var _ db.Inspector = (*GormInspector)(nil)
var _ Reader = (*GormInspector)(nil)

// NewGormInspector creates a new read-only inspector.
func NewGormInspector(db *gorm.DB) *GormInspector {
	inspector := &GormInspector{db: db, driver: dbschema.DriverKindUnknown}
	if db == nil {
		return inspector
	}
	if db.Dialector != nil {
		inspector.driver = normalizeDriverKind(db.Dialector.Name())
	}
	if migrator := db.Migrator(); migrator != nil {
		inspector.database = strings.TrimSpace(migrator.CurrentDatabase())
	}
	return inspector
}

// WithContext returns a shallow copy with the given database and schema labels.
func (i *GormInspector) WithContext(database string, schema string) Reader {
	if i == nil {
		return &GormInspector{database: strings.TrimSpace(database), schema: strings.TrimSpace(schema)}
	}
	clone := *i
	clone.database = strings.TrimSpace(database)
	clone.schema = strings.TrimSpace(schema)
	return &clone
}

func (i *GormInspector) InspectTables() ([]dbschema.Table, error) {
	tables, err := i.listTables()
	if err != nil {
		return nil, err
	}
	result := make([]dbschema.Table, 0, len(tables))
	for _, name := range tables {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		comment, err := i.inspectTableComment(name)
		if err != nil {
			return nil, err
		}
		columns, err := i.InspectColumns(name)
		if err != nil {
			return nil, err
		}
		indexes, err := i.inspectIndexes(name)
		if err != nil {
			return nil, err
		}
		foreignKeys, err := i.InspectRelations(name)
		if err != nil {
			return nil, err
		}
		table := dbschema.Table{Name: name, Comment: comment, Columns: columns, Indexes: indexes, ForeignKeys: foreignKeys}
		for _, column := range columns {
			if column.Primary {
				table.PrimaryKeys = append(table.PrimaryKeys, column.Name)
			}
		}
		table.Schema = i.schema
		table.Metadata = map[string]any{
			"driver": i.driver,
		}
		if i.database != "" {
			table.Metadata["database"] = i.database
		}
		result = append(result, table)
	}
	return result, nil
}

func applyColumnComments(columns []dbschema.Column, comments map[string]string) {
	if len(columns) == 0 || len(comments) == 0 {
		return
	}
	for idx := range columns {
		name := strings.TrimSpace(columns[idx].Name)
		if name == "" {
			continue
		}
		if comment := strings.TrimSpace(comments[name]); comment != "" {
			columns[idx].Comment = comment
			applyEnumCommentMetadata(&columns[idx])
		}
	}
}

func (i *GormInspector) inspectTableComment(table string) (string, error) {
	switch i.driver {
	case dbschema.DriverKindSQLite, dbschema.DriverKindUnknown:
		return "", nil
	case dbschema.DriverKindMySQL, dbschema.DriverKindSQLServer:
		return i.inspectTableCommentFromMigrator(table)
	case dbschema.DriverKindPostgreSQL:
		return i.inspectPostgresTableComment(table)
	default:
		return i.inspectTableCommentFromMigrator(table)
	}
}

func (i *GormInspector) inspectTableCommentFromMigrator(table string) (string, error) {
	migrator, err := i.migrator()
	if err != nil {
		return "", err
	}
	commenter, ok := migrator.(interface{ Comment(string) (string, bool) })
	if !ok {
		return "", nil
	}
	comment, ok := commenter.Comment(table)
	if !ok {
		return "", nil
	}
	return strings.TrimSpace(comment), nil
}

func (i *GormInspector) inspectPostgresTableComment(table string) (string, error) {
	if i == nil || i.db == nil {
		return "", fmt.Errorf("gorm inspector db is nil")
	}
	var row struct {
		Comment sql.NullString `gorm:"column:comment"`
	}
	query := `
SELECT obj_description((quote_ident(?)::regclass)::oid) AS comment
`
	if err := i.db.Raw(query, table).Scan(&row).Error; err != nil {
		return "", fmt.Errorf("inspect postgres table comment for %s: %w", table, err)
	}
	if !row.Comment.Valid {
		return "", nil
	}
	return strings.TrimSpace(row.Comment.String), nil
}

func (i *GormInspector) InspectColumns(table string) ([]dbschema.Column, error) {
	switch i.driver {
	case dbschema.DriverKindSQLite, dbschema.DriverKindUnknown:
		columns, err := i.inspectSQLiteColumns(table)
		if err != nil {
			return nil, err
		}
		indexes, err := i.inspectIndexes(table)
		if err != nil {
			return nil, err
		}
		applyIndexFlags(columns, indexes)
		return columns, nil
	default:
		return i.inspectGenericColumns(table)
	}
}

func (i *GormInspector) InspectRelations(table string) ([]dbschema.ForeignKey, error) {
	switch i.driver {
	case dbschema.DriverKindSQLite, dbschema.DriverKindUnknown:
		return i.inspectSQLiteForeignKeys(table)
	default:
		return nil, nil
	}
}

func (i *GormInspector) migrator() (gorm.Migrator, error) {
	if i == nil || i.db == nil {
		return nil, fmt.Errorf("gorm inspector db is nil")
	}
	migrator := i.db.Migrator()
	if migrator == nil {
		return nil, fmt.Errorf("gorm migrator is nil")
	}
	return migrator, nil
}

func (i *GormInspector) listTables() ([]string, error) {
	switch i.driver {
	case dbschema.DriverKindSQLite, dbschema.DriverKindUnknown:
		var rows []struct {
			Name string `gorm:"column:name"`
		}
		if err := i.db.Raw(`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`).Scan(&rows).Error; err != nil {
			return nil, fmt.Errorf("list sqlite tables: %w", err)
		}
		tables := make([]string, 0, len(rows))
		for _, row := range rows {
			if name := strings.TrimSpace(row.Name); name != "" {
				tables = append(tables, name)
			}
		}
		return tables, nil
	default:
		migrator, err := i.migrator()
		if err != nil {
			return nil, err
		}
		tables, err := migrator.GetTables()
		if err != nil {
			return nil, fmt.Errorf("list tables: %w", err)
		}
		return tables, nil
	}
}

func (i *GormInspector) inspectIndexes(table string) ([]dbschema.Index, error) {
	switch i.driver {
	case dbschema.DriverKindSQLite, dbschema.DriverKindUnknown:
		return i.inspectSQLiteIndexes(table)
	default:
		return nil, nil
	}
}

func (i *GormInspector) inspectGenericColumns(table string) ([]dbschema.Column, error) {
	migrator, err := i.migrator()
	if err != nil {
		return nil, err
	}
	columnTypes, err := migrator.ColumnTypes(table)
	if err != nil {
		return nil, fmt.Errorf("inspect columns for %s: %w", table, err)
	}
	columnComments, err := i.inspectColumnComments(table)
	if err != nil {
		return nil, err
	}
	result := make([]dbschema.Column, 0, len(columnTypes))
	for _, columnType := range columnTypes {
		column, err := toColumn(columnType)
		if err != nil {
			return nil, fmt.Errorf("inspect column in %s: %w", table, err)
		}
		result = append(result, column)
	}
	applyColumnComments(result, columnComments)
	return result, nil
}

func (i *GormInspector) inspectColumnComments(table string) (map[string]string, error) {
	switch i.driver {
	case dbschema.DriverKindMySQL:
		return i.inspectMySQLColumnComments(table)
	case dbschema.DriverKindSQLite, dbschema.DriverKindUnknown:
		return map[string]string{}, nil
	default:
		return map[string]string{}, nil
	}
}

func (i *GormInspector) inspectMySQLColumnComments(table string) (map[string]string, error) {
	if i == nil || i.db == nil {
		return map[string]string{}, fmt.Errorf("gorm inspector db is nil")
	}
	databaseName := strings.TrimSpace(i.database)
	if databaseName == "" {
		migrator, err := i.migrator()
		if err != nil {
			return map[string]string{}, err
		}
		databaseName = strings.TrimSpace(migrator.CurrentDatabase())
	}
	if databaseName == "" {
		return map[string]string{}, nil
	}
	type row struct {
		Name    string         `gorm:"column:column_name"`
		Comment sql.NullString `gorm:"column:column_comment"`
	}
	var rows []row
	query := `
SELECT column_name, column_comment
FROM information_schema.columns
WHERE table_schema = ? AND table_name = ?
ORDER BY ordinal_position
`
	if err := i.db.Raw(query, databaseName, table).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("inspect mysql column comments for %s: %w", table, err)
	}
	result := make(map[string]string, len(rows))
	for _, row := range rows {
		name := strings.TrimSpace(row.Name)
		if name == "" || !row.Comment.Valid {
			continue
		}
		result[name] = strings.TrimSpace(row.Comment.String)
	}
	return result, nil
}

func (i *GormInspector) inspectSQLiteColumns(table string) ([]dbschema.Column, error) {
	type row struct {
		CID     int            `gorm:"column:cid"`
		Name    string         `gorm:"column:name"`
		Type    string         `gorm:"column:type"`
		NotNull int            `gorm:"column:notnull"`
		Default sql.NullString `gorm:"column:dflt_value"`
		PK      int            `gorm:"column:pk"`
	}
	var rows []row
	query := fmt.Sprintf("PRAGMA table_info(%s)", quoteIdentifier(table))
	if err := i.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("inspect sqlite columns for %s: %w", table, err)
	}
	result := make([]dbschema.Column, 0, len(rows))
	for _, row := range rows {
		column := dbschema.Column{
			Name:     strings.TrimSpace(row.Name),
			Type:     strings.TrimSpace(row.Type),
			Nullable: row.NotNull == 0,
			Primary:  row.PK > 0,
			Index:    false,
			Unique:   false,
			Metadata: map[string]any{
				"cid":      row.CID,
				"notnull":  row.NotNull == 1,
				"pk_order": row.PK,
			},
		}
		if row.Default.Valid {
			column.Default = strings.TrimSpace(row.Default.String)
		}
		applyEnumCommentMetadata(&column)
		result = append(result, column)
	}
	if autoIncrementColumns, err := i.inspectSQLiteAutoIncrementColumns(table); err == nil {
		markAutoIncrementColumns(result, autoIncrementColumns)
	}
	return result, nil
}

func (i *GormInspector) inspectSQLiteIndexes(table string) ([]dbschema.Index, error) {
	type listRow struct {
		Seq     int    `gorm:"column:seq"`
		Name    string `gorm:"column:name"`
		Unique  int    `gorm:"column:unique"`
		Origin  string `gorm:"column:origin"`
		Partial int    `gorm:"column:partial"`
	}
	type infoRow struct {
		SeqNo int    `gorm:"column:seqno"`
		CID   int    `gorm:"column:cid"`
		Name  string `gorm:"column:name"`
	}
	var listRows []listRow
	query := fmt.Sprintf("PRAGMA index_list(%s)", quoteIdentifier(table))
	if err := i.db.Raw(query).Scan(&listRows).Error; err != nil {
		return nil, fmt.Errorf("inspect sqlite indexes for %s: %w", table, err)
	}
	result := make([]dbschema.Index, 0, len(listRows))
	for _, listRow := range listRows {
		if strings.TrimSpace(listRow.Name) == "" {
			continue
		}
		var infoRows []infoRow
		infoQuery := fmt.Sprintf("PRAGMA index_info(%s)", quoteIdentifier(listRow.Name))
		if err := i.db.Raw(infoQuery).Scan(&infoRows).Error; err != nil {
			return nil, fmt.Errorf("inspect sqlite index columns for %s.%s: %w", table, listRow.Name, err)
		}
		sort.Slice(infoRows, func(a, b int) bool { return infoRows[a].SeqNo < infoRows[b].SeqNo })
		columns := make([]string, 0, len(infoRows))
		for _, infoRow := range infoRows {
			if name := strings.TrimSpace(infoRow.Name); name != "" {
				columns = append(columns, name)
			}
		}
		result = append(result, dbschema.Index{
			Name:    strings.TrimSpace(listRow.Name),
			Columns: columns,
			Unique:  listRow.Unique != 0,
			Primary: strings.EqualFold(strings.TrimSpace(listRow.Origin), "pk"),
			Type:    strings.TrimSpace(listRow.Origin),
			Metadata: map[string]any{
				"partial": listRow.Partial != 0,
				"origin":  strings.TrimSpace(listRow.Origin),
			},
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (i *GormInspector) inspectSQLiteForeignKeys(table string) ([]dbschema.ForeignKey, error) {
	type row struct {
		ID       int    `gorm:"column:id"`
		Seq      int    `gorm:"column:seq"`
		Table    string `gorm:"column:table"`
		From     string `gorm:"column:from"`
		To       string `gorm:"column:to"`
		OnUpdate string `gorm:"column:on_update"`
		OnDelete string `gorm:"column:on_delete"`
		Match    string `gorm:"column:match"`
	}
	var rows []row
	query := fmt.Sprintf("PRAGMA foreign_key_list(%s)", quoteIdentifier(table))
	if err := i.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("inspect sqlite relations for %s: %w", table, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	grouped := make(map[int]*dbschema.ForeignKey)
	ordered := make([]int, 0)
	for _, row := range rows {
		if strings.TrimSpace(row.Table) == "" || strings.TrimSpace(row.From) == "" {
			continue
		}
		fk, ok := grouped[row.ID]
		if !ok {
			fk = &dbschema.ForeignKey{
				Name:     fmt.Sprintf("fk_%s_%d", table, row.ID),
				Metadata: map[string]any{"id": row.ID, "match": strings.TrimSpace(row.Match)},
			}
			grouped[row.ID] = fk
			ordered = append(ordered, row.ID)
		}
		fk.RefTable = strings.TrimSpace(row.Table)
		fk.OnUpdate = strings.TrimSpace(row.OnUpdate)
		fk.OnDelete = strings.TrimSpace(row.OnDelete)
		fk.Columns = append(fk.Columns, strings.TrimSpace(row.From))
		fk.RefColumns = append(fk.RefColumns, strings.TrimSpace(row.To))
	}
	sort.Ints(ordered)
	result := make([]dbschema.ForeignKey, 0, len(ordered))
	for _, id := range ordered {
		if fk := grouped[id]; fk != nil {
			result = append(result, *fk)
		}
	}
	return result, nil
}

func (i *GormInspector) inspectSQLiteAutoIncrementColumns(table string) ([]string, error) {
	type row struct {
		SQL sql.NullString `gorm:"column:sql"`
	}
	var rows []row
	if err := i.db.Raw(`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("inspect sqlite create sql for %s: %w", table, err)
	}
	if len(rows) == 0 || !rows[0].SQL.Valid {
		return nil, nil
	}
	createSQL := strings.ToUpper(rows[0].SQL.String)
	if !strings.Contains(createSQL, "AUTOINCREMENT") {
		return nil, nil
	}
	columns, err := i.inspectSQLiteColumnsWithoutAutoIncrement(table)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(columns))
	for _, column := range columns {
		if column.Primary && strings.EqualFold(strings.TrimSpace(column.Type), "INTEGER") {
			result = append(result, column.Name)
		}
	}
	return result, nil
}

func (i *GormInspector) inspectSQLiteColumnsWithoutAutoIncrement(table string) ([]dbschema.Column, error) {
	type row struct {
		CID     int            `gorm:"column:cid"`
		Name    string         `gorm:"column:name"`
		Type    string         `gorm:"column:type"`
		NotNull int            `gorm:"column:notnull"`
		Default sql.NullString `gorm:"column:dflt_value"`
		PK      int            `gorm:"column:pk"`
	}
	var rows []row
	query := fmt.Sprintf("PRAGMA table_info(%s)", quoteIdentifier(table))
	if err := i.db.Raw(query).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("inspect sqlite columns for %s: %w", table, err)
	}
	result := make([]dbschema.Column, 0, len(rows))
	for _, row := range rows {
		column := dbschema.Column{
			Name:     strings.TrimSpace(row.Name),
			Type:     strings.TrimSpace(row.Type),
			Nullable: row.NotNull == 0,
			Primary:  row.PK > 0,
			Metadata: map[string]any{"cid": row.CID, "notnull": row.NotNull == 1, "pk_order": row.PK},
		}
		if row.Default.Valid {
			column.Default = strings.TrimSpace(row.Default.String)
		}
		result = append(result, column)
	}
	return result, nil
}

func toColumn(columnType gorm.ColumnType) (dbschema.Column, error) {
	if columnType == nil {
		return dbschema.Column{}, fmt.Errorf("column type is nil")
	}
	column := dbschema.Column{
		Name:     strings.TrimSpace(columnType.Name()),
		Type:     strings.TrimSpace(columnType.DatabaseTypeName()),
		Metadata: map[string]any{},
	}
	if length, ok := columnType.Length(); ok {
		value := int(length)
		column.Length = &value
	}
	if precision, scale, ok := columnType.DecimalSize(); ok {
		p := int(precision)
		s := int(scale)
		column.Precision = &p
		column.Scale = &s
	}
	if nullable, ok := columnType.Nullable(); ok {
		column.Nullable = nullable
	}
	if defaultValue, ok := columnType.DefaultValue(); ok {
		column.Default = strings.TrimSpace(defaultValue)
	}
	if primary, ok := columnType.PrimaryKey(); ok {
		column.Primary = primary
	}
	if autoIncrement, ok := columnType.AutoIncrement(); ok {
		column.AutoIncrement = autoIncrement
	}
	if scanner := columnType.ScanType(); scanner != nil {
		column.Metadata["scan_type"] = scanner.String()
	}
	if commenter, ok := columnType.(interface{ Comment() (string, bool) }); ok {
		if comment, ok := commenter.Comment(); ok {
			column.Comment = strings.TrimSpace(comment)
		}
	}
	applyEnumCommentMetadata(&column)
	return column, nil
}

func applyEnumCommentMetadata(column *dbschema.Column) {
	if column == nil {
		return
	}
	parsed := parseColumnCommentMetadata(column.Comment)
	if parsed.Label != "" {
		if column.Metadata == nil {
			column.Metadata = map[string]any{}
		}
		column.Metadata["label"] = parsed.Label
	}
	if parsed.UIType != "" {
		column.UIType = parsed.UIType
		if column.Metadata == nil {
			column.Metadata = map[string]any{}
		}
		column.Metadata["ui_type"] = parsed.UIType
	}
	if !parsed.Enum.OK {
		return
	}
	if column.Metadata == nil {
		column.Metadata = map[string]any{}
	}
	if column.UIType == "" {
		column.UIType = parsed.Enum.Display
	}
	column.EnumKind = parsed.Enum.Kind
	column.EnumMode = parsed.Enum.Mode
	column.EnumDisplay = parsed.Enum.Display
	column.EnumSource = parsed.Enum.Source
	column.EnumSourceRef = parsed.Enum.Ref
	column.EnumValues = append([]string(nil), parsed.Enum.Values...)
	column.EnumOptions = cloneDBEnumOptions(parsed.Enum.Options)
	column.Metadata["enum_kind"] = parsed.Enum.Kind
	column.Metadata["enum_mode"] = parsed.Enum.Mode
	column.Metadata["enum_display"] = parsed.Enum.Display
	column.Metadata["enum_source"] = parsed.Enum.Source
	if parsed.Enum.Ref != "" {
		column.Metadata["enum_source_ref"] = parsed.Enum.Ref
	}
	if len(parsed.Enum.Values) > 0 {
		column.Metadata["enum_values"] = append([]string(nil), parsed.Enum.Values...)
	}
	if len(parsed.Enum.Options) > 0 {
		column.Metadata["enum_options"] = cloneEnumOptionMetadata(parsed.Enum.Options)
	}
}

func parseEnumComment(comment string) enumParseResult {
	parsed := parseColumnCommentMetadata(comment)
	if !parsed.Enum.OK {
		return enumParseResult{}
	}
	return parsed.Enum
}

type columnCommentMetadata struct {
	Label  string
	UIType string
	Enum   enumParseResult
}

func parseColumnCommentMetadata(comment string) columnCommentMetadata {
	text := strings.TrimSpace(comment)
	if text == "" {
		return columnCommentMetadata{}
	}
	label := text
	annotations := ""
	if left, right, ok := strings.Cut(text, "|"); ok {
		label = strings.TrimSpace(left)
		annotations = strings.TrimSpace(right)
	}
	result := columnCommentMetadata{Label: label}
	if annotations == "" {
		return result
	}
	inner := annotations
	if strings.HasPrefix(inner, "(") && strings.HasSuffix(inner, ")") {
		inner = strings.TrimSpace(inner[1 : len(inner)-1])
	}
	parseSemiCommentAnnotations(inner, &result)
	if !result.Enum.OK && strings.HasPrefix(annotations, "(") && strings.HasSuffix(annotations, ")") {
		parseParenCommentAnnotations(inner, &result)
	}
	if !result.Enum.OK {
		if enum := parseBareEnumDefinition(inner); enum.OK {
			result.Enum = enum
		}
	}
	if result.Enum.OK {
		if result.UIType == "" {
			result.UIType = "select"
		}
		if result.UIType != "" {
			result.Enum.Display = result.UIType
		}
	}
	return result
}

func parseBareEnumDefinition(text string) enumParseResult {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return enumParseResult{}
	}
	if !strings.Contains(trimmed, ",") {
		return enumParseResult{}
	}
	if strings.ContainsAny(trimmed, ";:()") {
		return enumParseResult{}
	}
	return parseEnumDefinition(trimmed)
}

func parseSemiCommentAnnotations(text string, result *columnCommentMetadata) {
	parts := strings.Split(text, ";")
	for _, part := range parts {
		if key, value, ok := strings.Cut(strings.TrimSpace(part), "="); ok {
			parseCommentAnnotationValue(key, value, result)
		}
	}
}

func parseParenCommentAnnotations(text string, result *columnCommentMetadata) {
	parts := strings.Split(text, ",")
	for i := 0; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, ":")
		if !ok {
			continue
		}
		if normalizeCommentAnnotationKey(key) == "enum" {
			if i < len(parts)-1 {
				value = value + "," + strings.Join(trimmedParts(parts[i+1:]), ",")
				i = len(parts)
			}
			parseCommentAnnotationValue("enum", value, result)
			continue
		}
		parseCommentAnnotationValue(key, value, result)
	}
}

func trimmedParts(parts []string) []string {
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func parseCommentAnnotationValue(key string, value string, result *columnCommentMetadata) {
	switch normalizeCommentAnnotationKey(key) {
	case "ui", "ui_type":
		if ui := codeschema.NormalizeUIType(value); ui != "" {
			result.UIType = ui
		}
	case "enum":
		if enum := parseEnumDefinition(value); enum.OK {
			result.Enum = enum
		}
	}
}

func normalizeCommentAnnotationKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

func parseEnumDefinition(spec string) enumParseResult {
	text := strings.TrimSpace(spec)
	if text == "" {
		return enumParseResult{}
	}
	parts := splitEnumCommentParts(text)
	if len(parts) == 0 {
		return enumParseResult{}
	}
	result := enumParseResult{Kind: "comment", Mode: "single", Display: "select", Source: "comment", OK: true}
	allValueLabel := true
	for i, part := range parts {
		option, ok := parseEnumCommentOption(part)
		if !ok {
			return enumParseResult{}
		}
		if option.Label != option.Value {
			allValueLabel = false
		}
		if option.Order == 0 {
			option.Order = i + 1
		}
		result.Options = append(result.Options, option)
	}
	if len(result.Options) == 0 {
		return enumParseResult{}
	}
	result.Values = enumValuesFromDBOptions(result.Options)
	if !allValueLabel {
		result.Kind = "comment-mapped"
	}
	return result
}

func splitEnumCommentParts(text string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	parts := strings.Split(text, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func parseEnumCommentOption(text string) (dbschema.EnumOption, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return dbschema.EnumOption{}, false
	}
	left, right, ok := strings.Cut(trimmed, "=")
	if !ok {
		return dbschema.EnumOption{}, false
	}
	value := strings.TrimSpace(left)
	label := strings.TrimSpace(right)
	if value == "" || label == "" {
		return dbschema.EnumOption{}, false
	}
	return dbschema.EnumOption{Value: value, Label: label}, true
}

func enumValuesFromDBOptions(options []dbschema.EnumOption) []string {
	if len(options) == 0 {
		return nil
	}
	values := make([]string, 0, len(options))
	for _, option := range options {
		value := strings.TrimSpace(option.Value)
		if value == "" {
			continue
		}
		values = append(values, value)
	}
	return values
}

func cloneDBEnumOptions(options []dbschema.EnumOption) []dbschema.EnumOption {
	if len(options) == 0 {
		return nil
	}
	cloned := make([]dbschema.EnumOption, 0, len(options))
	for _, option := range options {
		clone := option
		if option.Metadata != nil {
			clone.Metadata = make(map[string]any, len(option.Metadata))
			for key, value := range option.Metadata {
				clone.Metadata[key] = value
			}
		}
		cloned = append(cloned, clone)
	}
	return cloned
}

func cloneEnumOptionMetadata(options []dbschema.EnumOption) []map[string]any {
	if len(options) == 0 {
		return nil
	}
	result := make([]map[string]any, 0, len(options))
	for _, option := range options {
		metadata := map[string]any{
			"value":    option.Value,
			"label":    option.Label,
			"color":    option.Color,
			"disabled": option.Disabled,
			"order":    option.Order,
		}
		if option.Metadata != nil {
			for key, value := range option.Metadata {
				metadata[key] = value
			}
		}
		result = append(result, metadata)
	}
	return result
}

func markAutoIncrementColumns(columns []dbschema.Column, autoIncrementColumns []string) {
	if len(columns) == 0 || len(autoIncrementColumns) == 0 {
		return
	}
	allowed := make(map[string]struct{}, len(autoIncrementColumns))
	for _, name := range autoIncrementColumns {
		allowed[strings.ToLower(strings.TrimSpace(name))] = struct{}{}
	}
	for idx := range columns {
		if _, ok := allowed[strings.ToLower(strings.TrimSpace(columns[idx].Name))]; ok {
			columns[idx].AutoIncrement = true
		}
	}
}

func applyIndexFlags(columns []dbschema.Column, indexes []dbschema.Index) {
	if len(columns) == 0 || len(indexes) == 0 {
		return
	}
	lookup := make(map[string]*dbschema.Column, len(columns))
	for idx := range columns {
		column := &columns[idx]
		lookup[strings.ToLower(strings.TrimSpace(column.Name))] = column
	}
	for _, index := range indexes {
		for _, columnName := range index.Columns {
			if column, ok := lookup[strings.ToLower(strings.TrimSpace(columnName))]; ok {
				column.Index = true
				if index.Unique && len(index.Columns) == 1 {
					column.Unique = true
				}
			}
		}
	}
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(strings.TrimSpace(value), `"`, `""`) + `"`
}

func normalizeDriverKind(name string) dbschema.DriverKind {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "mysql":
		return dbschema.DriverKindMySQL
	case "postgres", "postgresql":
		return dbschema.DriverKindPostgreSQL
	case "sqlite":
		return dbschema.DriverKindSQLite
	case "sqlserver", "mssql":
		return dbschema.DriverKindSQLServer
	default:
		return dbschema.DriverKindUnknown
	}
}
