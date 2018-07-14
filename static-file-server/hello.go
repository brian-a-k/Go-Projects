package main

import "fmt"

func main() {
	var name string
	fmt.Println("Enter Your Name: ")
	fmt.Scan(&name)
	fmt.Printf("Hello, %s!", name)
}
