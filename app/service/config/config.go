package config

import (
	"github.com/spf13/viper"
	"github.com/zeromicro/go-zero/core/logx"
)

// Config 配置结构体
type Config struct {
	DB              *DBConfig      `toml:"db" mapstructure:"db" json:"db"`
	Monitor         *MonitorConfig `toml:"monitor" mapstructure:"monitor" json:"monitor"`
	Log             logx.LogConf   `toml:"log" mapstructure:"log" json:"log"`
	Redis           *RedisConfig   `toml:"redis" mapstructure:"redis" json:"redis"`
	ChainID         int64          `toml:"chainId" mapstructure:"chainId" json:"chainId"`
	RPCURL          string         `toml:"rpcUrl" mapstructure:"rpcUrl" json:"rpcUrl"`
	ContractABI     string         `toml:"contractAbi" mapstructure:"contractAbi" json:"contractAbi"`
	ContractAddress string         `toml:"contractAddress" mapstructure:"contractAddress" json:"contractAddress"`
}

// DBConfig 数据库配置
type DBConfig struct {
	DSN string `toml:"dsn" mapstructure:"dsn" json:"dsn"`
}

// MonitorConfig 监控配置
type MonitorConfig struct {
	PprofEnable bool  `toml:"pprof_enable" mapstructure:"pprof_enable" json:"pprof_enable"`
	PprofPort   int64 `toml:"pprof_port" mapstructure:"pprof_port" json:"pprof_port"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `toml:"host" mapstructure:"host" json:"host"`
	Port     int    `toml:"port" mapstructure:"port" json:"port"`
	Password string `toml:"password" mapstructure:"password" json:"password"`
	DB       int    `toml:"db" mapstructure:"db" json:"db"`
}

func UnmarshalCmdConfig() (*Config, error) {
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var c Config

	if err := viper.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
