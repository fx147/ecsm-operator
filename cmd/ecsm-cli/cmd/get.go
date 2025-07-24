// file: cmd/ecsm-cli/cmd/get.go

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/fx147/ecsm-operator/internal/ecsm-cli/util"
	"github.com/fx147/ecsm-operator/pkg/ecsm-client/clientset"
	"github.com/spf13/cobra"
)

// newGetCmd 创建 get 命令
func newGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [resource]",
		Short: "Display one or many resources",
		Long:  `Prints a table of the most important information about the specified resources.`,
		Run: func(cmd *cobra.Command, args []string) {
			// 如果只输入 "ecsm-cli get"，就打印帮助信息
			cmd.Help()
		},
	}

	// 添加 get 的子命令
	cmd.AddCommand(newGetNodesCmd())
	cmd.AddCommand(newGetImagesCmd())
	cmd.AddCommand(newGetServicesCmd())
	cmd.AddCommand(newGetContainersCmd())

	return cmd
}

// newGetNodesCmd 创建 "get nodes" 子命令
func newGetNodesCmd() *cobra.Command {
	var pageSize int
	var pageNum int
	var nameFilter string
	var basicInfo bool
	cmd := &cobra.Command{
		Use:     "nodes",
		Short:   "Display a list of nodes",
		Aliases: []string{"node", "no"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := util.NewClientsetFromFlags()
			if err != nil {
				return err
			}

			opts := clientset.NodeListOptions{
				PageSize:  pageSize,
				Name:      nameFilter,
				BasicInfo: basicInfo,
			}

			var nodesToPrint []clientset.NodeInfo

			// --- 核心修复 ---
			// 通过检查用户是否在命令行中明确设置了 "page" 标志，
			// 来决定是分页还是获取全部。
			if cmd.Flags().Changed("page") {
				// 用户明确指定了页码，执行分页 List
				opts.PageNum = pageNum
				nodeList, err := cs.Nodes().List(context.Background(), opts)
				if err != nil {
					return err
				}
				nodesToPrint = nodeList.Items
			} else {
				// 默认行为：获取所有节点
				allNodes, err := cs.Nodes().ListAll(context.Background(), opts)
				if err != nil {
					return err
				}
				nodesToPrint = allNodes
			}

			if len(nodesToPrint) > 0 {
				util.PrintNodesTable(os.Stdout, nodesToPrint)
			} else {
				fmt.Println("No nodes found.")
			}
			return nil
		},
	}

	// 标志定义保持不变
	cmd.Flags().IntVarP(&pageNum, "page", "p", 1, "Page number to retrieve (disables listing all pages)")
	cmd.Flags().IntVarP(&pageSize, "page-size", "s", 100, "Number of items per page (used for both single and all-page listing)")
	cmd.Flags().StringVarP(&nameFilter, "name", "n", "", "Filter nodes by name (fuzzy match)")
	cmd.Flags().BoolVar(&basicInfo, "basic", false, "Display basic information only")

	return cmd
}

// newGetImagesCmd 创建 "get images" 子命令
func newGetImagesCmd() *cobra.Command {
	// 定义 get images 命令的本地标志
	var registryID, nameFilter, osFilter, authorFilter string
	var pageNum, pageSize int
	var listAll bool

	cmd := &cobra.Command{
		Use:     "images",
		Short:   "Display a list of images",
		Aliases: []string{"image", "img"},
		// 我们不希望 get images 后面跟任何参数
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. 创建客户端
			cs, err := util.NewClientsetFromFlags()
			if err != nil {
				return fmt.Errorf("failed to create clientset: %w", err)
			}

			// 2. 准备请求参数
			opts := clientset.ImageListOptions{
				RegistryID: registryID,
				PageSize:   pageSize,
				Name:       nameFilter,
				OS:         osFilter,
				Author:     authorFilter,
			}

			var imagesToPrint []clientset.ImageListItem

			// 3. 根据标志决定是分页还是获取全部
			if listAll {
				allImages, err := cs.Images().ListAll(context.Background(), opts)
				if err != nil {
					return err
				}
				imagesToPrint = allImages
			} else {
				opts.PageNum = pageNum
				imageList, err := cs.Images().List(context.Background(), opts)
				if err != nil {
					return err
				}
				imagesToPrint = imageList.Items
			}

			// 4. 使用 printer 打印结果
			if len(imagesToPrint) > 0 {
				util.PrintImagesTable(os.Stdout, imagesToPrint)
			} else {
				fmt.Println("No images found.")
			}

			return nil
		},
	}

	// 绑定本地标志
	cmd.Flags().StringVar(&registryID, "registry-id", "local", "The ID of the registry to query")
	cmd.Flags().StringVar(&nameFilter, "name", "", "Filter images by name")
	cmd.Flags().StringVar(&osFilter, "os", "", "Filter images by OS (e.g., 'linux', 'sylixos')")
	cmd.Flags().StringVar(&authorFilter, "author", "", "Filter images by author")

	cmd.Flags().BoolVarP(&listAll, "all", "A", true, "List all pages of images (default behavior)")
	cmd.Flags().IntVar(&pageNum, "page", 1, "Page number to retrieve (if --all=false)")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "Number of items per page")

	return cmd
}

