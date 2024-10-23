package models

type Request struct {
	Name     string    `json:"name"`
	Method   string    `json:"method"`
	Password string    `json:"password"`
	Port     int       `json:"port"`
	Limit    DataLimit `json:"limit"`
}

type DataLimit struct {
	Bytes int `json:"bytes"`
}
