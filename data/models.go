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
}

type StoryImage struct {
	Filename        string
	AlternativeText string
}
