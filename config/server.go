package config

type ConfigItemServer struct {
	ListenAddress string `toml:"listen_address"`
	Password      string `toml:"password"`
	Token         string `toml:"token"`
}

type ConfigServer struct {
	Server *ConfigItemServer `toml:"server"`
}
