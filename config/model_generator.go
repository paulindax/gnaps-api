package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

// convertToString converts interface{} to string, handling both string and []byte
func convertToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// DatabaseToGoTypeMapper maps database types to Go types
func DatabaseToGoTypeMapper(dbType string) string {
	dbType = strings.ToUpper(dbType)

	// Handle common database types
	switch {
	case strings.Contains(dbType, "INT"):
		if strings.Contains(dbType, "BIGINT") {
			return "int64"
		}
		if strings.Contains(dbType, "TINYINT(1)") {
			return "bool"
		}
		return "int"
	case strings.Contains(dbType, "VARCHAR"), strings.Contains(dbType, "TEXT"), strings.Contains(dbType, "CHAR"):
		return "string"
	case strings.Contains(dbType, "DECIMAL"), strings.Contains(dbType, "NUMERIC"):
		return "float64"
	case strings.Contains(dbType, "FLOAT"), strings.Contains(dbType, "DOUBLE"):
		return "float64"
	case strings.Contains(dbType, "BOOL"):
		return "bool"
	case strings.Contains(dbType, "TIMESTAMP"), strings.Contains(dbType, "DATETIME"), strings.Contains(dbType, "DATE"):
		return "time.Time"
	case strings.Contains(dbType, "JSON"):
		return "datatypes.JSON"
	case strings.Contains(dbType, "BLOB"), strings.Contains(dbType, "BINARY"):
		return "[]byte"
	default:
		return "interface{}"
	}
}

// GenerateStructField generates a Go struct field from column information
func GenerateStructField(columnName, goType string, nullable bool) string {
	// Convert column name to PascalCase for field name
	fieldName := toPascalCase(columnName)

	// Handle nullable fields with pointers
	if nullable && goType != "time.Time" {
		goType = "*" + goType
	}

	// Generate struct tags
	jsonTag := toSnakeCase(columnName)
	gormTag := fmt.Sprintf("column:%s", columnName)

	return fmt.Sprintf("\t%s %s `json:\"%s\" gorm:\"%s\"`", fieldName, goType, jsonTag, gormTag)
}

// toPascalCase converts snake_case to PascalCase
func toPascalCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

