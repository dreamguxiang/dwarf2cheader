package dwarfhelper

import (
	"debug/dwarf"
)

type typeEnum struct {
	Size      int64
	Base      string
	EnumClass bool
	EnumType  dwarf.EnumType
}
