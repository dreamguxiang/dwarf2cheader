package rtti

import (
	"debug/elf"
	"fmt"
	"strings"
)

// 读取所有rtti符号，并生成继承结构树
func ReadRttiSymbols(file elf.File) error {
	symbols, err := file.Symbols()
	if err != nil {
		return err
	}
	for _, symbol := range symbols {
		if strings.Contains(symbol.Name, "RTTI") {
			fmt.Println(symbol.Name)

		}
	}
	fmt.Println("finish")

	return nil
}
