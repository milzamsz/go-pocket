package main

import (
	"fmt"
	"log"
	"os"

	"github.com/milzamsz/go-pocket/internal/app"
	"github.com/pocketbase/pocketbase"

	_ "github.com/milzamsz/go-pocket/migrations"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "seed":
			if err := app.RunLocalSeed(); err != nil {
				log.Fatal(err)
			}
			return
		case "smoke-local":
			baseURL := "http://localhost:8090"
			if len(os.Args) > 2 && os.Args[2] != "" {
				baseURL = os.Args[2]
			}
			if err := app.RunLocalSmoke(baseURL); err != nil {
				log.Fatal(err)
			}
			return
		case "help-seed":
			fmt.Println("Usage:")
			fmt.Println("  go run . seed")
			fmt.Println("  go run . smoke-local [base-url]")
			return
		}
	}

	pb := pocketbase.New()
	if err := app.Bootstrap(pb); err != nil {
		log.Fatal(err)
	}
	if err := pb.Start(); err != nil {
		log.Fatal(err)
	}
}
