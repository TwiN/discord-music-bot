package core

type Media struct {
	Title    string
	FilePath string
}

func NewMedia(title, filePath string) *Media {
	return &Media{
		Title:    title,
		FilePath: filePath,
	}
}