// newGetServicesCmd 创建 "get services" 子命令
func newGetServicesCmd() *cobra.Command {
	// 定义 get services 命令的本地标志
	var pageNum, pageSize int
	var nameFilter, imageID, nodeID, labelFilter string
	var listAll bool

	cmd := &cobra.Command{
		Use:     "services",
		Short:   "Display a list of services",
		Aliases: []string{"service", "svc"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := util.NewClientsetFromFlags()
			if err != nil {
				return err
			}

			opts := clientset.ListServicesOptions{
				PageSize: pageSize,
				Name:     nameFilter,
				ImageID:  imageID,
				NodeID:   nodeID,
				Label:    labelFilter,
			}

			var servicesToPrint []clientset.ProvisionListRow

			if listAll {
				allServices, err := cs.Services().ListAll(context.Background(), opts)
				if err != nil {
					return err
				}
				servicesToPrint = allServices
			} else {
				opts.PageNum = pageNum
				serviceList, err := cs.Services().List(context.Background(), opts)
				if err != nil {
					return err
				}
				servicesToPrint = serviceList.Items
			}

			if len(servicesToPrint) > 0 {
				util.PrintServicesTable(os.Stdout, servicesToPrint)
			} else {
				fmt.Println("No services found.")
			}

			return nil
		},
	}

	// 绑定本地标志
	cmd.Flags().StringVarP(&nameFilter, "name", "n", "", "Filter services by name (fuzzy match)")
	cmd.Flags().StringVar(&imageID, "image-id", "", "Filter services by image ID")
	cmd.Flags().StringVar(&nodeID, "node-id", "", "Filter services by node ID")
	cmd.Flags().StringVarP(&labelFilter, "label", "l", "", "Filter services by path label (fuzzy match)")

	cmd.Flags().BoolVarP(&listAll, "all", "A", true, "List all pages of services (default behavior)")
	cmd.Flags().IntVar(&pageNum, "page", 1, "Page number to retrieve (if --all=false)")
	cmd.Flags().IntVar(&pageSize, "page-size", 100, "Number of items per page")

	return cmd
}

// newGetContainersCmd 创建 "get containers" 子命令
func newGetContainersCmd() *cobra.Command {
	// 定义 get containers 命令的本地标志
	var serviceFilter string
	var nodeFilter string
	var listAll bool

	cmd := &cobra.Command{
		Use:     "containers",
		Short:   "Display a list of containers",
		Aliases: []string{"container", "co"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := util.NewClientsetFromFlags()
			if err != nil {
				return err
			}
			ctx := context.Background()

			var containersToPrint []clientset.ContainerInfo

			// --- 核心逻辑：根据标志决定如何获取容器 ---
			if serviceFilter != "" {
				// 按服务过滤
				// 1. 智能查找 Service ID
				serviceOpts := clientset.ListServicesOptions{Name: serviceFilter}
				allServices, err := cs.Services().ListAll(ctx, serviceOpts)
				if err != nil {
					return fmt.Errorf("failed to list services to find service '%s': %w", serviceFilter, err)
				}

				if len(allServices) == 0 {
					return fmt.Errorf("service '%s' not found", serviceFilter)
				}

				var targetServiceIDs []string
				// List API 的 name 可能是模糊匹配，所以我们需要收集所有匹配项
				for _, svc := range allServices {
					targetServiceIDs = append(targetServiceIDs, svc.ID)
				}

				// 2. 使用找到的 ID 列表来获取容器
				containerOpts := clientset.ListContainersByServiceOptions{ServiceIDs: targetServiceIDs}
				containersToPrint, err = cs.Containers().ListAllByService(ctx, containerOpts)
				if err != nil {
					return fmt.Errorf("failed to list containers for service(s) '%s': %w", serviceFilter, err)
				}

			} else if nodeFilter != "" {
				// --- 按节点过滤 (已实现) ---

				// 1. 智能查找 Node ID
				nodeOpts := clientset.NodeListOptions{Name: nodeFilter}
				allNodes, err := cs.Nodes().ListAll(ctx, nodeOpts)
				if err != nil {
					return fmt.Errorf("failed to list nodes to find node '%s': %w", nodeFilter, err)
				}

				if len(allNodes) == 0 {
					return fmt.Errorf("node '%s' not found", nodeFilter)
				}

				var targetNodeIDs []string
				for _, node := range allNodes {
					targetNodeIDs = append(targetNodeIDs, node.ID)
				}

				// 2. 使用找到的 ID 列表来获取容器
				// (我们需要一个新的 ListAllContainersByNode 辅助函数)
				containerOpts := clientset.ListContainersByNodeOptions{NodeIDs: targetNodeIDs}
				containersToPrint, err = cs.Containers().ListAllByNode(ctx, containerOpts)
				if err != nil {
					return fmt.Errorf("failed to list containers for node(s) '%s': %w", nodeFilter, err)
				}

			} else {
				// 获取所有容器：遍历所有服务
				allServices, err := cs.Services().ListAll(ctx, clientset.ListServicesOptions{})
				if err != nil {
					return fmt.Errorf("failed to list services: %w", err)
				}

				var allServiceIDs []string
				for _, svc := range allServices {
					allServiceIDs = append(allServiceIDs, svc.ID)
				}

				if len(allServiceIDs) > 0 {
					opts := clientset.ListContainersByServiceOptions{ServiceIDs: allServiceIDs}
					containersToPrint, err = cs.Containers().ListAllByService(ctx, opts)
					if err != nil {
						return fmt.Errorf("failed to list containers: %w", err)
					}
				}
			}

			// 打印结果
			if len(containersToPrint) > 0 {
				util.PrintContainersTable(os.Stdout, containersToPrint)
			} else {
				fmt.Println("No containers found.")
			}

			return nil
		},
	}

	// 绑定本地标志
	cmd.Flags().StringVarP(&serviceFilter, "service", "s", "", "Filter containers by service name or ID")
	cmd.Flags().StringVarP(&nodeFilter, "node", "n", "", "Filter containers by node name or ID")

	cmd.Flags().BoolVarP(&listAll, "all", "A", true, "List all pages of containers (default behavior)")

	return cmd
}
