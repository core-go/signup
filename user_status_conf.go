package signup

type UserStatusConf struct {
	Registered string `mapstructure:"registered"`
	Verifying  string `mapstructure:"verifying"`
	Activated  string `mapstructure:"activated"`
}
