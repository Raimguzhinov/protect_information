package common

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Cipher interface {
	Encrypt() error
	Decrypt() error
	Do() error
}

// WriteNumbers Запись чисел в io.Writer
func WriteNumbers(w io.Writer, data []int64) error {
	for _, num := range data {
		if err := binary.Write(w, binary.LittleEndian, uint8(num)); err != nil {
			return fmt.Errorf("error writing number to output: %v", err)
		}
	}
	return nil
}

// WriteData - запись данных в io.Writer
func WriteData(w io.Writer, data []byte) error {
	for _, b := range data {
		if err := binary.Write(w, binary.LittleEndian, b); err != nil {
			return fmt.Errorf("error writing data: %v", err)
		}
	}
	return nil
}

// WritePair - запись зашифрованных данных в io.Writer
func WritePair(w io.Writer, encryptedMessage [][2]int64) error {
	for _, pair := range encryptedMessage {
		if err := binary.Write(w, binary.LittleEndian, pair[0]); err != nil {
			return fmt.Errorf("error writing encrypted data: %v", err)
		}
		if err := binary.Write(w, binary.LittleEndian, pair[1]); err != nil {
			return fmt.Errorf("error writing encrypted data: %v", err)
		}
	}
	return nil
}
