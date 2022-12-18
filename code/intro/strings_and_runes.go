package main

import "fmt"

func main() {
	fmt.Printf("We have a cool fmt package: %c %c %c!\n", '\u2713', '☢', '😎')
	fmt.Printf("こんにちは世界!\n")
	fmt.Printf(`With backticks we can ignore backslashes like \n!`)
}
