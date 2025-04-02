package store

import (
	"log"

	"community-climate-justice-archive/data"
)

func GetThemes() []data.Theme {
	log.Println("Getting themes")

	return []data.Theme{
		{
			Title: "people",
		},
		{
			Title: "planet",
		},
		{
			Title: "architecture",
		},
		{
			Title: "old",
		},
	}
}
