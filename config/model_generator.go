package config

import (
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

// GenerateModelFromTable generates a Go struct definition from a database table
func GenerateModelFromTable(db *gorm.DB, tableName string) (string, error) {
	// Get column types using raw SQL to describe the table
	rows, err := db.Raw(fmt.Sprintf("DESCRIBE %s", tableName)).Rows()
	if err != nil {
		return "", fmt.Errorf("failed to describe table %s: %v", tableName, err)
	}
	defer rows.Close()

	var fields []string
	imports := make(map[string]bool)

	structName := toPascalCase(tableName)
	structName = strings.TrimSuffix(structName, "s") // Remove trailing 's' for singular model name

	// Track standard gorm.Model fields separately
	var standardFields []string
	var customFields []string

	// Always need time import for standard fields
	imports["time"] = true

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

		// Track required imports
		if strings.Contains(goType, "time.Time") {
			imports["time"] = true
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
			continue
		}
		if columnName == "updated_at" {
			standardFields = append(standardFields, "\tUpdatedAt time.Time      `json:\"updated_at\"`")
			continue
		}
		if columnName == "deleted_at" {
			standardFields = append(standardFields, "\tDeletedAt gorm.DeletedAt `json:\"deleted_at,omitempty\" gorm:\"index\"`")
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

	// Add imports
	if len(imports) > 0 || len(fields) > 0 {
		structBuilder.WriteString("import (\n")
		structBuilder.WriteString("\t\"gorm.io/gorm\"\n")
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

		structCode, err := GenerateModelFromTable(db, tableName)
		if err != nil {
			log.Printf("Error generating model for table %s: %v\n", tableName, err)
			continue
		}

		// Determine file name
		structName := toPascalCase(tableName)
		structName = strings.TrimSuffix(structName, "s")
		fileName := fmt.Sprintf("%s.go", structName)
		filePath := filepath.Join(modelsDir, fileName)

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
