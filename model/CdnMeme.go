package model

type CdnMeme []struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Mtime string `json:"mtime"`
	Size  int    `json:"size,omitempty"`
}
