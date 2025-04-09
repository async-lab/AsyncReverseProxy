package config

type ConfigItemRemote struct {
	Name  string `toml:"name"`
	Addr  string `toml:"addr"`
	Token string `toml:"token"`
}

type ConfigItemProxy struct {
	Name     string   `toml:"name"`
	Proto    string   `toml:"proto"`
	Remotes  []string `toml:"remotes"`
	Backend  string   `toml:"backend"`
	Frontend string   `toml:"frontend"`
	Priority uint32   `toml:"priority"`
	Weight   uint32   `toml:"weight"`
}

type ConfigClient struct {
	Remotes []ConfigItemRemote `toml:"remotes"`
	Proxies []ConfigItemProxy  `toml:"proxies"`
}
