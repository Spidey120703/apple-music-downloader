package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type CliConfig struct {
	Storage    StorageSettings  `mapstructure:"storage"     json:"storage"`
	Network    NetworkSettings  `mapstructure:"network"     json:"network"`
	AppleMusic AppleMusicConfig `mapstructure:"apple_music" json:"apple_music"`
}

type StorageSettings struct {
	TargetPath     string `mapstructure:"target_path"      json:"target_path"`
	TempPath       string `mapstructure:"temp_path"        json:"temp_path"`
	UseOriginalExt bool   `mapstructure:"use_original_ext" json:"use_original_ext"`
}

type NetworkSettings struct {
	FairPlay   FairPlayConfig `mapstructure:"fairplay"    json:"fairplay"`
	HTTP       HttpConfig     `mapstructure:"http"        json:"http"`
	NumThreads int            `mapstructure:"num_threads" json:"num_threads"`
}

type FairPlayConfig struct {
	ServerAddr string `mapstructure:"server_addr" json:"server_addr"`
}

type HttpConfig struct {
	UserAgent string `mapstructure:"user_agent" json:"user_agent"`
	Origin    string `mapstructure:"origin"     json:"origin"`
	Referer   string `mapstructure:"referer"    json:"referer"`
}

type AppleMusicConfig struct {
	Storefront     string `mapstructure:"storefront"       json:"storefront"`
	MediaUserToken string `mapstructure:"media_user_token" json:"media_user_token"`
	Language       string `mapstructure:"language"         json:"language"`
}

var config CliConfig

func LoadConfig() (err error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("storage.target_path", DefaultTargetPath)
	viper.SetDefault("storage.temp_path", DefaultTempPath)
	viper.SetDefault("storage.use_original_ext", true)
	viper.SetDefault("network.fairplay.server_addr", DefaultFairPlayServerAddr)
	viper.SetDefault("network.http.user_agent", DefaultUserAgent)
	viper.SetDefault("network.http.origin", DefaultOrigin)
	viper.SetDefault("network.http.referer", DefaultReferer)
	viper.SetDefault("network.num_threads", DefaultNumThreads)
	viper.SetDefault("apple_music.storefront", DefaultStorefront)
	viper.SetDefault("apple_music.language", DefaultAMLanguage)

	if err = viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("error to read config file: %w", err)
		}
	}

	if err = viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return
}

func Get() CliConfig {
	return config
}
