package utils

import(
	"bytes"
	"io/ioutil"
	"fmt"
	"strings"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"regexp"
	"unicode/utf16"
	"strconv"
)

func Utf8ToUcs2(in string) (string, error) {
	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()) //UTF-16 bigendian, no-bom
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	hexStr := fmt.Sprintf("%X", out)
	return hexStr, nil
}

func Ucs2ToUtf8(in string) (string, error) {
	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()) //UTF-16 bigendian, no-bom
	out, err := ioutil.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func IsUcs(body string) bool {
	reg := regexp.MustCompile(`[^[:xdigit:]]`)
	idx := reg.FindStringIndex(body)
	if len(idx) == 0 {
		return true
	} else {
		return false
	}
}

func GetStrUnicode(s string) string {
	result := fmt.Sprintf("%U", []rune(s))
	result = strings.ReplaceAll(result, "[", "")
	result = strings.ReplaceAll(result, "]", "")
	result = strings.ReplaceAll(result, "U+", "")
	result = strings.ReplaceAll(result, " ", "")
	return result
}

// DecodeUcs2 decodes the given UCS2 (UTF-16) octet data into a UTF-8 encoded string.
func DecodeUcs2(a string) (str string) {
	var i int = 0
	var j int = 1
	k := 0
	octets := make([]byte, len(a)/2)
	for ; j < len(a); {
		n, err := strconv.ParseUint(string(a[i])+string(a[j]), 16, 64)
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
