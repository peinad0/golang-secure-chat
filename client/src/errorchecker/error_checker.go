package errorchecker

import "fmt"

// Check checks error and logs if is not nil
func Check(message string, e error) bool {
	isError := false
	if e != nil {
		fmt.Println(message)
		isError = true
	}
	return isError
}
