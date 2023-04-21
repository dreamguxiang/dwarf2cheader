package dwarfhelper

import (
	"crypto/sha1"
	"debug/dwarf"
	"debug/elf"
	"dwarf2cheader/utils"
	"fmt"
)

type DwarfInfo struct {
	elfFile      *elf.File
	data         *dwarf.Data
	enumMap      map[string]typeEnum
	udtMap       map[string]typeUDT
	Offset2entry map[dwarf.Offset]*dwarf.Entry
}

type typeEnum struct {
	Size      int64
	Base      string
	EnumClass bool
	EnumType  dwarf.EnumType
}

type typeUDT struct {
	Size          int64
	StructType    dwarf.StructType
	ExStructField []*vStructField
}

type vStructField struct {
	Name          string
	Entry         dwarf.Entry
	ByteOffset    int64
	ByteSize      int64 // usually zero; use Type.Size() for normal fields
	BitOffset     int64
	DataBitOffset int64
	BitSize       int64 // zero if not a bit field
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

func (_this *DwarfInfo) getEntryByOffset(offset dwarf.Offset) (*dwarf.Entry, error) {
	entry, ok := _this.Offset2entry[offset]
	if !ok {
		return nil, fmt.Errorf("offset %d not found", offset)
	}
	return entry, nil
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

			isSigned := false
			for _, v := range enumType.Val {
				if v.Val < 0 {
					isSigned = true
				}
			}
			tempEnum.Base = utils.GetEnumType(tempEnum.Size, isSigned)
			_this.enumMap[GetEnumName(*enumType)] = tempEnum
		}
	case dwarf.TagClassType, dwarf.TagStructType:
		{
			//genericType, err := _this.data.Type(entry.Offset)
			//if err != nil {
			//	break
			//}
			//structType, ok := genericType.(*dwarf.StructType)

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

			for kid := next(); kid != nil; kid = next() {
				if kid.Tag == dwarf.TagMember {
					f := new(vStructField)
					off := kid.Val(dwarf.AttrType)
					if off != nil {
						if offset, ok := off.(dwarf.Offset); ok {
							offsetout, err := _this.getEntryByOffset(offset)
							if err != nil {
								continue
							}
							f.Entry = *offsetout
						}
					}
					f.Name, _ = kid.Val(dwarf.AttrName).(string)
					f.ByteOffset, _ = kid.Val(dwarf.AttrDataMemberLoc).(int64)
					f.BitOffset, _ = kid.Val(dwarf.AttrBitOffset).(int64)
					f.BitSize, _ = kid.Val(dwarf.AttrBitSize).(int64)

					//fmt.Println("member name:", f.Name, "offset:", f.ByteOffset, "bit offset:", f.BitOffset, "bit size:", f.BitSize)
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
			// 如果不存在，就添加，如果存在，判断ExStructField是否为空，如果为空，就添加
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
