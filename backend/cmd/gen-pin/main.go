package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/joshu-sajeev/paisa/internal/security"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter PIN (6 digits): ")
	pin, _ := reader.ReadString('\n')
	pin = strings.TrimSpace(pin)

	hash, err := security.HashPIN(pin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(hash)
}
