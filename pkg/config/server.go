package config

type ConfigItemServer struct {
	Listen string `toml:"listen"`
	Token  string `toml:"token"`
}

type ConfigServer struct {
	Server *ConfigItemServer `toml:"server"`
}
