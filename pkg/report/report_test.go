package report

import (
	"encoding/csv"
	"fmt"
	"main/pkg/types"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateReport(t *testing.T) {
	// Create a temporary directory for the report
	tmpDir := t.TempDir()
	reportDir := filepath.Join(tmpDir, "reports")
	if err := os.Mkdir(reportDir, 0755); err != nil {
		t.Fatalf("Failed to create reports directory: %v", err)
	}

	// Mock column data
	columnList := []types.ColumnInfo{
		{OrdinalPosition: 1, ColumnName: "id", DataType: "integer", IsNullable: "NO"},
		{OrdinalPosition: 2, ColumnName: "name", DataType: "varchar", IsNullable: "YES"},
		{OrdinalPosition: 3, ColumnName: "created_at", DataType: "timestamp with time zone", IsNullable: "NO"},
	}

	// Set the current working directory to the temporary directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(origDir) // Restore original directory after test
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory to temporary directory: %v", err)
	}

	// Call GenerateReport
	tableName := "test_table"
	err = GenerateReport(columnList, tableName)
	if err != nil {
		t.Fatalf("GenerateReport failed: %v", err)
	}

	// Verify the report file exists
	reportPath := filepath.Join(reportDir, fmt.Sprintf("%s_report.csv", tableName))
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		t.Fatalf("Report file does not exist: %s", reportPath)
	}

	// Open and read the report file
	file, err := os.Open(reportPath)
	if err != nil {
		t.Fatalf("Failed to open report file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	// Verify the content of the report
	expectedHeader := []string{"Ordinal Position", "Column Name", "Data Type", "Nullable", "Data Type Size (B)", "Wasted Padding", "Recommended Position"}
	if !equalSlices(rows[0], expectedHeader) {
		t.Errorf("Header row mismatch. Expected %v, got %v", expectedHeader, rows[0])
	}

	expectedRows := [][]string{
		{"1", "id", "integer", "NO", "4 bytes", "", ""},
		{"2", "name", "varchar", "YES", "10 bytes", "", ""},
		{"3", "created_at", "timestamp with time zone", "NO", "8 bytes", "", ""},
	}

	for i, expectedRow := range expectedRows {
		if !equalSlices(rows[i+1], expectedRow) {
			t.Errorf("Row %d mismatch. Expected %v, got %v", i+1, expectedRow, rows[i+1])
		}
	}
}

// Helper function to compare two slices
func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
