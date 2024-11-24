package example

import "fmt"

func main() {
	// @check:NotNullable
	var a = "Hello"

	// @check:Range:10,100
	var b = 50

	// @check:Range:10,100
	var c = 200

	// @check:MatchPattern:^[a-zA-Z0-9]+$
	var d = "HelloWorld"

	fmt.Println(a, b, c, d)
}
