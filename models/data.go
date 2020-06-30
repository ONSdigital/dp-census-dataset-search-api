package models

// Datasets ...
type Datasets struct {
	Items []Dataset `json:"items"`
}

// Dataset ...
type Dataset struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Name        string `json:"name"`
	Title       string `json:"title"`
}
