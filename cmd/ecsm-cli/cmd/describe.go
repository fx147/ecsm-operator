// file: cmd/ecsm-cli/cmd/describe.go

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fx147/ecsm-operator/internal/ecsm-cli/util"
	"github.com/fx147/ecsm-operator/pkg/ecsm-client/clientset"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

// newDescribeCmd 创建 describe 命令
func newDescribeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe [resource] [name]",
		Short: "Show detailed information about a resource",
		Long:  `Prints a detailed description of the specified resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// 添加 describe 的子命令
	cmd.AddCommand(newDescribeImageCmd())
	cmd.AddCommand(newDescribeNodeCmd()) // 未来在这里添加
	cmd.AddCommand(newDescribeServiceCmd())
	cmd.AddCommand(newDescribeContainerCmd())

	return cmd
}

// newDescribeNodeCmd 创建 describe node 子命令
func newDescribeNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node <NODE_NAME_OR_ID>",
		Short: "Show detailed information about a specific node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := util.NewClientsetFromFlags()
			if err != nil {
				return err
			}

			identifier := args[0]
			ctx := context.Background()

			// --- 核心逻辑：智能查找 Node ID ---
			var targetNodeID string

			// 1. 获取所有节点，以便查找
			allNodes, err := cs.Nodes().ListAll(ctx, clientset.NodeListOptions{})
			if err != nil {
				return fmt.Errorf("failed to list nodes to find identifier: %w", err)
			}

			// 2. 尝试将 identifier 作为 ID 直接匹配
			var foundByName []*clientset.NodeInfo
			for i, node := range allNodes {
				if node.ID == identifier {
					targetNodeID = identifier
					break
				}
				if node.Name == identifier {
					foundByName = append(foundByName, &allNodes[i])
				}
			}

			// 3. 如果通过 ID 没找到，则检查按名称查找的结果
			if targetNodeID == "" {
				if len(foundByName) == 0 {
					return fmt.Errorf("node '%s' not found", identifier)
				}
				if len(foundByName) > 1 {
					// --- 关键的用户友好提示 ---
					var ids []string
					for _, n := range foundByName {
						ids = append(ids, n.ID)
					}
					return fmt.Errorf("multiple nodes found with name '%s', please use one of the following IDs: %v", identifier, ids)
				}
				// 名称唯一，查找成功
				targetNodeID = foundByName[0].ID
			}

			// --- 数据聚合 ---
			// 4. 现在我们有了唯一的 targetNodeID，可以进行所有查询
			nodeView, err := cs.Nodes().GetNodeView(ctx, targetNodeID)
			if err != nil {
				return fmt.Errorf("failed to get node view: %w", err)
			}

			metricsList, err := cs.Nodes().GetNodeMetrics(ctx, clientset.NodeMetricsOptions{NodeID: targetNodeID, Instant: true})
			if err != nil {
				return fmt.Errorf("failed to get node metrics: %w", err)
			}
			if len(metricsList) == 0 {
				return fmt.Errorf("no metrics returned for node '%s'", identifier)
			}

			// --- 打印 ---
			// 5. 将聚合后的数据传递给打印机
			util.PrintNodeDetails(os.Stdout, nodeView, &metricsList[0])
			return nil
		},
	}
	return cmd
}

// newDescribeImageCmd 创建 "describe image" 子命令
func newDescribeImageCmd() *cobra.Command {
	var registryID string

	cmd := &cobra.Command{
		Use:     "image <NAME@TAG[#OS]>",
		Short:   "Show detailed information about a specific image",
		Aliases: []string{"img"},
		// 确保用户必须提供且只提供一个参数
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. 获取客户端
			cs, err := util.NewClientsetFromFlags()
			if err != nil {
				return err
			}

			// 2. 获取参数
			imageRef := args[0]

			// 3. 调用我们之前设计好的高级辅助方法
			details, err := cs.Images().GetDetailsByRef(context.Background(), registryID, imageRef)
			if err != nil {
				return err
			}

			// 4. 将获取到的详情对象传递给专门的打印机
			util.PrintImageDetails(os.Stdout, details)
			return nil
		},
	}

	cmd.Flags().StringVar(&registryID, "registry-id", "local", "The ID of the registry to query")
	return cmd
}

// newDescribeServiceCmd 创建 "describe service" 子命令
func newDescribeServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service <SERVICE_NAME_OR_ID>",
		Short:   "Show detailed information about a specific service",
		Aliases: []string{"svc"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := util.NewClientsetFromFlags()
			if err != nil {
				return err
			}

			identifier := args[0]
			ctx := context.Background()

			// --- 1. 智能查找 Service ID ---
			allServices, err := cs.Services().ListAll(ctx, clientset.ListServicesOptions{})
			if err != nil {
				return fmt.Errorf("failed to list services: %w", err)
			}

			var targetServiceID string
			var foundByName []*clientset.ProvisionListRow
			for i, svc := range allServices {
				if svc.ID == identifier {
					targetServiceID = identifier
					break
				}
				if svc.Name == identifier {
					foundByName = append(foundByName, &allServices[i])
				}
			}

			if targetServiceID == "" {
				if len(foundByName) == 0 {
					return fmt.Errorf("service '%s' not found", identifier)
				}
				if len(foundByName) > 1 {
					var ids []string
					for _, s := range foundByName {
						ids = append(ids, s.ID)
					}
					return fmt.Errorf("multiple services found with name '%s', please use one of the following IDs: %v", identifier, ids)
				}
				targetServiceID = foundByName[0].ID
			}

			// --- 2. 数据聚合 ---
			// 主调用: 获取服务详情
			serviceDetails, err := cs.Services().Get(ctx, targetServiceID)
			if err != nil {
				return fmt.Errorf("failed to get service details: %w", err)
			}

			// 辅助调用: 获取容器列表
			containerList, err := cs.Containers().ListByService(ctx, clientset.ListContainersByServiceOptions{
				PageNum:    1,
				PageSize:   1000, // 获取该服务下的所有容器
				ServiceIDs: []string{targetServiceID},
			})
			if err != nil {
				return fmt.Errorf("failed to list containers for service: %w", err)
			}

			// --- 3. 打印 ---
			util.PrintServiceDetails(os.Stdout, serviceDetails, containerList.Items)
			return nil
		},
	}
	return cmd
}

// newDescribeContainerCmd 创建 "describe container" 子命令
func newDescribeContainerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "container <CONTAINER_NAME>",
		Short:   "Show detailed information about a specific container",
		Aliases: []string{"co"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := util.NewClientsetFromFlags()
			if err != nil {
				return err
			}

			containerName := args[0]
			ctx := context.Background()

			// 1. 使用高级辅助函数，通过 Name 智能查找容器
			containerInfo, err := cs.Containers().GetByName(ctx, cs.Services(), containerName)
			if err != nil {
				return err
			}

			// 2. 获取操作历史
			historyOpts := clientset.ContainerHistoryOptions{
				TaskID:   containerInfo.TaskID,
				PageNum:  1,
				PageSize: 100, // 获取最近100条历史
			}
			historyList, err := cs.Containers().GetHistory(ctx, historyOpts)
			if err != nil {
				// 如果获取历史失败，只打印一个警告，而不是让整个命令失败
				klog.Warningf("Could not retrieve action history for container %s: %v", containerName, err)
			}

			// 3. 打印聚合后的信息
			util.PrintContainerDetails(os.Stdout, containerInfo, historyList)
			return nil
		},
	}
	return cmd
}
