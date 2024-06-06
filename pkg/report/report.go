package report

import (
	"encoding/csv"
	"fmt"
	"log"
	"main/pkg/types"
	"os"
	"strconv"
)

var dataTypeMap = map[string]int{
	"smallint":                    2,
	"integer":                     4,
	"bigint":                      8,
	"boolean":                     1,
	"real":                        4,
	"double precision":            8,
	"date":                        4,
	"timestamp without time zone": 8,
	"timestamp with time zone":    8,
	"uuid":                        16,
	"text":                        10, // Variable size, often larger
	"varchar":                     10, // Variable size, often larger
	"bytea":                       10, // Variable size, often larger
}

func GenerateReport(columnList []types.ColumnInfo, tableName string) error {
	reportName := fmt.Sprintf("reports/%s_report.csv", tableName)
	file, err := os.Create(reportName)
	if err != nil {
		log.Fatalf("Unable to create report: %s", reportName)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{"Ordinal Position", "Column Name", "Data Type", "Nullable", "Data Type Size (B)", "Wasted Padding", "Recommended Position"})
	if err != nil {
		return fmt.Errorf("failed to write header row: %v", err)
	}

	currentOffset := 0
	for i, col := range columnList {
		currentSize := dataTypeMap[col.DataType]
		var wastedPadding int

		if i < len(columnList)-1 {
			nextSize := dataTypeMap[columnList[i+1].DataType]
			if nextSize != 0 {
				wastedPadding = (currentSize - (currentOffset % nextSize)) % nextSize
			} else {
				wastedPadding = 0
			}
		} else {
			wastedPadding = 0
		}

		currentOffset += currentSize + wastedPadding

		err := writer.Write([]string{
			strconv.Itoa(col.OrdinalPosition),
			col.ColumnName,
			col.DataType,
			col.IsNullable,
			strconv.Itoa(dataTypeMap[col.DataType]),
			strconv.Itoa(wastedPadding),
			strconv.Itoa(i + 1),
		})
		if err != nil {
			return fmt.Errorf("failed to write data row: %v", err)
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Printf("Report %s/%s generated successfully.\n", dir, reportName)

	return nil
}
