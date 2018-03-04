package main

// Meta - Most json responses contain a metadata object
type meta struct {
	Publisher   string `json:"publisher"`
	License     string `json:"license"`
	Version     string `json:"version"`
	ResultLimit uint32 `json:"resultLimit"`
}

func newMeta(limit int) meta {
	metaData := meta{}
	metaData.License = "Creative Commons"
	metaData.Publisher = "Kent Network"
	metaData.Version = "0.1"
	metaData.ResultLimit = resultLimit
	return metaData
}
