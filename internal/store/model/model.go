package model

type Server struct {
	IPAddr string
	Port   string
	Key    string
}

type AccessURL struct {
	ID        int
	AccessKey string
	ApiURL    string
}
