package models

import (
	"log"
	"time"
)

const (
	copenhagenLocation = "Europe/Copenhagen"
)

var (
	// copenhagen is the location for the time zone database.
	copenhagen *time.Location
)

func LoadLocation() {
	cph, err := time.LoadLocation(copenhagenLocation)
	if err != nil {
		log.Fatalf("failed to load location: %v", err)
	}

	copenhagen = cph
}
