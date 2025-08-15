package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"bome-backend/internal/database"

	"github.com/gin-gonic/gin"
)

// DatabaseExportHandler handles database export requests from admin dashboard
func DatabaseExportHandler(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "Database not available",
			})
			return
		}

		format := c.DefaultQuery("format", "sql")
		includeData := c.DefaultQuery("include_data", "true") == "true"

		// Generate timestamp for filename
		timestamp := time.Now().Format("2006-01-02_15-04-05")

		var filename string
		var contentType string

		switch format {
		case "sql":
			filename = fmt.Sprintf("bome_database_export_%s.sql", timestamp)
			contentType = "application/sql"
		case "json":
			filename = fmt.Sprintf("bome_database_export_%s.json", timestamp)
			contentType = "application/json"
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Unsupported format. Use 'sql' or 'json'",
			})
			return
		}

		// Set headers for file download
		c.Header("Content-Type", contentType)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

		if format == "sql" {
			// Generate SQL dump
			if err := generateSQLDump(db, c, includeData); err != nil {
				log.Printf("Failed to generate SQL dump: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to generate database export",
				})
				return
			}
		} else {
			// Generate JSON export
			if err := generateJSONExport(db, c, includeData); err != nil {
				log.Printf("Failed to generate JSON export: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to generate database export",
				})
				return
			}
		}

		// Log admin action
		adminID := c.GetInt("user_id")
		go db.CreateAdminLog(&adminID, "database_exported", "database", nil, map[string]interface{}{
			"format":       format,
			"include_data": includeData,
			"filename":     filename,
		}, c.ClientIP(), c.GetHeader("User-Agent"))
	}
}

// generateSQLDump creates a PostgreSQL dump using docker exec
func generateSQLDump(db *database.DB, c *gin.Context, includeData bool) error {
	// Use docker exec to run pg_dump inside the postgres container
	args := []string{
		"exec", "bome_postgres", "pg_dump",
		"-U", "bome_user",
		"-d", "bome",
	}

	if !includeData {
		args = append(args, "--schema-only")
	}

	// Execute docker command
	cmd := exec.Command("docker", args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	c.DataFromReader(http.StatusOK, -1, "application/sql", stdout, nil)
	return cmd.Wait()
}

// generateJSONExport creates a JSON export of the database
func generateJSONExport(db *database.DB, c *gin.Context, includeData bool) error {
	export := make(map[string]interface{})

	// Add metadata
	export["export_info"] = map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"format":       "json",
		"include_data": includeData,
		"version":      "1.0",
		"database":     "bome",
	}

	if includeData {
		// Export all table data
		tables := []string{"users", "videos", "subscriptions", "audit_logs", "sessions", "roles", "departments"}

		for _, table := range tables {
			data, err := exportTableData(db, table)
			if err != nil {
				log.Printf("Failed to export table %s: %v", table, err)
				continue
			}
			export[table] = data
		}
	} else {
		// Export only schema information
		schema, err := exportSchema(db)
		if err != nil {
			return fmt.Errorf("failed to export schema: %w", err)
		}
		export["schema"] = schema
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	c.Data(http.StatusOK, "application/json", jsonData)
	return nil
}

// exportTableData exports data from a specific table
func exportTableData(db *database.DB, tableName string) ([]map[string]interface{}, error) {
	// This is a simplified version - you'd want to implement proper table exports
	// based on your actual database schema

	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		// Scan the result into the column pointers...
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Create our map and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		entry := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Convert []byte to string for better JSON representation
			if b, ok := val.([]byte); ok {
				entry[col] = string(b)
			} else {
				entry[col] = val
			}
		}

		result = append(result, entry)
	}

	return result, nil
}

// exportSchema exports database schema information
func exportSchema(db *database.DB) (map[string]interface{}, error) {
	schema := make(map[string]interface{})

	// Get table information
	query := `
		SELECT table_name, column_name, data_type, is_nullable, column_default
		FROM information_schema.columns 
		WHERE table_schema = 'public'
		ORDER BY table_name, ordinal_position
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tables := make(map[string][]map[string]interface{})

	for rows.Next() {
		var tableName, columnName, dataType, isNullable string
		var columnDefault sql.NullString

		if err := rows.Scan(&tableName, &columnName, &dataType, &isNullable, &columnDefault); err != nil {
			continue
		}

		column := map[string]interface{}{
			"name":     columnName,
			"type":     dataType,
			"nullable": isNullable == "YES",
		}

		if columnDefault.Valid {
			column["default"] = columnDefault.String
		}

		tables[tableName] = append(tables[tableName], column)
	}

	schema["tables"] = tables
	return schema, nil
}
