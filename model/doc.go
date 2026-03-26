package model

// Doc represents a document in the search index.
type Doc struct {
	FID     int    `json:"fid"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Cate    string `json:"cate"`
	CTime   string `json:"ctime"`
}
