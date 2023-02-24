package utils

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
)

// UTF-8 转 UCS-2
func EncodeUcs2(s string) string {
	result := fmt.Sprintf("%U", []rune(s))
	result = strings.ReplaceAll(result, "[", "")
	result = strings.ReplaceAll(result, "]", "")
	result = strings.ReplaceAll(result, "U+", "")
	result = strings.ReplaceAll(result, " ", "")
	return result
}

// UCS-2 转 UTF-8
func DecodeUcs2(in string) (str string) {
	i := 0
	j := 1
	k := 0
	octets := make([]byte, len(in)/2)
	for j < len(in) {
		n, err := strconv.ParseUint(string(in[i])+string(in[j]), 16, 64)
		if err != nil {
			fmt.Println(err)
		}
		octets[k] = byte(n)
		i += 2
		j += 2
		k++
	}
	if len(octets)%2 != 0 {
		return
	}
	buf := make([]uint16, 0, len(octets)/2)
	for i := 0; i < len(octets); i += 2 {
		buf = append(buf, uint16(octets[i])<<8|uint16(octets[i+1]))
	}
	runes := utf16.Decode(buf)
	return string(runes)
}
