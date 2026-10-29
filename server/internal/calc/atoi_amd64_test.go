package calc

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestAtoi(t *testing.T) {
	t.Run("atoi", func(t *testing.T) {
		var data [64]byte
		copy(data[:], "1234")

		// fmt.Println(hex.EncodeToString(data[:]))
		fmt.Println("Before")
		for _, n := range data[:] {
			fmt.Printf("%08b ", n) // prints 00000000 11111101
		}
		result := __atoi_AVX512(&data)
		fmt.Println("\nAfter")
		fmt.Println(hex.EncodeToString(data[:]))
		for _, n := range data[:] {
			fmt.Printf("%08b ", n) // prints 00000000 11111101
		}
		fmt.Println("\nNumeric")
		fmt.Println(result)
		// fmt.Println(binary.LittleEndian.Uint64(data[:8]))
		// Do it MAGICALLY!
		// fmt.Println(*(*int64)(unsafe.Pointer(&data[0])))
		// fmt.Println(*(*int64)(unsafe.Pointer(&data[8])))
		// fmt.Println(*(*int64)(unsafe.Pointer(&data[16])))
		// fmt.Println(*(*int64)(unsafe.Pointer(&data[24])))
		// fmt.Println(*(*int64)(unsafe.Pointer(&data[32])))

	})
}
