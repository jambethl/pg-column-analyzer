package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"

	"main/pkg/common"
)

// This stores a map of fixed-length Postgres data types and their column padding
var dataTypeMap = map[string]int{
	// TODO - bit(n); determine bit length by value of n
	"boolean":                     1,
	"char":                        1,
	"macaddr":                     1,
	"smallint":                    2,
	"smallserial":                 2,
	"date":                        4,
	"integer":                     4,
	"real":                        4,
	"serial":                      4,
	"bigint":                      8,
	"bigserial":                   8,
	"double precision":            8,
	"interval":                    8,
	"money":                       8,
	"point":                       8,
	"uuid":                        8,
	"time without time zone":      8,
	"timestamp":                   8,
	"timestamp without time zone": 8,
	"timestamp with time zone":    8,
}

// This stores variable-length Postgres data types, where we ignore column padding
var variableLengthDataTypes = map[string]struct{}{
	"bytea":   {},
	"text":    {},
	"varchar": {},
}

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
		currentSize := dataTypeMap[col.DataType]
		var nextSize int
		if i < len(columnList)-1 {
			nextSize = dataTypeMap[columnList[i+1].DataType]
		}

		wastedPadding := calculateWastedPadding(currentSize, nextSize)
		recommendedPosition := findRecommendedPosition(col.ColumnName, sortedColumnList)

		row := []string{
			strconv.Itoa(col.OrdinalPosition),
			col.ColumnName,
			col.DataType,
			col.IsNullable,
			strconv.Itoa(currentSize),
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
	var fixedLengthColumns, variableLengthColumns []common.ColumnInfo
	for _, col := range columnList {
		_, exists := variableLengthDataTypes[col.DataType]
		if exists {
			variableLengthColumns = append(variableLengthColumns, col)
		} else {
			fixedLengthColumns = append(fixedLengthColumns, col)
		}
	}
	sort.SliceStable(fixedLengthColumns, func(i, j int) bool {
		return dataTypeMap[fixedLengthColumns[i].DataType] > dataTypeMap[fixedLengthColumns[j].DataType]
	})

	sortedColumnList := append(fixedLengthColumns, variableLengthColumns...)
	return sortedColumnList
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
