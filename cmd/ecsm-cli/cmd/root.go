// file: cmd/ecsm-cli/cmd/root.go

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
)

var (
	// cfgFile 用于存储配置文件的路径
	cfgFile string

	// rootCmd 代表没有调用子命令时的基础命令
	rootCmd = &cobra.Command{
		Use:   "ecsm-cli",
		Short: "A CLI for interacting with the ECSM platform API",
		Long: `ecsm-cli is a command-line interface that allows you to directly
interact with the ECSM (Edge Container Service Mesh) platform.

You can use it to manage resources like nodes, services, and containers
without going through the ecsm-operator's declarative layer.`,
		// 如果用户只输入 ecsm-cli 而没有子命令，就打印帮助信息
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

// Execute 将所有子命令添加到根命令中，并设置标志。
// 这是 main.go 将调用的主函数。
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// 在所有命令执行前运行的初始化函数
	cobra.OnInitialize(initConfig)

	// --- 定义全局持久标志 ---
	// 这些标志对 ecsm-cli 的所有子命令都有效

	// --config 标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ecsm-cli.yaml)")

	// ECSM Server 连接相关的标志
	rootCmd.PersistentFlags().String("host", "localhost", "The host of the ECSM API server")
	rootCmd.PersistentFlags().String("port", "3001", "The port of the ECSM API server")
	rootCmd.PersistentFlags().String("protocol", "http", "The protocol to use (http or https)")

	// --- 将标志与 Viper 绑定 ---
	// 这使得我们可以通过配置文件或环境变量来设置这些值
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("protocol", rootCmd.PersistentFlags().Lookup("protocol"))

	// --- 添加子命令 ---
	// 我们将在这里添加 get, describe 等命令
	rootCmd.AddCommand(newGetCmd())
	rootCmd.AddCommand(newDescribeCmd())
}

// initConfig 读取配置文件和环境变量（如果设置了的话）。
func initConfig() {
	if cfgFile != "" {
		// 使用 --config 标志指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找家目录
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// 1. 先在当前工作目录查找
		viper.AddConfigPath(".")
		// 2. 再在家目录查找
		viper.AddConfigPath(home)

		viper.SetConfigName(".ecsm-cli")
		viper.SetConfigType("yaml")
	}

	// 设置环境变量前缀，例如 ECSMCLI_HOST
	viper.SetEnvPrefix("ECSMCLI")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // 读取匹配的环境变量

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			klog.Warningf("Error reading config file: %v", err)
		}
	}
}

// GetRootCmd 导出 rootCmd 以便 main.go 可以添加 klog 标志
func GetRootCmd() *cobra.Command {
	return rootCmd
}
