package main

import (
	"log"

	"github.com/fornellas/rrb/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal("%w", err)
		// os.exit(1)
	}
}
