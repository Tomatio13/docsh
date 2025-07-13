package main

import (
	"fmt"
	"os"

	"cherrysh/shell"
)

func main() {
	fmt.Println("🌸 Cherry Shell v1.0.0 - Beautiful & Simple Shell 🌸")
	fmt.Println("Named after the cherry blossom shell (Sakura-gai) - small but beautiful")
	fmt.Println("Type 'exit' to quit")
	
	sh := shell.NewShell()
	if err := sh.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}