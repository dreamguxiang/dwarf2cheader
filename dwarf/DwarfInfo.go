package dwarfhelper

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"debug/dwarf"
	"debug/elf"
	"dwarf2cheader/utils"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

type DwarfInfo struct {
	elfFile *elf.File
	data    *dwarf.Data
	enumMap map[string]typeEnum
}

type typeEnum struct {
	Size     int64
	Base     string
	EnumType dwarf.EnumType
}

func DWARF(f *elf.File) (*dwarf.Data, error) {
	dwarfSuffix := func(s *elf.Section) string {
		switch {
		case strings.HasPrefix(s.Name, ".debug_"):
			return s.Name[7:]
		case strings.HasPrefix(s.Name, ".zdebug_"):
			return s.Name[8:]
		default:
			return ""
		}

	}
	// sectionData gets the data for s, checks its size, and
	// applies any applicable relations.
	sectionData := func(i int, s *elf.Section) ([]byte, error) {
		b, err := s.Data()
		if err != nil && uint64(len(b)) < s.Size {
			return nil, err
		}

		if len(b) >= 12 && string(b[:4]) == "ZLIB" {
			dlen := binary.BigEndian.Uint64(b[4:12])
			dbuf := make([]byte, dlen)
			r, err := zlib.NewReader(bytes.NewBuffer(b[12:]))
			if err != nil {
				return nil, err
			}
			if _, err := io.ReadFull(r, dbuf); err != nil {
				return nil, err
			}
			if err := r.Close(); err != nil {
				return nil, err
			}
			b = dbuf
		}

		// NOTE: removed relocations from original code here

		return b, nil
	}

	// There are many DWARf sections, but these are the ones
	// the debug/dwarf package started with.
	var dat = map[string][]byte{"abbrev": nil, "info": nil, "str": nil, "line": nil, "ranges": nil}
	for i, s := range f.Sections {
		suffix := dwarfSuffix(s)
		if suffix == "" {
			continue
		}
		if _, ok := dat[suffix]; !ok {
			continue
		}
		b, err := sectionData(i, s)
		if err != nil {
			return nil, err
		}
		dat[suffix] = b
	}

	d, err := dwarf.New(dat["abbrev"], nil, nil, dat["info"], dat["line"], nil, dat["ranges"], dat["str"])
	if err != nil {
		return nil, err
	}

	// Look for DWARF4 .debug_types sections and DWARF5 sections.
	for i, s := range f.Sections {
		suffix := dwarfSuffix(s)
		if suffix == "" {
			continue
		}
		if _, ok := dat[suffix]; ok {
			// Already handled.
			continue
		}

		b, err := sectionData(i, s)
		if err != nil {
			return nil, err
		}

		if suffix == "types" {
			if err := d.AddTypes(fmt.Sprintf("types-%d", i), b); err != nil {
				return nil, err
			}
		} else {
			if err := d.AddSection(".debug_"+suffix, b); err != nil {
				return nil, err
			}
		}
	}

	return d, nil
}

func NewDwarfInfo(input string) (*DwarfInfo, error) {
	elfFile, err := elf.Open(input)
	if err != nil {
		return nil, err
	}
	dwarfOut, err := DWARF(elfFile)
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

		//genericType, err := _this.data.Type(entry.Offset)
		//if err != nil {
		//	break
		//}

		//enumType, ok := genericType.(*dwarf.EnumType)
		//if ok != true {
		//	return fmt.Errorf("%s is not an EnumType?", genericType.String())
		//}
		//if !utils.FilterEnumName(enumType.EnumName) {
		tempEnum := typeEnum{
			Size:     enumType.ByteSize,
			Base:     "void",
			EnumType: *enumType,
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
		if _, ok := _this.enumMap[GetEnumName(*enumType)]; ok {
			_this.enumMap[GetEnumName(*enumType)+"_"+utils.GetRandomString(4)] = tempEnum
		} else {
			_this.enumMap[GetEnumName(*enumType)] = tempEnum
		}
	}
	//}
	return nil
}
