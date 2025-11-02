package config

type Config struct {
	Host        string
	Port        int
	Requests    int
	Concurrency int
}

func NewConfig(host string, port int) *Config {
	return &Config{
		Host:        host,
		Port:        port,
		Requests:    10,
		Concurrency: 4,
	}
}

func GetDefaultConfig() *Config {
	return &Config{
		Host:        "localhost",
		Port:        5000,
		Requests:    7500,
		Concurrency: 15,
	}
}
