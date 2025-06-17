package installer

type AppConfig struct {
	Name string   `json:"name"`
	URL  string   `json:"url"`
	Path string   `json:"path"`
	Cmd  string   `json:"cmd"`
	Args []string `json:"args"`
	OS   string   `json:"os"`
}
