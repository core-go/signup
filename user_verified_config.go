package signup

type UserVerifiedConfig struct {
	Secure  bool   `mapstructure:"secure"`
	Domain  string `mapstructure:"domain"`
	Port    int    `mapstructure:"port"`
	Action  string `mapstructure:"action"`
}
