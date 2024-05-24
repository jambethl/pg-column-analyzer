package report

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func GenerateReport(columnList []ColumnInfo, tableName string) {
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
			col.OrdinalPosition,
			col.ColumnName,
			col.DataType,
			col.IsNullable,
		})
	}

	fmt.Printf("Report %s generated successfully.\n", reportName)
}
