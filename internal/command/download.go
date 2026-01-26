package command

// DownloadProgress represents download progress
type DownloadProgress struct {
	Name       string
	Downloaded int64
	Total      int64
	Done       bool
	Error      error
}
