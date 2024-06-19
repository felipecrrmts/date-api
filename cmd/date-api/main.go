package main

import (
	"log"

	"github.com/muzzapp/date-api/internal/app"
)

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	if err = a.Start(); err != nil {
		return
	}
}
