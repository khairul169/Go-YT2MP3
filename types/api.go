package types

type GetVideoInfoRes struct {
	Url       string `json:"url"`
	Slug      string `json:"slug"`
	Thumbnail string `json:"thumbnail"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
}

type CreateTaskBody struct {
	Url       string `json:"url"`
	Slug      string `json:"slug"`
	Thumbnail string `json:"thumbnail"`
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
}
