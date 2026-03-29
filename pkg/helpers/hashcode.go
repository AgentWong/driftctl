package helpers

import (
	"bytes"
	"fmt"
	"hash/crc32"
)

// HashcodeString hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// a non negative integer. Here we cast to an integer
// and invert it if the result is negative.
func HashcodeString(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	// v == MinInt
	return 0
}

// HashcodeStrings hashes a list of strings to a unique hashcode.
func HashcodeStrings(strings []string) string {
	var buf bytes.Buffer

	for _, s := range strings {
		buf.WriteString(fmt.Sprintf("%s-", s))
	}

	return fmt.Sprintf("%d", HashcodeString(buf.String()))
}