// toSnakeCase converts PascalCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// extractTransientFields reads an existing model file and extracts transient fields
func extractTransientFields(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		// File doesn't exist, no transient fields to preserve
		return nil, nil
	}
	defer file.Close()

	var transientFields []string
	var inTransientSection bool
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Start capturing when we see the transient fields comment
		if strings.Contains(trimmed, "// Transient fields") {
			inTransientSection = true
			transientFields = append(transientFields, line)
			continue
		}

		// Stop capturing when we hit the closing brace of the struct
		if inTransientSection && trimmed == "}" {
			break
		}

		// Capture lines in the transient section
		if inTransientSection {
			transientFields = append(transientFields, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return transientFields, nil
}

// GenerateModelFromTable generates a Go struct definition from a database table
func GenerateModelFromTable(db *gorm.DB, tableName string, existingFilePath string) (string, error) {
	// Extract transient fields from existing file if it exists
	transientFields, err := extractTransientFields(existingFilePath)
	if err != nil {
		log.Printf("Warning: Could not extract transient fields from %s: %v", existingFilePath, err)
	}
	// Get column types using raw SQL to describe the table
	rows, err := db.Raw(fmt.Sprintf("DESCRIBE %s", tableName)).Rows()
	if err != nil {
		return "", fmt.Errorf("failed to describe table %s: %v", tableName, err)
	}
	defer rows.Close()

	var fields []string
	imports := make(map[string]bool)
	needsGorm := false // Track if we need gorm import
	needsTime := false // Track if we need time import

	structName := toPascalCase(tableName)
	structName = strings.TrimSuffix(structName, "s") // Remove trailing 's' for singular model name

	// Track standard gorm.Model fields separately
	var standardFields []string
	var customFields []string

	// Parse column information
	for rows.Next() {
		var field, colType, null, key, def, extra any
		if err := rows.Scan(&field, &colType, &null, &key, &def, &extra); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		columnName := convertToString(field)
		dbType := convertToString(colType)
		nullable := convertToString(null) == "YES"

		goType := DatabaseToGoTypeMapper(dbType)

		// Track required imports based on actual field types
		if strings.Contains(goType, "time.Time") {
			needsTime = true
		}
		if strings.Contains(goType, "datatypes.JSON") {
			imports["gorm.io/datatypes"] = true
		}

		// Handle standard gorm.Model fields explicitly with proper JSON tags
		if columnName == "id" {
			standardFields = append(standardFields, "\tID        uint           `json:\"id\" gorm:\"primarykey\"`")
			continue
		}
		if columnName == "created_at" {
			standardFields = append(standardFields, "\tCreatedAt time.Time      `json:\"created_at\"`")
			needsTime = true
			continue
		}
		if columnName == "updated_at" {
			standardFields = append(standardFields, "\tUpdatedAt time.Time      `json:\"updated_at\"`")
			needsTime = true
			continue
		}
		if columnName == "deleted_at" {
			standardFields = append(standardFields, "\tDeletedAt gorm.DeletedAt `json:\"deleted_at,omitempty\" gorm:\"index\"`")
			needsGorm = true // We need gorm import for gorm.DeletedAt
			needsTime = true
			continue
		}

		fieldDef := GenerateStructField(columnName, goType, nullable)
		customFields = append(customFields, fieldDef)
	}

	// Combine standard fields first, then custom fields
	fields = append(standardFields, customFields...)

	// Build the complete struct definition
	var structBuilder strings.Builder

	structBuilder.WriteString("package models\n\n")

	// Add imports only if needed
	if needsGorm || needsTime || len(imports) > 0 {
		structBuilder.WriteString("import (\n")
		// Only import gorm if we have deleted_at field
		if needsGorm {
			structBuilder.WriteString("\t\"gorm.io/gorm\"\n")
		}
		// Only import time if we have time.Time fields
		if needsTime {
			structBuilder.WriteString("\t\"time\"\n")
		}
		// Import other packages
		for imp := range imports {
			structBuilder.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
		}
		structBuilder.WriteString(")\n\n")
	}

	// Add struct definition
	structBuilder.WriteString(fmt.Sprintf("// %s model generated from database table '%s'\n", structName, tableName))
	structBuilder.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	// Write fields with proper spacing
	hasStandardFields := false
	for i, field := range fields {
		structBuilder.WriteString(field + "\n")
		// Add blank line after standard fields (first 4 fields if present)
		if i == 3 && len(fields) > 4 {
			structBuilder.WriteString("\n")
			hasStandardFields = true
		}
	}

	// If we don't have standard fields but have custom fields, no extra line needed
	if !hasStandardFields && len(fields) > 0 {
		// Fields already written, no extra spacing needed
	}

	// Add preserved transient fields if any exist
	if len(transientFields) > 0 {
		structBuilder.WriteString("\n")
		for _, field := range transientFields {
			structBuilder.WriteString(field + "\n")
		}
	}

	structBuilder.WriteString("}\n\n")

	// Add TableName method
	structBuilder.WriteString(fmt.Sprintf("func (%s) TableName() string {\n", structName))
	structBuilder.WriteString(fmt.Sprintf("\treturn \"%s\"\n", tableName))
	structBuilder.WriteString("}\n")

	return structBuilder.String(), nil
}

// GenerateModelsMiddleware scans database tables and generates model files
func GenerateModelsMiddleware(db *gorm.DB, modelsDir string) error {
	log.Println("=== Starting Model Generation from Database ===")

	// Ensure models directory exists
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return fmt.Errorf("failed to create models directory: %v", err)
	}

	// Get list of tables from database
	var tables []string
	rows, err := db.Raw("SHOW TABLES").Rows()
	if err != nil {
		return fmt.Errorf("failed to get tables: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Printf("Error scanning table name: %v", err)
			continue
		}
		tables = append(tables, tableName)
	}

	log.Printf("Found %d tables in database\n", len(tables))

	// Generate model for each table
	for _, tableName := range tables {
		log.Printf("Generating model for table: %s\n", tableName)

		// Determine file name and path first
		structName := toPascalCase(tableName)
		structName = strings.TrimSuffix(structName, "s")
		fileName := fmt.Sprintf("%s.go", structName)
		filePath := filepath.Join(modelsDir, fileName)

		// Generate model, preserving any existing transient fields
		structCode, err := GenerateModelFromTable(db, tableName, filePath)
		if err != nil {
			log.Printf("Error generating model for table %s: %v\n", tableName, err)
			continue
		}

		// Write to file
		if err := os.WriteFile(filePath, []byte(structCode), 0644); err != nil {
			log.Printf("Error writing model file %s: %v\n", fileName, err)
			continue
		}

		log.Printf("Successfully generated model: %s\n", filePath)
	}

	log.Println("=== Model Generation Complete ===")
	return nil
}
