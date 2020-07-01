package models

// Datasets ...
type Datasets struct {
	Items []Dataset `json:"items"`
}

// Dataset ...
type Dataset struct {
	Alias       string `json:"alias"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Title       string `json:"title"`
}
