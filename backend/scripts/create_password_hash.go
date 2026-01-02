package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func Generate(password string) {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	fmt.Println(string(hash))
}

// Use this script if you want to manually add a user to the DB and need a hash of your password
func main() {
	fmt.Print("Enter password: ")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		fmt.Println("Error reading input:", err)
		os.Exit(1)
	}
	Generate(input)
}
