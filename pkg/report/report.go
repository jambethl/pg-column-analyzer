package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"

	"main/pkg/common"
)

func GenerateReport(columnList []common.ColumnInfo, tableName string) error {
	reportName := fmt.Sprintf("reports/%s_report.csv", tableName)
	file, err := os.Create(reportName)
	if err != nil {
		return fmt.Errorf("unable to create report: %s, error: %v", reportName, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writeCSVHeader(writer); err != nil {
		return fmt.Errorf("unable to write CSV header: %v", err)
	}

	// Build a map here to store the recommended position as the map's value.
	// We could sort the list of columns by their type alignment, but retrieving
	// the given column would take O(n) since we don't know its position.
	alignmentMap := buildAlignmentMap(columnList)

	// Write current column order with padding information
	for i, col := range columnList {
		currentSize := col.TypAlign
		var nextSize int
		if i < len(columnList)-1 {
			nextSize = columnList[i+1].TypAlign
		}

		wastedPadding := calculateWastedPadding(currentSize, nextSize)

		row := []string{
			strconv.Itoa(col.OrdinalPosition),
			col.ColumnName,
			col.DataType,
			col.IsNullable,
			strconv.Itoa(col.TypLen),
			strconv.Itoa(col.TypAlign),
			strconv.Itoa(wastedPadding),
			strconv.Itoa(alignmentMap[col.ColumnName]),
			strconv.Itoa(col.EntryCount * wastedPadding),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("unable to write CSV row: %v", err)
		}
	}

	fmt.Printf("Report %s generated successfully.\n", reportName)
	return nil
}

func writeCSVHeader(writer *csv.Writer) error {
	return writer.Write([]string{
		"Ordinal Position",
		"Column Name",
		"Data Type",
		"Nullable",
		"Data Type Size (B)",
		"Type Alignment (B)",
		"Wasted Padding Per Entry (B)",
		"Recommended Position",
		"Total Wasted Space (B)",
	})
}

func calculateWastedPadding(currentSize, nextSize int) int {
	if nextSize == 0 {
		return 0
	}

	remainder := currentSize % nextSize
	if remainder == 0 {
		return 0
	}

	return nextSize - remainder
}

func buildAlignmentMap(columnList []common.ColumnInfo) map[string]int {
	copiedList := make([]common.ColumnInfo, len(columnList))
	copy(copiedList, columnList)

	sort.Slice(copiedList, func(i, j int) bool {
		return copiedList[i].TypAlign > copiedList[j].TypAlign
	})

	columnMap := make(map[string]int)

	for i, colInfo := range copiedList {
		columnMap[colInfo.ColumnName] = i + 1
	}

	return columnMap
}
