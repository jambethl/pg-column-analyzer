package report

import (
	"encoding/csv"
	"fmt"
	"main/pkg/types"
	"os"
	"path/filepath"
	"testing"
)

var expectedHeader = []string{"Ordinal Position", "Column Name", "Data Type", "Nullable", "Data Type Size (B)", "Wasted Padding", "Recommended Position"}

func TestGenerateReport(t *testing.T) {
	columnList := []types.ColumnInfo{
		{OrdinalPosition: 1, ColumnName: "enabled", DataType: "boolean", IsNullable: "NO"},
		{OrdinalPosition: 2, ColumnName: "age", DataType: "smallint", IsNullable: "NO"},
		{OrdinalPosition: 3, ColumnName: "count", DataType: "integer", IsNullable: "NO"},
		{OrdinalPosition: 4, ColumnName: "id", DataType: "bigint", IsNullable: "NO"},
	}

	generateReportTest(t, columnList, [][]string{
		{"1", "enabled", "boolean", "NO", "1", "1", "4"},
		{"2", "age", "smallint", "NO", "2", "2", "3"},
		{"3", "count", "integer", "NO", "4", "4", "2"},
		{"4", "id", "bigint", "NO", "8", "0", "1"},
	})
}

func TestGenerateReport_NullableColumns(t *testing.T) {
	columnList := []types.ColumnInfo{
		{OrdinalPosition: 1, ColumnName: "description", DataType: "text", IsNullable: "YES"},
		{OrdinalPosition: 2, ColumnName: "price", DataType: "real", IsNullable: "YES"},
	}

	generateReportTest(t, columnList, [][]string{
		{"1", "description", "text", "YES", "10", "2", "1"},
		{"2", "price", "real", "YES", "4", "0", "2"},
	})
}

func TestGenerateReport_SameDataType(t *testing.T) {
	columnList := []types.ColumnInfo{
		{OrdinalPosition: 1, ColumnName: "first_name", DataType: "varchar", IsNullable: "YES"},
		{OrdinalPosition: 2, ColumnName: "last_name", DataType: "varchar", IsNullable: "YES"},
	}

	generateReportTest(t, columnList, [][]string{
		{"1", "first_name", "varchar", "YES", "10", "0", "1"},
		{"2", "last_name", "varchar", "YES", "10", "0", "2"},
	})
}

func TestGenerateReport_SOExample(t *testing.T) {
	columnList := []types.ColumnInfo{
		{OrdinalPosition: 1, ColumnName: "e", DataType: "smallint", IsNullable: "NO"},
		{OrdinalPosition: 2, ColumnName: "a", DataType: "bigint", IsNullable: "NO"},
		{OrdinalPosition: 3, ColumnName: "f", DataType: "smallint", IsNullable: "NO"},
		{OrdinalPosition: 4, ColumnName: "b", DataType: "bigint", IsNullable: "NO"},
		{OrdinalPosition: 5, ColumnName: "g", DataType: "smallint", IsNullable: "NO"},
		{OrdinalPosition: 6, ColumnName: "c", DataType: "bigint", IsNullable: "NO"},
		{OrdinalPosition: 7, ColumnName: "h", DataType: "smallint", IsNullable: "NO"},
		{OrdinalPosition: 8, ColumnName: "d", DataType: "bigint", IsNullable: "NO"},
	}

	generateReportTest(t, columnList, [][]string{
		{"1", "e", "smallint", "NO", "2", "6", "5"},
		{"2", "a", "bigint", "NO", "8", "0", "1"},
		{"3", "f", "smallint", "NO", "2", "6", "6"},
		{"4", "b", "bigint", "NO", "8", "0", "2"},
		{"5", "g", "smallint", "NO", "2", "6", "7"},
		{"6", "c", "bigint", "NO", "8", "0", "3"},
		{"7", "h", "smallint", "NO", "2", "6", "8"},
		{"8", "d", "bigint", "NO", "8", "0", "4"},
	})
}

func TestGenerateReport_AllDataTypes(t *testing.T) {
	columnList := []types.ColumnInfo{
		{OrdinalPosition: 1, ColumnName: "id", DataType: "smallint", IsNullable: "NO"},
		{OrdinalPosition: 2, ColumnName: "status", DataType: "boolean", IsNullable: "NO"},
		{OrdinalPosition: 3, ColumnName: "created_at", DataType: "timestamp without time zone", IsNullable: "YES"},
		{OrdinalPosition: 4, ColumnName: "score", DataType: "double precision", IsNullable: "YES"},
		{OrdinalPosition: 5, ColumnName: "unique_id", DataType: "uuid", IsNullable: "NO"},
		{OrdinalPosition: 6, ColumnName: "data", DataType: "bytea", IsNullable: "YES"},
	}

	generateReportTest(t, columnList, [][]string{
		{"1", "id", "smallint", "NO", "2", "0", "4"},
		{"2", "status", "boolean", "NO", "1", "7", "6"},
		{"3", "created_at", "timestamp without time zone", "NO", "8", "0", "2"},
		{"4", "score", "double precision", "YES", "8", "8", "3"},
		{"5", "unique_id", "uuid", "YES", "16", "4", "1"},
		{"6", "data", "bytea", "YES", "10", "0", "5"},
	})
}

func TestGenerateReport_SingleColumn(t *testing.T) {
	columnList := []types.ColumnInfo{
		{OrdinalPosition: 1, ColumnName: "id", DataType: "uuid", IsNullable: "NO"},
	}

	generateReportTest(t, columnList, [][]string{
		{"1", "id", "uuid", "NO", "16", "0", "1"},
	})
}

func generateReportTest(t *testing.T, columnList []types.ColumnInfo, expected [][]string) {
	tmpDir := t.TempDir()
	reportDir := createReportsDirectory(t, tmpDir)

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
	rows := readFile(t, tableName, reportDir)

	assertResult(t, rows, expected)
}

func assertResult(t *testing.T, rows [][]string, expectedRows [][]string) {
	// Verify the content of the report
	if !equalSlices(rows[0], expectedHeader) {
		t.Errorf("Header row mismatch. Expected %v, got %v", expectedHeader, rows[0])
	}

	for i, expectedRow := range expectedRows {
		if !equalSlices(rows[i+1], expectedRow) {
			t.Errorf("Row %d mismatch. Expected %v, got %v", i+1, expectedRow, rows[i+1])
		}
	}
}

func createReportsDirectory(t *testing.T, tmpDir string) string {
	// Create a temporary directory for the report
	reportDir := filepath.Join(tmpDir, "reports")
	if err := os.Mkdir(reportDir, 0755); err != nil {
		t.Fatalf("Failed to create reports directory: %v", err)
	}
	return reportDir
}

func readFile(t *testing.T, tableName string, reportDir string) [][]string {
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
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	return rows
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
