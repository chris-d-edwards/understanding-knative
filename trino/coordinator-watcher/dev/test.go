package main

import (
	"fmt"
	"strconv"
	"time"
)

func main() {

	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)

	fmt.Println(timestamp) // prints: 1436773875771421417
}
