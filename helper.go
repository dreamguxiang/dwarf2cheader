package main

import (
	"debug/dwarf"
	dwarfhelper "dwarf2cheader/dwarf"
	"fmt"
	"os"
	"strings"
)

func DwarfHelper(ipath string) error {
	info, err := dwarfhelper.NewDwarfInfo(ipath)
	if err != nil {
		return err
	}
	reader := info.GetData().Reader()

	//log, _ := os.Create("log.log")
	//defer log.Close()
	//存入临时偏移变量
	for {
		entry, err := reader.Next()
		if err != nil {
			break
		}
		if entry == nil {
			break
		}
		if entry.Tag == dwarf.TagNamespace ||
			entry.Tag == dwarf.TagClassType ||
			entry.Tag == dwarf.TagStructType ||
			entry.Tag == dwarf.TagEnumerationType ||
			entry.Tag == dwarf.TagBaseType {
			info.Offset2entry[entry.Offset] = entry
		}
	}
	reader.Seek(0)
	for {
		entry, err := reader.Next()
		if err != nil {
			fmt.Println("finish")
			break
		}
		err = info.GetType(entry, reader)
		if err != nil {
			break
		}
	}
	err = GenerateEnumCHeaderFile(info)
	if err != nil {
		return err
	}
	err = GenerateUdtCHeaderFile(info)
	if err != nil {
		return err
	}
	return nil
}

func GenerateEnumCHeaderFile(info *dwarfhelper.DwarfInfo) error {
	create, err := os.Create("enums.h")
	if err != nil {
		return err
	}
	defer create.Close()
	_, err = create.WriteString("#pragma once\n")
	if err != nil {
		return err
	}
	for _, v := range info.GetEnumMap() {
		if v.EnumClass {
			_, err = create.WriteString(fmt.Sprintf("enum class %s : %s {\n", dwarfhelper.GetEnumName(v.EnumType), v.Base))
		} else {
			_, err = create.WriteString(fmt.Sprintf("enum %s : %s {\n", dwarfhelper.GetEnumName(v.EnumType), v.Base))
		}
		if err != nil {
			return err
		}
		for _, v1 := range v.EnumType.Val {
			switch v.Base {
			case "__uint8":
				fallthrough
			case "__int8":
				_, err = create.WriteString(fmt.Sprintf("\t%s = 0x%X\n", v1.Name, uint8(v1.Val)))
			case "__uint16":
				fallthrough
			case "__int16":
				_, err = create.WriteString(fmt.Sprintf("\t%s = 0x%X,\n", v1.Name, uint16(v1.Val)))
			case "__uint32":
				fallthrough
			case "__int32":
				_, err = create.WriteString(fmt.Sprintf("\t%s = 0x%X,\n", v1.Name, uint32(v1.Val)))
			case "__uint64":
				fallthrough
			case "__int64":
				_, err = create.WriteString(fmt.Sprintf("\t%s = 0x%X,\n", v1.Name, uint64(v1.Val)))
			}
			if err != nil {
				return err
			}
		}
		_, err = create.WriteString(fmt.Sprintf("};\n\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func GenerateUdtCHeaderFile(info *dwarfhelper.DwarfInfo) error {
	create, err := os.Create("udts.h")
	if err != nil {
		return err
	}
	defer create.Close()
	_, err = create.WriteString("#pragma once\n")
	if err != nil {
		return err
	}
	//过滤 std:: 标准库
	//新建一个过滤切片
	filter := make([]string, 0)
	filter = append(filter, "std::", "__gnu_cxx")
	for k, v := range info.GetUDTMap() {
		for _, v1 := range filter {
			if strings.Contains(k, v1) {
				continue
			}
		}
		_, err = create.WriteString(fmt.Sprintf("//size: \n"))
		if v.StructType.Kind == "struct" {
			_, err = create.WriteString(fmt.Sprintf("struct %s {\n", k))
		} else {
			_, err = create.WriteString(fmt.Sprintf("class %s {\n", k))
		}
		if err != nil {
			return err
		}
		for _, v1 := range v.ExStructField {
			_, err = create.WriteString(fmt.Sprintf("\t%s %s; // %d\n", v1.Entry.Val(dwarf.AttrName), v1.Name, v1.ByteOffset))
			if err != nil {
				return err
			}
		}

		_, err = create.WriteString(fmt.Sprintf("};\n\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
