package config

type ConfigItemServer struct {
	ListenAddress string `toml:"listen_address"`
	Password      string `toml:"password"`
}

type ConfigServer struct {
	Server *ConfigItemServer `toml:"server"`
}
