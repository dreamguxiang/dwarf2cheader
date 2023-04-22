package dwarfhelper

import (
	"debug/dwarf"
	"fmt"
)

func (_this *DwarfInfo) getTypeName(entry *dwarf.Entry, isConst bool) string {
	var name string
	if entry.Val(dwarf.AttrName) != nil {
		name = entry.Val(dwarf.AttrName).(string)
	}
	switch entry.Tag {
	case dwarf.TagConstType:
		isConst = true
	case dwarf.TagEnumerationType, dwarf.TagTypedef:
		if name != "" {
			return name
		}
	}
	if entry.Val(dwarf.AttrType) != nil {
		offset, ok := entry.Val(dwarf.AttrType).(dwarf.Offset)
		if !ok {
			return "void**"
		}
		subEntry, err := _this.getEntryByOffset(offset)
		if err != nil {
			return "void**"
		}
		return _this.getTypeName(subEntry, isConst)
	}
	if name == "" {
		return ""
	}
	if isConst {
		return "const " + name
	}
	return name
}

func (_this *DwarfInfo) hasDataMemberLoc(entry *dwarf.Entry) bool {
	if entry.Val(dwarf.AttrDataMemberLoc) != nil {
		return true
	} else {
		return false
	}
}

func (_this *DwarfInfo) hasContainingType(entry *dwarf.Entry) bool {
	if entry.Val(dwarf.AttrContainingType) != nil {
		return true
	} else {
		return false
	}
}

func (_this *DwarfInfo) getEntryByOffset(offset dwarf.Offset) (*dwarf.Entry, error) {
	entry, ok := _this.Offset2entry[offset]
	if !ok {
		return nil, fmt.Errorf("offset %d not found", offset)
	}
	return entry, nil
}
