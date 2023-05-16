package server

type Server struct {
	Ip     string `json:"ip"`
	Port   string `json:"port"`
	Peers  int    `json:"peers"`
	Policy string `json:"policy"`
}
