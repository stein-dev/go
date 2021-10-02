package libinject

var (
	Loop          = true
	DefaultConfig = &Config{
		ListenPort:    "8989",
		Payload: "HTTP/1.1 200 [crlf]Host: bing.com [lf][lf]",
		ProxyHost:     "127.0.0.1",
		ProxyPort:     "8080",
		Username: "",
		Password: "",
		Timer:     0,
		FileName: "config/config.ovpn",
		AuthFileName: "config/config.auth",
	}
)

type Config struct {
	ListenPort	  string
	Payload	string
	ProxyHost     string
	ProxyPort     string
	Username string
	Password string
	Timer 	 int
	FileName string
	AuthFileName string
}

type Inject struct {
	Config   *Config
}

