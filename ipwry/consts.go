package ipwry

import (
	"net/http"
	"os"
)

const (
	IndexLen      = 7
	RedirectMode1 = 0x01
	RedirectMode2 = 0x02
)

type ResultQQwry struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Area    string `json:"area"`
}

type fileData struct {
	Data  []byte
	Path  *os.File
	IPNum int64
}

type QQwry struct {
	Data   *fileData
	Offset int64
}

type Response struct {
	r *http.Request
	w http.ResponseWriter
}
