package dwarfhelper

import (
	"crypto/sha1"
	"debug/dwarf"
	"debug/elf"
	"dwarf2cheader/utils"
	"fmt"
)

type DwarfInfo struct {
	elfFile *elf.File
	data    *dwarf.Data
	enumMap map[string]typeEnum
}

type typeEnum struct {
	Size      int64
	Base      string
	EnumClass bool
	EnumType  dwarf.EnumType
}

func NewDwarfInfo(input string) (*DwarfInfo, error) {
	elfFile, err := elf.Open(input)
	if err != nil {
		return nil, err
	}
	dwarfOut, err := elfFile.DWARF()
	if err != nil {
		return nil, err
	}
	return &DwarfInfo{
		elfFile: elfFile,
		data:    dwarfOut,
		enumMap: make(map[string]typeEnum),
	}, nil
}

func (_this *DwarfInfo) GetData() *dwarf.Data {
	return _this.data
}

func (_this *DwarfInfo) GetEnumMap() map[string]typeEnum {
	return _this.enumMap
}

func GetEnumName(dwarfEnum dwarf.EnumType) string {
	if dwarfEnum.EnumName != "" {
		return dwarfEnum.EnumName
	}
	data := sha1.Sum([]byte(dwarfEnum.String()))
	return fmt.Sprintf("$%8x", data[0:8])
}

func (_this *DwarfInfo) GetType(entry *dwarf.Entry, reader *dwarf.Reader) error {
	if entry == nil {
		return fmt.Errorf("entry is nil")
	}
	nextDepth := 0

	next := func() *dwarf.Entry {
		if !entry.Children {
			return nil
		}
		// Only return direct children.
		// Skip over composite entries that happen to be nested
		// inside this one. Most DWARF generators wouldn't generate
		// such a thing, but clang does.
		// See golang.org/issue/6472.
		for {
			kid, err1 := reader.Next()
			if err1 != nil {
				return nil
			}
			if kid == nil {
				return nil
			}
			if kid.Tag == 0 {
				if nextDepth > 0 {
					nextDepth--
					continue
				}
				return nil
			}
			if kid.Children {
				nextDepth++
			}
			if nextDepth > 0 {
				continue
			}
			return kid
		}
	}

	switch entry.Tag {
	case dwarf.TagEnumerationType:
		enumType := new(dwarf.EnumType)
		enumType.EnumName, _ = entry.Val(dwarf.AttrName).(string)
		enumClass, _ := entry.Val(dwarf.AttrEnumClass).(bool)
		enumType.Val = make([]*dwarf.EnumValue, 0, 8)
		for kid := next(); kid != nil; kid = next() {
			if kid.Tag == dwarf.TagEnumerator {
				f := new(dwarf.EnumValue)
				f.Name, _ = kid.Val(dwarf.AttrName).(string)
				f.Val, _ = kid.Val(dwarf.AttrConstValue).(int64)
				n := len(enumType.Val)
				if n >= cap(enumType.Val) {
					val := make([]*dwarf.EnumValue, n, n*2)
					copy(val, enumType.Val)
					enumType.Val = val
				}
				enumType.Val = enumType.Val[0 : n+1]
				enumType.Val[n] = f
			}
		}

		tempEnum := typeEnum{
			Size:      enumType.ByteSize,
			Base:      "void",
			EnumType:  *enumType,
			EnumClass: enumClass,
		}

		if tempEnum.Size < 0 {
			tempEnum.Size = 0
		}

		isSigned := false
		for _, v := range enumType.Val {
			if v.Val < 0 {
				isSigned = true
			}
		}
		tempEnum.Base = utils.GetEnumType(tempEnum.Size, isSigned)
		_this.enumMap[GetEnumName(*enumType)] = tempEnum
	}
	return nil
}
