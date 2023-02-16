package config

type configuration struct {
	Tunnel    []*Tunnel    `json:"tunnel"`
	SshConfig []*SshConfig `json:"sshConfig"`
}

type Tunnel struct {
	ListenOn  string `json:"listenOn"`
	ListenAt  string `json:"listenAt"`
	SshAlias  string `json:"sshAlias"`
	ForwardTo string `json:"forwardTo"`
}

type SshConfig struct {
	Alias   string         `json:"alias"`
	Address string         `json:"address"`
	User    string         `json:"user"`
	Auth    *SshConfigAuth `json:"auth"`
}

type SshConfigAuth struct {
	Pass string   `json:"pass"`
	Keys []string `json:"keys"`
}
