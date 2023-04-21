package utils

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
