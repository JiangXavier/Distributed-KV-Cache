package config

var Conf *Config

type Config struct {
	Services map[string]*Service `yaml:"services"`
}

type Service struct {
	Name        string   `yaml:"name"`
	LoadBalance bool     `yaml:"loadBalance"`
	Addr        []string `yaml:"addr"`
	TTL         int      `yaml:"ttl"`
}
