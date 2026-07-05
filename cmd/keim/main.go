package main

import (
	"fmt"

	"keim/internal/validator"
)

func main() {
	fmt.Println("--- Probando mi validador en vivo ---")

	if err := validator.Validate("./case"); err != nil {
		fmt.Println("[Validator]: ", err)
	}
	
	if err := validator.Validate("./cmd"); err != nil {
		fmt.Println("[Validator]: ", err)
	}
}