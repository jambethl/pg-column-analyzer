package report

import (
	"encoding/csv"
	"fmt"
	"log"
	"main/pkg/types"
	"os"
	"strconv"
)

var dataTypeMap = map[string]string{
	"smallint":                    "2 bytes",
	"integer":                     "4 bytes",
	"bigint":                      "8 bytes",
	"boolean":                     "1 byte",
	"real":                        "4 bytes",
	"double precision":            "8 bytes",
	"date":                        "4 bytes",
	"timestamp without time zone": "8 bytes",
	"timestamp with time zone":    "8 bytes",
	"uuid":                        "16 bytes",
	"text":                        "10 bytes", // Variable size, often larger
	"varchar":                     "10 bytes", // Variable size, often larger
	"bytea":                       "10 bytes", // Variable size, often larger
}

func GenerateReport(columnList []types.ColumnInfo, tableName string) error {
	reportName := fmt.Sprintf("reports/%s_report.csv", tableName)
	file, err := os.Create(reportName)
	if err != nil {
		log.Fatalf("Unable to create report: %s", reportName)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Ordinal Position", "Column Name", "Data Type", "Nullable", "Data Type Size (B)", "Wasted Padding", "Recommended Position"})

	for _, col := range columnList {
		writer.Write([]string{
			strconv.Itoa(col.OrdinalPosition),
			col.ColumnName,
			col.DataType,
			col.IsNullable,
			dataTypeMap[col.DataType],
			"",
			"",
		})
	}

	dir, err := os.Getwd()
	fmt.Printf("Report %s/%s generated successfully.\n", dir, reportName)

	return nil
}
