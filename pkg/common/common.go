package common

type ColumnInfo struct {
	OrdinalPosition int
	ColumnName      string
	DataType        string
	IsNullable      string
	EntryCount      int
	TypAlign        int
}
