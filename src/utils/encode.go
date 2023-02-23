package utils

import(
	"bytes"
	"io/ioutil"
	"fmt"
	"strings"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"regexp"
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
