package model

// BleveType returns the type name used for indexing in Bleve
func (Page) BleveType() string {
	return "page"
}

// BleveType returns the type name used for indexing in Bleve
func (Folder) BleveType() string {
	return "folder"
}

type ContentMetaWithURL struct {
	ContentMeta
	Url string
}
