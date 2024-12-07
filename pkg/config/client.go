package config

type ConfigItemRemoteServer struct {
	Name    string `toml:"name"`
	Address string `toml:"address"`
	Token   string `toml:"token"`
}

type ConfigItemProxy struct {
	Name            string   `toml:"name"`
	RemoteServers   []string `toml:"remote_servers"`
	BackendAddress  string   `toml:"backend_address"`
	FrontendAddress string   `toml:"frontend_address"`
	Priority        uint32   `toml:"priority"`
	Weight          uint32   `toml:"weight"`
}

type ConfigClient struct {
	RemoteServers []*ConfigItemRemoteServer `toml:"remote_servers"`
	Proxies       []*ConfigItemProxy        `toml:"proxies"`
}
