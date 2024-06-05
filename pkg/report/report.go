package report

import (
	"encoding/csv"
	"fmt"
	"log"
	"main/pkg/types"
	"os"
	"strconv"
)

func GenerateReport(columnList []types.ColumnInfo, tableName string) error {
	reportName := fmt.Sprintf("reports/%s_report.csv", tableName)
	file, err := os.Create(reportName)
	if err != nil {
		log.Fatalf("Unable to create report: %s", reportName)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Ordinal Position", "Column Name", "Data Type", "Nullable"})

	for _, col := range columnList {
		writer.Write([]string{
			strconv.Itoa(col.OrdinalPosition),
			col.ColumnName,
			col.DataType,
			col.IsNullable,
		})
	}

	dir, err := os.Getwd()
	fmt.Printf("Report %s/%s generated successfully.\n", dir, reportName)

	return nil
}
