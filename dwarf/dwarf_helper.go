package dwarfhelper

import (
	"crypto/sha1"
	"debug/dwarf"
	"debug/elf"
	"fmt"
)

type DwarfInfo struct {
	elfFile      *elf.File
	data         *dwarf.Data
	enumMap      map[string]typeEnum
	udtMap       map[string]typeUDT
	Offset2entry map[dwarf.Offset]*dwarf.Entry
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
		elfFile:      elfFile,
		data:         dwarfOut,
		enumMap:      make(map[string]typeEnum),
		udtMap:       make(map[string]typeUDT),
		Offset2entry: make(map[dwarf.Offset]*dwarf.Entry),
	}, nil
}

func (_this *DwarfInfo) GetData() *dwarf.Data {
	return _this.data
}

func (_this *DwarfInfo) GetEnumMap() map[string]typeEnum {
	return _this.enumMap
}

func (_this *DwarfInfo) GetUDTMap() map[string]typeUDT {
	return _this.udtMap
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
		{
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
			tempEnum.Base = _this.getEnumTypeName(entry)
			_this.enumMap[GetEnumName(*enumType)] = tempEnum
		}
	case dwarf.TagClassType, dwarf.TagStructType:
		{
			t := new(dwarf.StructType)
			switch entry.Tag {
			case dwarf.TagClassType:
				t.Kind = "class"
			case dwarf.TagStructType:
				t.Kind = "struct"
			}
			t.StructName, _ = entry.Val(dwarf.AttrName).(string)
			t.Incomplete = entry.Val(dwarf.AttrDeclaration) != nil
			UDT := typeUDT{
				Size: t.ByteSize,
			}
			//继承关系
			if _this.hasContainingType(entry) {
				contaoff := entry.Val(dwarf.AttrContainingType)
				if contaoff != nil {
					if offset, ok := contaoff.(dwarf.Offset); ok {
						offsetout, _ := _this.getEntryByOffset(offset)
						UDT.Containing = offsetout.Val(dwarf.AttrName).(string)
					}
				}
			}

			for kid := next(); kid != nil; kid = next() {
				if kid.Tag == dwarf.TagMember {
					if !_this.hasDataMemberLoc(kid) {
						continue
					}
					f := new(vStructField)
					off := kid.Val(dwarf.AttrType)
					if off != nil {
						if offset, ok := off.(dwarf.Offset); ok {
							offsetout, err := _this.getEntryByOffset(offset)
							var temp vEntry
							if err != nil {
								temp = vEntry{
									TypeName: "NoSupport",
								}
							} else {
								temp = vEntry{
									TypeName: _this.getTypeName(offsetout, false),
								}
							}
							f.Entry = temp
						}
					}
					f.Name, _ = kid.Val(dwarf.AttrName).(string)
					f.ByteOffset, _ = kid.Val(dwarf.AttrDataMemberLoc).(int64)
					f.BitOffset, _ = kid.Val(dwarf.AttrBitOffset).(int64)
					f.BitSize, _ = kid.Val(dwarf.AttrBitSize).(int64)

					n := len(UDT.ExStructField)
					if n >= cap(UDT.ExStructField) {
						val := make([]*vStructField, n, n*2)
						copy(val, UDT.ExStructField)
						UDT.ExStructField = val
					}
					if len(UDT.ExStructField) == 0 {
						UDT.ExStructField = make([]*vStructField, 0, 8)
					}
					UDT.ExStructField = UDT.ExStructField[0 : n+1]
					UDT.ExStructField[n] = f

				}
			}
			UDT.StructType = *t
			if len(UDT.ExStructField) == 0 && UDT.Containing == "" {
				return nil
			}
			if _, ok := _this.udtMap[t.StructName]; !ok {
				_this.udtMap[t.StructName] = UDT
			} else {
				if len(_this.udtMap[t.StructName].ExStructField) == 0 {
					_this.udtMap[t.StructName] = UDT
				}
			}

		}
	}
	return nil
}
