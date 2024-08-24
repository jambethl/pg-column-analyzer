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

	sortedColumnList := sortColumnsBySize(columnList)

	// Write current column order with padding information
	for i, col := range columnList {
		currentSize := col.TypAlign
		var nextSize int
		if i < len(columnList)-1 {
			nextSize = columnList[i+1].TypAlign
		}

		wastedPadding := calculateWastedPadding(currentSize, nextSize)
		recommendedPosition := findRecommendedPosition(col.ColumnName, sortedColumnList)

		row := []string{
			strconv.Itoa(col.OrdinalPosition),
			col.ColumnName,
			col.DataType,
			col.IsNullable,
			strconv.Itoa(col.TypLen),
			strconv.Itoa(wastedPadding),
			strconv.Itoa(recommendedPosition),
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
		"Wasted Padding Per Entry",
		"Recommended Position",
		"Total Wasted Space",
	})
}

func sortColumnsBySize(columnList []common.ColumnInfo) []common.ColumnInfo {
	sort.SliceStable(columnList, func(i, j int) bool {
		return columnList[i].TypLen > columnList[j].TypLen
	})

	return columnList
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

func findRecommendedPosition(columnName string, sortedColumnList []common.ColumnInfo) int {
	for i, col := range sortedColumnList {
		if col.ColumnName == columnName {
			return i + 1
		}
	}
	return -1
}
