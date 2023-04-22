package dwarfhelper

import (
	"debug/dwarf"
	"fmt"
)

func (_this *DwarfInfo) getEnumTypeName(entry *dwarf.Entry) string {
	if entry.Val(dwarf.AttrType) != nil {
		offset, _ := entry.Val(dwarf.AttrType).(dwarf.Offset)
		entryoff, _ := _this.getEntryByOffset(offset)
		if entryoff == nil {
			return "unknown"
		}
		return _this.getEnumTypeName(entryoff)
	} else {
		if entry.Val(dwarf.AttrName) != nil {
			return entry.Val(dwarf.AttrName).(string)
		} else {
			return "unknown"
		}
	}
}

func (_this *DwarfInfo) getTypeName(entry *dwarf.Entry, isConst bool) string {
	var result string
	if isConst {
		result += "const "
	}
	switch entry.Tag {
	case dwarf.TagConstType:
		isConst = true
	case dwarf.TagEnumerationType:
		if entry.Val(dwarf.AttrName) != nil {
			return result + entry.Val(dwarf.AttrName).(string)
		}
	case dwarf.TagTypedef:
		if entry.Val(dwarf.AttrName) != nil {
			return result + entry.Val(dwarf.AttrName).(string)
		}
	}
	if entry.Val(dwarf.AttrType) != nil {
		offset, _ := entry.Val(dwarf.AttrType).(dwarf.Offset)
		entryoff, _ := _this.getEntryByOffset(offset)
		if entryoff == nil {
			return "void**"
		}
		return _this.getTypeName(entryoff, isConst)
	} else {
		if entry.Val(dwarf.AttrName) != nil {
			return result + entry.Val(dwarf.AttrName).(string)
		} else {
			return ""
		}
	}
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
