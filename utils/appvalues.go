package utils

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	v *viper.Viper
}

func NewAppConfig() *AppConfig {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath("./config/")
	err := v.ReadInConfig()
	if err != nil {
		panic(err.Error())
	}

	return &AppConfig{v}
}

func NewCustomAppConfig(v *viper.Viper) *AppConfig {
	return &AppConfig{v}
}

func (val *AppConfig) GetDBName() string {
	return val.v.GetString("DB_NAME")
}

func (val *AppConfig) GetPassDB() string {
	return val.v.GetString("DB_PASS")
}

func (val *AppConfig) GetHTTPDir() string {
	return val.v.GetString("HTTP_DIR")
}

func (val *AppConfig) GetExcelDir() string {
	return val.GetHTTPDir() + "excel/"
}

func (val *AppConfig) GetAdminPass() string {
	return val.v.GetString("ADMIN_PASS")
}

func (val *AppConfig) GetServerPort() string {
	return val.v.GetString("SERVER_PORT")
}
