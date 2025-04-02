package store

import (
	"log"

	"community-climate-justice-archive/data"
)

func GetTypes() []data.Type {
	log.Println("Getting types")

	return []data.Type{
		{Title: "text"},
		{Title: "image"},
		{Title: "video"},
		{Title: "sound"},
		{Title: "textile"},
	}
}
