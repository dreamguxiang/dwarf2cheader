package dwarfhelper

import "debug/dwarf"

type typeUDT struct {
	Size          int64
	StructType    dwarf.StructType
	ExStructField []*vStructField
	Base          []vEntry
}

type vEntry struct {
	TypeName string
}

type vStructField struct {
	Name          string
	Entry         vEntry
	ByteOffset    int64
	ByteSize      int64
	BitOffset     int64
	DataBitOffset int64
	BitSize       int64
}
