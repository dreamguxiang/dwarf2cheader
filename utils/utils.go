package utils

import (
	mapset "github.com/deckarep/golang-set"
	"math/rand"
	"time"
)

// GetEnumType 根据enum的大小判断是c类型
func GetEnumType(size int64, isSigned bool) string {
	if isSigned {
		switch size {
		case 1:
			return "__int8"
		case 2:
			return "__int16"
		case 4:
			return "__int32"
		case 8:
			return "__int64"
		}
	} else {
		switch size {
		case 1:
			return "__uint8"
		case 2:
			return "__uint16"
		case 4:
			return "__uint32"
		case 8:
			return "__uint64"
		}
	}
	return "__int32"
}

var filterList = []interface{}{"__cxx", "_ZN", "_ZT", "std::"}

// FilterEnumName 过滤掉不需要的EnumName
func FilterEnumName(name string) bool {
	s := mapset.NewSetFromSlice(filterList)
	if s.Contains(name) || len(name) == 0 {
		return true
	}
	return false
}

func NewRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GetRandomString 生成随机字符串
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := NewRand()
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)

}