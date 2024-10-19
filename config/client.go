package config

type ConfigItemRemoteServer struct {
	Name    string `toml:"name"`
	Address string `toml:"address"`
	Token string `toml:"token"`
}

type ConfigItemProxy struct {
	Name             string `toml:"name"`
	RemoteServerName string `toml:"remote_server_name"`
	BackendAddress   string `toml:"backend_address"`
	FrontendAddress  string `toml:"frontend_address"`
}

type ConfigClient struct {
	RemoteServers []*ConfigItemRemoteServer `toml:"remote_servers"`
	Proxies       []*ConfigItemProxy        `toml:"proxies"`
}
