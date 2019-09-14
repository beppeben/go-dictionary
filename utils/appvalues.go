package utils

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	v *viper.Viper
}

func NewConfig(path string) *AppConfig {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(path)
	err := v.ReadInConfig()
	if err != nil {
		panic(err.Error())
	}

	return &AppConfig{v}
}

func NewAppConfig() *AppConfig {
	return NewConfig("./config/")
}

func NewCustomAppConfig(v *viper.Viper) *AppConfig {
	return &AppConfig{v}
}

func (val *AppConfig) GetDBName() string {
	return val.v.GetString("DB_NAME")
}

func (val *AppConfig) GetUserDB() string {
	return val.v.GetString("DB_USER")
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

func (val *AppConfig) GetAdminEmails() []string {
	return val.v.GetStringSlice("ADMIN_EMAILS")
}

func (val *AppConfig) GetSlackHook() string {
	return val.v.GetString("SLACK_HOOK")
}
