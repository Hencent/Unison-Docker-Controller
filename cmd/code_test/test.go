package main

import (
	"fmt"
	"strconv"
)

func main() {
	f := 101313130.12345678901234567890123456789
	fmt.Println(strconv.FormatFloat(f, 'f', 5, 32))

	var ii int64 = 1231414131241431414
	fmt.Println(strconv.FormatInt(ii, 10))

}
