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

func (val *AppConfig) GetServiceEmail() string {
	return val.v.GetString("EMAIL")
}

func (val *AppConfig) GetEmailPass() string {
	return val.v.GetString("EMAIL_PASS")
}

func (val *AppConfig) GetSMTP() string {
	return val.v.GetString("SMTP")
}

func (val *AppConfig) GetSMTPPort() string {
	return val.v.GetString("SMTP_PORT")
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

func (val *AppConfig) GetServerHost() string {
	return val.v.GetString("SERVER_HOST")
}

func (val *AppConfig) GetServerPort() string {
	return val.v.GetString("SERVER_PORT")
}

func (val *AppConfig) GetServerUrl() string {
	host := val.GetServerHost()
	port := val.GetServerPort()
	if port == "80" {
		return host
	} else {
		return host + ":" + port
	}
}
