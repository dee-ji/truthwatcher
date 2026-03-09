package main

import (
	"context"
	"log"

	"github.com/truthwatcher/truthwatcher/internal/app"
)

func main() {
	if err := app.Run(context.Background(), "radiant"); err != nil {
		log.Fatal(err)
	}
}
