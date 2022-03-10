package config

type ServerConfig struct {
	Host string
	Port string
}

func (srv ServerConfig) Addr() string {
	return /*srv.Host + */":" + srv.Port
}

type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
}

type Config struct {
	Proxy    ServerConfig
	Repeater ServerConfig
	DB       DBConfig
}
