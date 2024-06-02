package entity

type AccessURL struct {
	ID        string
	AccessKey string
	ApiURL    string
}

type Server struct {
	IPAddr string
	Port   int
	Key    string
}
