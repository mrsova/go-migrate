package config

var config *Config

type Config struct {
	Database DatabaseConfig `toml:"database"`
	Migrate  MigrateConfig  `toml:"migrate"`
}
type MigrateConfig struct {
	Dir       string `toml:"dir"`
	TableName string `toml:"tablename"`
}
type DatabaseConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Host     string `toml:"host"`
	Port     string `toml:"port"`
	Database string `toml:"database"`
}

func NewConfig() *Config {
	return &Config{}
}
func SetConfig(c *Config) {
	config = c
}

func GetConfig() Config {
	return *config
}
