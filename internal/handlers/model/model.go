package model

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

type Response struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	Port      int    `json:"port"`
	Method    string `json:"method"`
	AccessURL string `json:"accessUrl"`
}
