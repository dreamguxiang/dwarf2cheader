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

	for {
		entry, err := reader.Next()
		if err != nil {
			break
		}
		if entry == nil {
			break
		}
		if entry.Tag == dwarf.TagClassType ||
			entry.Tag == dwarf.TagStructType ||
			entry.Tag == dwarf.TagEnumerationType ||
			entry.Tag == dwarf.TagBaseType ||
			entry.Tag == dwarf.TagPointerType ||
			entry.Tag == dwarf.TagTypedef ||
			entry.Tag == dwarf.TagConstType ||
			entry.Tag == dwarf.TagNamespace ||
			entry.Tag == dwarf.TagReferenceType ||
			entry.Tag == dwarf.TagArrayType {
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
	for k, v := range info.GetEnumMap() {
		if v.EnumClass {
			_, err = create.WriteString(fmt.Sprintf("enum class %s : %s {\n", k, v.Base))
		} else {
			_, err = create.WriteString(fmt.Sprintf("enum %s : %s {\n", k, v.Base))
		}
		if err != nil {
			return err
		}
		for _, v1 := range v.EnumType.Val {
			switch v.Base {
			case "unsigned char":
				fallthrough
			case "signed char":
				fallthrough
			case "char":
				_, err = create.WriteString(fmt.Sprintf("\t%s = 0x%X\n", v1.Name, uint8(v1.Val)))
			case "unsigned short":
				fallthrough
			case "short":
				_, err = create.WriteString(fmt.Sprintf("\t%s = 0x%X,\n", v1.Name, uint16(v1.Val)))
			case "unsigned int":
				fallthrough
			case "int":
				_, err = create.WriteString(fmt.Sprintf("\t%s = 0x%X,\n", v1.Name, uint32(v1.Val)))
			case "long":
				fallthrough
			case "long unsigned int":
				fallthrough
			case "unsigned long":
				_, err = create.WriteString(fmt.Sprintf("\t%s = 0x%X,\n", v1.Name, uint64(v1.Val)))
			default:
				_, err = create.WriteString(fmt.Sprintf("\t%s = 0x%X,\n", v1.Name, v1.Val))
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

	filter := make([]string, 0)
	filter = append(filter, "std::", "__gnu_cxx", "__aligned_buffer", "_Multi_array", "__shared_ptr_access", "__result_of_success", "_Tuple_impl")
	for k, v := range info.GetUDTMap() {
		for _, v1 := range filter {
			if strings.Contains(k, v1) {
				continue
			}
		}
		_, err = create.WriteString(fmt.Sprintf("//size: \n"))
		if v.StructType.Kind == "struct" {
			if len(v.Base) == 0 {
				_, err = create.WriteString(fmt.Sprintf("struct %s {\n", k))
			} else {
				var t string
				for k2, v2 := range v.Base {
					if k2 == len(v.Base)-1 {
						t += v2.TypeName
					} else {
						t += v2.TypeName + ","
					}
				}
				_, err = create.WriteString(fmt.Sprintf("struct %s : %s {\n", k, t))
			}
		} else {
			if len(v.Base) == 0 {
				_, err = create.WriteString(fmt.Sprintf("class %s {\n", k))
			} else {
				var t string
				for k2, v2 := range v.Base {
					if k2 == len(v.Base)-1 {
						t += v2.TypeName
					} else {
						t += v2.TypeName + ","
					}
				}
				_, err = create.WriteString(fmt.Sprintf("class %s : %s {\n", k, t))
			}

		}
		if err != nil {
			return err
		}
		for _, v1 := range v.ExStructField {
			_, err = create.WriteString(fmt.Sprintf("\t%s %s; // %d\n", v1.Entry.TypeName, v1.Name, v1.ByteOffset))
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
