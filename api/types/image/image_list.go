package image

type ImageListItem struct {
	Name        string `json:"name"`
	CreatedTime string `json:"created_time"`
	Size        int64  `json:"size"`
}
