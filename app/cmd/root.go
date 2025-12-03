package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "config",
	Short: "root command",
	Long:  "root command",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== 配置文件信息 ===")
		fmt.Printf("配置文件路径: %s\n", viper.ConfigFileUsed())
		fmt.Println()

		fmt.Println("数据库配置:")
		fmt.Printf("  DSN: %s\n", viper.GetString("db.dsn"))
		fmt.Println()

		fmt.Println("监控配置:")
		fmt.Printf("  PProf 启用: %v\n", viper.GetBool("monitor.pprof_enable"))
		fmt.Printf("  PProf 端口: %d\n", viper.GetInt("monitor.pprof_port"))
		fmt.Println()

		fmt.Println("Redis配置:")
		fmt.Printf("  主机: %s\n", viper.GetString("redis.host"))
		fmt.Printf("  端口: %d\n", viper.GetInt("redis.port"))
		fmt.Printf("  密码: %s\n", viper.GetString("redis.password"))
		fmt.Printf("  数据库: %d\n", viper.GetInt("redis.db"))
		fmt.Println()

		fmt.Println("日志配置:")
		fmt.Printf("  压缩: %v\n", viper.GetBool("log.compress"))
		fmt.Printf("  保留天数: %d\n", viper.GetInt("log.keep_days"))
		fmt.Printf("  日志级别: %s\n", viper.GetString("log.level"))
		fmt.Printf("  日志模式: %s\n", viper.GetString("log.mode"))
		fmt.Printf("  日志路径: %s\n", viper.GetString("log.path"))
		fmt.Printf("  服务名称: %s\n", viper.GetString("log.service_name"))
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&cfgFile, "config", "c", "./config/config.yaml", "config file (default is $HOME/.config.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		// 从flag中获取配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 主目录 /Users/$HOME$
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// 从主目录下搜索后缀名为 ".toml" 文件 (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("StakeSync")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}
