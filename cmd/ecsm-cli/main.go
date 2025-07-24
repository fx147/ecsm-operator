// file: cmd/ecsm-cli/main.go

package main

import (
	"flag"

	"github.com/fx147/ecsm-operator/cmd/ecsm-cli/cmd"
	"k8s.io/klog/v2"
)

func main() {
	// 1. 初始化一个空的 FlagSet
	fs := flag.NewFlagSet("klog", flag.ContinueOnError)
	klog.InitFlags(fs)

	// 2. 将 klog 的标志添加到 cobra 的根命令上
	//    这样 cobra 就能解析 -v, --logtostderr 等 klog 参数了
	cmd.GetRootCmd().PersistentFlags().AddGoFlagSet(fs)

	// 3. 正常执行 cobra
	cmd.Execute()
}
