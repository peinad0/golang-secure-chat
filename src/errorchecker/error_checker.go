package errorchecker

import (
	"fmt"
	"log"
)

// Check checks error and logs if is not nil
func Check(message string, e error) bool {
	isError := false
	if e != nil {
		fmt.Println(message)
		log.Fatal(e)
		isError = true
	}
	return isError
}
