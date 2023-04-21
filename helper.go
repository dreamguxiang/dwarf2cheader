package main

import (
	dwarfhelper "dwarf2cheader/dwarf"
	"fmt"
	"os"
)

func DwarfHelper(ipath string) error {
	info, err := dwarfhelper.NewDwarfInfo(ipath)
	if err != nil {
		return err
	}
	reader := info.GetData().Reader()

	//log, _ := os.Create("log.log")
	//defer log.Close()
	for {
		entry, err := reader.Next()
		if err != nil {
			fmt.Println("finish")
			break
		}
		//log.WriteString(fmt.Sprintf("%v\n", entry))
		err = info.GetType(entry, reader)
		if err != nil {
			break
		}
	}
	err = GenerateEnumCHeaderFile(info)
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
		_, err = create.WriteString(fmt.Sprintf("enum %s : %s {\n", dwarfhelper.GetEnumName(v.EnumType), v.Base))
		if err != nil {
			return err
		}
		for _, v1 := range v.EnumType.Val {
			_, err = create.WriteString(fmt.Sprintf("	%s = 0x%d,\n", v1.Name, v1.Val))
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
