package main

import (
	"containerproject/pkg/container"
	"fmt"
	"os"
)

func main() {

	// if err := container.loadAllContainers(); err != nil {
    //     fmt.Printf("Error loading containers: %v\n", err)
    //     os.Exit(1)
    // }

    if len(os.Args) < 2 {
        fmt.Println("Usage: go run main.go <command> [args...]")
        os.Exit(1)
    }

    switch os.Args[1] {
    case "create":
        if len(os.Args) < 4 {
            fmt.Println("Usage: go run main.go create <container_id> <command>")
            os.Exit(1)
        }
        container.Create(os.Args[2], os.Args[3])
    case "run":
        container.Run()
    case "child":
        container.Child()
    default:
        fmt.Println("Unknown command")
        os.Exit(1)
    }
}
