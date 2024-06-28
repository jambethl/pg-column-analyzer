package report

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"

	"main/pkg/common"
)

var dataTypeMap = map[string]int{
	"boolean":                     1,
	"smallint":                    2,
	"date":                        4,
	"integer":                     4,
	"real":                        4,
	"bigint":                      8,
	"double precision":            8,
	"timestamp without time zone": 8,
	"timestamp with time zone":    8,
	"bytea":                       10, // Variable size, often larger
	"text":                        10, // Variable size, often larger
	"varchar":                     10, // Variable size, often larger
	"uuid":                        16,
}

func GenerateReport(columnList []common.ColumnInfo, tableName string) error {
	reportName := fmt.Sprintf("reports/%s_report.csv", tableName)
	file, err := os.Create(reportName)
	if err != nil {
		log.Fatalf("Unable to create report: %s", reportName)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Ordinal Position", "Column Name", "Data Type", "Nullable", "Data Type Size (B)", "Wasted Padding Per Entry", "Recommended Position", "Total Wasted Space"})

	// Sort columns to recommend optimal order
	sortedColumnList := make([]common.ColumnInfo, len(columnList))
	copy(sortedColumnList, columnList)
	sort.SliceStable(sortedColumnList, func(i, j int) bool {
		sizeI := dataTypeMap[sortedColumnList[i].DataType]
		sizeJ := dataTypeMap[sortedColumnList[j].DataType]
		return sizeI > sizeJ // Larger sizes come first
	})

	// Write current column order with padding information
	for i, col := range columnList {
		currentSize := dataTypeMap[col.DataType]
		var nextSize int
		if i < len(columnList)-1 {
			nextSize = dataTypeMap[columnList[i+1].DataType]
		} else {
			nextSize = 0
		}
		wastedPadding := calculateWastedPadding(currentSize, nextSize)
		recommendedPosition := findRecommendedPosition(col.ColumnName, sortedColumnList)

		writer.Write([]string{
			strconv.Itoa(col.OrdinalPosition),
			col.ColumnName,
			col.DataType,
			col.IsNullable,
			strconv.Itoa(currentSize),
			strconv.Itoa(wastedPadding),
			strconv.Itoa(recommendedPosition),
			strconv.Itoa(col.EntryCount * wastedPadding),
		})
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Unable to get current directory: %v", err)
	}
	fmt.Printf("Report %s/%s generated successfully.\n", dir, reportName)

	return nil
}

func calculateWastedPadding(currentSize, nextSize int) int {
	if nextSize == 0 {
		return 0
	}
	return (nextSize - (currentSize % nextSize)) % nextSize
}

func findRecommendedPosition(columnName string, sortedColumnList []common.ColumnInfo) int {
	for i, col := range sortedColumnList {
		if col.ColumnName == columnName {
			return i + 1
		}
	}
	return -1
}
