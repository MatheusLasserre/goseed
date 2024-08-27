package size

import (
	"fmt"
	"unsafe"
)

type MyStruct struct {
	Field1 string
	Field2 int
	Field3 float64
}

func StructSize() int {
	data := []MyStruct{
		{"example1", 42, 3.14},
		{"example2", 100, 1.23},
	}

	var totalSize uintptr
	for _, item := range data {
		// Add the size of the struct itself
		totalSize += unsafe.Sizeof(item)

		// Add the size of dynamically allocated fields
		totalSize += uintptr(len(item.Field1)) // Size of the string content
	}

	fmt.Printf("Total size of struct array: %d bytes\n", totalSize)
	return int(totalSize)
}
