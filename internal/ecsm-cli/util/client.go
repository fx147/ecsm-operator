// file: internal/ecsm-cli/util/client.go

package util

import (
	"fmt"

	"github.com/fx147/ecsm-operator/pkg/ecsm-client/clientset"
	"github.com/spf13/viper"
)

// NewClientsetFromFlags 从 viper 中读取全局标志，并创建一个新的 ecsm-client Clientset。
func NewClientsetFromFlags() (*clientset.Clientset, error) {
	host := viper.GetString("host")
	port := viper.GetString("port")
	protocol := viper.GetString("protocol")

	if host == "" || port == "" || protocol == "" {
		return nil, fmt.Errorf("host, port, and protocol must be specified")
	}

	return clientset.NewClientset(protocol, host, port) // http.Client 先用 nil
}
