package mpvctl

type ErrorResponse struct {
	Error string `json:"error"`
	Ok    bool   `json:"ok"`
}

type ResponseModel struct {
	Ok bool `json:"ok"`
}
