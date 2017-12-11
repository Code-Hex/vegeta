package common

type ResultJSON struct {
	IsSuccess bool   `json:"is_success"`
	Reason    string `json:"reason"`
}

type PostDataJSON struct {
	Payload    string `json:"payload"`
	Hostname   string `json:"hostname"`
	RemoteAddr string `json:"remote_addr"`
	TagName    string `json:"tag_name"`
}

type TagJSON struct {
	TagName string `json:"tag_name"`
}
