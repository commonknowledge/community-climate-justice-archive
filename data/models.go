package data

type Page struct {
	Title       string
	Description string
	Themes      []Theme
	Types       []Type
	Images      []StoryImage
}

type Theme struct {
	Title string
}

type Type struct {
	Title string
	URL   string
}

type StoryImage struct {
	Filename        string
	AlternativeText string
}

type Story struct {
	ID                      string
	CreatedTime             string
	Finding                 string
	HighStExperiment        string
	WhatWasIsIf             string
	Image                   string
	SourceImage             string
	Location                string
	StartDateTime           string
	EndDateTime             string
	Season                  string
	Weather                 string
	StreetDetectoristClue   string
	Themes                  string
	Experience              string
	TimeSpan                string
	OtherComments           string
	Type                    string
	PersonFinder            string
	MapCache                string
	MapSize                 string
	Created                 string
	StreetDetectoristMapURL string
	OtherTheme              string
	OtherWeather            string
	ShareStatus             string
	PostDate                string
	TwitterText             string
	CharacterCount          string
	InstaText               string
	InstaCount              string
	InstaImage              string
	ImageData               []byte
}
