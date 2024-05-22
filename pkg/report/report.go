package report

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func GenerateReport(column_list []ColumnInfo, table_name string) {
	report_name := fmt.Sprintf("reports/%s_report.csv", table_name)
	file, err := os.Create(report_name)
	if err != nil {
		log.Fatalf("Unable to create report: %s", report_name)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Ordinal Position", "Column Name", "Data Type", "Nullable"})

	for _, col := range column_list {
		writer.Write([]string{
			col.OrdinalPosition,
			col.ColumnName,
			col.DataType,
			col.IsNullable,
		})
	}

	fmt.Printf("Report %s generated successfully.\n", report_name)
}
