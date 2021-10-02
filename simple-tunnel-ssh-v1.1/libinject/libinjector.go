package libinject

var (
	Loop          = true
	DefaultConfig = &Config{
		ListenPort:    "8989",
		Payload: "HTTP//1.1 200 [crlf]Host: wattpad.com [lf][lf][lf]",
		ProxyHost:     "127.0.0.1",
		ProxyPort:     "8080",
		Username: "",
		Password: "",
		Timer:     0,
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
}

type Inject struct {
	Config   *Config
}

