package tools

import (
    "crypto/md5"
    "io/ioutil"
    "fmt"
)

func B2i(b bool) uint8 {
    if b {
        return 1
    }
    return 0
}

func Flipbyte(b uint8) uint8 {
    b = (b & 0xF0) >> 4 | (b & 0x0F) << 4
    b = (b & 0xCC) >> 2 | (b & 0x33) << 2
    b = (b & 0xAA) >> 1 | (b & 0x55) << 1
    return b
}


func HashFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(data)), nil
}

