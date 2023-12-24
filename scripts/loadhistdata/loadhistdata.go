package main

import (
	"fmt"
	"os"
)

func main() {
	var filePath string
	var ticker string
	var accessKey string

	fmt.Print("Enter the path to your txt file: ")
	fmt.Scanln(&filePath)
	fmt.Print("What company ticker does this data apply to?: ")
	fmt.Scanln(&ticker)
	fmt.Print("Enter your API access key: ")
	fmt.Scanln(&accessKey)

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	// Convert content to a string
	text := string(content)

	fmt.Println("File content:\n", text)
}
