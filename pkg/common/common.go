package common

type ColumnInfo struct {
	OrdinalPosition int
	ColumnName      string
	DataType        string
	IsNullable      string
	EntryCount      int
	TypLen          int
	TypAlign        int
}
