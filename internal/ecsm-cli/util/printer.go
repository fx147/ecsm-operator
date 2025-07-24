// file: internal/ecsm-cli/util/printer.go

package util

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fx147/ecsm-operator/pkg/ecsm-client/clientset"
)

// PrintNodesTable 将节点列表以表格形式打印到指定的 writer。
func PrintNodesTable(out io.Writer, nodes []clientset.NodeInfo) {
	// 初始化 tabwriter
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// 打印表头
	fmt.Fprintln(w, "NAME\tSTATUS\tADDRESS\tTYPE\tARCH\tCONTAINERS\tCREATED\tUPTIME\tID")

	// 打印每一行
	for _, node := range nodes {
		containerInfo := fmt.Sprintf("%d/%d", node.ContainerEcsmRunning, node.ContainerEcsmTotal)
		uptimeDuration := time.Duration(node.UpTime) * time.Second
		uptimeStr := formatUptime(uptimeDuration)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			node.Name,
			node.Status,
			node.Address,
			node.Type,
			node.Arch,
			containerInfo,
			node.CreatedTime,
			uptimeStr,
			node.ID,
		)
	}
}

// formatUptime 是一个新的辅助函数，用于将时长格式化为 "XdYhZm" 的形式
func formatUptime(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dd%dh%dm", days, hours, minutes)
}

// PrintImagesTable 将镜像列表以表格形式打印到指定的 writer。
func PrintImagesTable(out io.Writer, images []clientset.ImageListItem) {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	defer w.Flush()

	// 打印表头
	fmt.Fprintln(w, "NAME\tTAG\tOS\tARCH\tSIZE(MB)\tCREATED")

	for _, img := range images {
		// 解析并格式化创建时间
		createdTime, err := time.Parse(time.RFC3339Nano, img.CreatedTime)
		var createdStr string
		if err == nil {
			// 使用一个更友好的格式，例如 "2023-11-17"
			createdStr = createdTime.Format("2006-01-02")
		} else {
			createdStr = "N/A" // 如果时间格式解析失败，则优雅地处理
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%.2f\t%s\n",
			img.Name,
			img.Tag,
			img.OS,
			img.Arch,
			img.Size,
			createdStr,
		)
	}
}

// PrintImageDetails 将单个镜像的详细信息以分层、人类可读的格式打印出来。
func PrintImageDetails(out io.Writer, details *clientset.ImageDetails) {
	// --- 打印顶层基础信息 ---
	fmt.Fprintf(out, "Name:         %s\n", details.Name)
	fmt.Fprintf(out, "ID:           %s\n", details.ID)
	fmt.Fprintf(out, "Tag:          %s\n", details.Tag)
	fmt.Fprintf(out, "Path:         %s\n", details.Path)
	fmt.Fprintf(out, "OS/Arch:      %s/%s\n", details.OS, details.Arch)
	fmt.Fprintf(out, "Size:         %.2f MB\n", details.Size)
	fmt.Fprintf(out, "Created:      %s\n", details.CreatedTime)
	if details.Author != nil {
		fmt.Fprintf(out, "Author:       %s\n", *details.Author)
	}
	fmt.Fprintf(out, "OCI Version:  %s\n", details.OCIVersion)
	fmt.Fprintf(out, "Pulled:       %s\n", strconv.FormatBool(details.Pulled))

	// --- 打印 Config 部分 (严格按照 EcsImageConfig 结构) ---
	if details.Config != nil {
		config := details.Config
		fmt.Fprintf(out, "Configuration:\n")

		if config.Hostname != "" {
			fmt.Fprintf(out, "  Hostname:     %s\n", config.Hostname)
		}

		// --- 核心修复 1: 打印 Root 信息 ---
		if config.Root != nil {
			fmt.Fprintf(out, "  Root Filesystem:\n")
			fmt.Fprintf(out, "    Path:       %s\n", config.Root.Path)
			fmt.Fprintf(out, "    Read Only:  %t\n", config.Root.Readonly)
		}

		// 打印 Process 信息
		if config.Process != nil {
			fmt.Fprintf(out, "  Process:\n")
			if len(config.Process.Args) > 0 {
				// 使用 Args 作为 Command
				fmt.Fprintf(out, "    Command:      %s\n", strings.Join(config.Process.Args, " "))
			}
			if config.Process.Cwd != "" {
				fmt.Fprintf(out, "    Working Dir:  %s\n", config.Process.Cwd)
			}
			if len(config.Process.Env) > 0 {
				fmt.Fprintf(out, "    Environment:\n")
				for _, env := range config.Process.Env {
					fmt.Fprintf(out, "      %s\n", env)
				}
			}
		}

		// 打印 Mounts
		if len(config.Mounts) > 0 {
			fmt.Fprintf(out, "  Mounts:\n")
			for _, mount := range config.Mounts {
				fmt.Fprintf(out, "    - Destination: %s\n", mount.Destination)
				fmt.Fprintf(out, "      Source:      %s\n", mount.Source)
				fmt.Fprintf(out, "      Options:     %s\n", strings.Join(mount.Options, ","))
			}
		}

		// 打印 SylixOS 特有配置
		if config.SylixOS != nil {
			s := config.SylixOS
			fmt.Fprintf(out, "  SylixOS:\n")
			// 1. Commands
			if len(s.Commands) > 0 {
				fmt.Fprintf(out, "    Commands: (%s)\n", strings.Join(s.Commands, ", "))
			}
			// 2. Devices
			if len(s.Devices) > 0 {
				fmt.Fprintf(out, "    Devices:\n")
				for _, device := range s.Devices {
					fmt.Fprintf(out, "      %s (access: %s)\n", device.Path, device.Access)
				}
			}
			// 3. Network
			if s.Network != nil {
				fmt.Fprintf(out, "    Network:\n")
				fmt.Fprintf(out, "      FTPD Enabled:    %t\n", s.Network.FtpdEnable)
				fmt.Fprintf(out, "      TelnetD Enabled: %t\n", s.Network.TelnetdEnable)
			}
			// 4. Resources
			if s.Resources != nil {
				fmt.Fprintf(out, "    Resources:\n")
				if s.Resources.Memory != nil {
					fmt.Fprintf(out, "      Memory Limit:  %d MB\n", s.Resources.Memory.MemoryLimitMB)
					fmt.Fprintf(out, "      KHeap Limit:   %d\n", s.Resources.Memory.KheapLimit)
				}
				if s.Resources.Disk != nil {
					fmt.Fprintf(out, "      Disk Limit:    %d MB\n", s.Resources.Disk.LimitMB)
				}
				if s.Resources.CPU != nil {
					fmt.Fprintf(out, "      CPU Priority (High/Low): %d/%d\n", s.Resources.CPU.HighestPrio, s.Resources.CPU.LowestPrio)
				}
				if s.Resources.KernelObject != nil {
					ko := s.Resources.KernelObject
					fmt.Fprintf(out, "      Kernel Objects:\n")
					fmt.Fprintf(out, "        Thread Limit:         %d\n", ko.ThreadLimit)
					fmt.Fprintf(out, "        Thread Pool Limit:    %d\n", ko.ThreadPoolLimit)
					fmt.Fprintf(out, "        Event Limit:          %d\n", ko.EventLimit)
					fmt.Fprintf(out, "        Event Set Limit:      %d\n", ko.EventSetLimit)
					fmt.Fprintf(out, "        Partition Limit:      %d\n", ko.PartitionLimit)
					fmt.Fprintf(out, "        Region Limit:         %d\n", ko.RegionLimit)
					fmt.Fprintf(out, "        Msg Queue Limit:      %d\n", ko.MsgQueueLimit)
					fmt.Fprintf(out, "        Timer Limit:          %d\n", ko.TimerLimit)
					if ko.RMSLimit > 0 {
						fmt.Fprintf(out, "        RMS Limit:            %d\n", ko.RMSLimit)
					}
					if ko.ThreadVarLimit > 0 {
						fmt.Fprintf(out, "        Thread Var Limit:     %d\n", ko.ThreadVarLimit)
					}
					if ko.PosixMqueueLimit > 0 {
						fmt.Fprintf(out, "        Posix Mqueue Limit:   %d\n", ko.PosixMqueueLimit)
					}
					if ko.DlopenLibraryLimit > 0 {
						fmt.Fprintf(out, "        Dlopen Library Limit: %d\n", ko.DlopenLibraryLimit)
					}
					if ko.XSIIPCLimit > 0 {
						fmt.Fprintf(out, "        XSIIPC Limit:         %d\n", ko.XSIIPCLimit)
					}
					if ko.SocketLimit > 0 {
						fmt.Fprintf(out, "        Socket Limit:         %d\n", ko.SocketLimit)
					}
					if ko.SRTPLimit > 0 {
						fmt.Fprintf(out, "        SRTP Limit:           %d\n", ko.SRTPLimit)
					}
					if ko.DeviceLimit > 0 {
						fmt.Fprintf(out, "        Device Limit:         %d\n", ko.DeviceLimit)
					}
				}
			}
		}
	} else {
		fmt.Fprintf(out, "Configuration: Not available\n")
	}
}

// PrintNodeDetails 打印聚合后的节点详细信息。
func PrintNodeDetails(out io.Writer, view *clientset.NodeView, metrics *clientset.NodeMetrics) {
	// --- 打印静态/关系信息 (来自 NodeView) ---
	fmt.Fprintf(out, "Name:         %s\n", view.Name)
	fmt.Fprintf(out, "ID:           %s\n", view.ID)
	fmt.Fprintf(out, "Status:       %s\n", view.Status)
	fmt.Fprintf(out, "Type:         %s\n", view.Type)
	fmt.Fprintf(out, "\n")

	// --- 打印实时指标 (来自 NodeMetrics) ---
	fmt.Fprintf(out, "Metrics (real-time):\n")
	fmt.Fprintf(out, "  Uptime:       %s\n", (time.Duration(metrics.Uptime) * time.Second).String())
	fmt.Fprintf(out, "  CPU Usage:    %s%%\n", metrics.CPU.Percent)

	// Memory Usage with dynamic unit
	ramSizeGB := float64(metrics.RAM.Size) / 1024 / 1024 / 1024
	if ramSizeGB >= 1 {
		fmt.Fprintf(out, "  Memory Usage: %s%% (%.2f GiB)\n", metrics.RAM.Percent, ramSizeGB)
	} else {
		ramSizeMB := float64(metrics.RAM.Size) / 1024 / 1024
		if ramSizeMB >= 1 {
			fmt.Fprintf(out, "  Memory Usage: %s%% (%.2f MiB)\n", metrics.RAM.Percent, ramSizeMB)
		} else {
			ramSizeKB := float64(metrics.RAM.Size) / 1024
			if ramSizeKB >= 1 {
				fmt.Fprintf(out, "  Memory Usage: %s%% (%.2f KiB)\n", metrics.RAM.Percent, ramSizeKB)
			} else {
				fmt.Fprintf(out, "  Memory Usage: %s%% (%d B)\n", metrics.RAM.Percent, int64(metrics.RAM.Size))
			}
		}
	}

	// Disk Usage with dynamic unit (assuming ROM.Size is in MB)
	romSizeMB := metrics.ROM.Size
	if romSizeMB >= 1024 {
		romSizeGB := romSizeMB / 1024
		fmt.Fprintf(out, "  Disk Usage:   %s%% (%.2f GiB)\n", metrics.ROM.Percent, romSizeGB)
	} else {
		fmt.Fprintf(out, "  Disk Usage:   %s%% (%.2f MiB)\n", metrics.ROM.Percent, romSizeMB)
	}
	fmt.Fprintf(out, "  Containers:   %d running / %d stopped\n", metrics.Running, metrics.Stop)
	fmt.Fprintf(out, "  Processes:    %d\n", metrics.ProcessCount)
	fmt.Fprintf(out, "\n")

	// --- 打印容器列表 (来自 NodeView) ---
	if len(view.Children) > 0 {
		fmt.Fprintf(out, "Containers on this node (%d):\n", len(view.Children))
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  NAME\tID\tSTATUS\tSERVICE")
		for _, c := range view.Children {
			serviceName := "N/A"
			if len(c.Children) > 0 {
				serviceName = c.Children[0].Name
			}
			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", c.Name, c.ID, c.Status, serviceName)
		}
		w.Flush()
	} else {
		fmt.Fprintf(out, "No containers found on this node.\n")
	}
}

// PrintServicesTable 将服务列表以表格形式打印到指定的 writer。
func PrintServicesTable(out io.Writer, services []clientset.ProvisionListRow) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// 打印表头
	fmt.Fprintln(w, "NAME\tDEPLOY_STATUS\tPOLICY\tONLINE\tDESIRED\tIMAGE\tID")

	for _, svc := range services {
		// 组合一个易于阅读的镜像名
		imageName := "N/A"
		if len(svc.ImageList) > 0 {
			img := svc.ImageList[0]
			imageName = fmt.Sprintf("%s:%s", img.Name, img.Tag)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\t%s\n",
			svc.Name,
			svc.Status,
			svc.Policy,
			svc.InstanceOnline,
			svc.Factor, // Factor 代表期望的副本数
			imageName,
			svc.ID,
		)
	}
}

// PrintServiceDetails 打印聚合后的服务详细信息。
func PrintServiceDetails(out io.Writer, details *clientset.ServiceGet, containers []clientset.ContainerInfo) {
	// --- 基础信息 ---
	fmt.Fprintf(out, "Name:           %s\n", details.Name)
	fmt.Fprintf(out, "ID:             %s\n", details.ID)
	fmt.Fprintf(out, "Deploy Status:  %s\n", details.Status)
	fmt.Fprintf(out, "Healthy:        %t\n", details.Healthy)
	fmt.Fprintf(out, "Created:        %s\n", details.CreatedTime)
	fmt.Fprintf(out, "Updated:        %s\n", details.UpdatedTime)

	// --- 部署信息 ---
	fmt.Fprintf(out, "Deployment:\n")
	fmt.Fprintf(out, "  Policy:       %s\n", details.Policy)
	fmt.Fprintf(out, "  Replicas:     %d desired, %d online, %d active\n", details.Factor, details.InstanceOnline, details.InstanceActive)

	// --- 镜像信息 (简洁版) ---
	if details.Image != nil {
		fmt.Fprintf(out, "Image:\n")
		fmt.Fprintf(out, "  Reference:    %s\n", details.Image.Ref)
		fmt.Fprintf(out, "  Pull Policy:  %s\n", details.Image.PullPolicy)
		fmt.Fprintf(out, "  Auto Upgrade: %s\n", details.Image.AutoUpgrade)
		fmt.Fprintf(out, "  (Use 'ecsm-cli describe image %s' for full details)\n", details.Image.Ref)
	}
	fmt.Fprintf(out, "\n")

	// --- 节点信息 ---
	if len(details.NodeList) > 0 {
		fmt.Fprintf(out, "Nodes (%d):\n", len(details.NodeList))
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  NAME\tADDRESS\tID")
		for _, node := range details.NodeList {
			fmt.Fprintf(w, "  %s\t%s\t%s\n", node.NodeName, node.Address, node.NodeID)
		}
		w.Flush()
		fmt.Fprintf(out, "\n")
	}

	// --- 容器实例 ---
	if len(containers) > 0 {
		fmt.Fprintf(out, "Containers (%d):\n", len(containers))
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  NAME\tSTATUS\tRESTARTS\tNODE\tID")
		for _, c := range containers {
			fmt.Fprintf(w, "  %s\t%s\t%d\t%s\t%s\n", c.Name, c.Status, c.RestartCount, c.NodeName, c.ID)
		}
		w.Flush()
	} else {
		fmt.Fprintf(out, "No container instances found for this service.\n")
	}
}

// PrintContainersTable 将容器列表以表格形式打印到指定的 writer。
func PrintContainersTable(out io.Writer, containers []clientset.ContainerInfo) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// 打印表头
	fmt.Fprintln(w, "NAME\tSTATUS\tRESTARTS\tIMAGE\tSERVICE\tNODE")

	for _, c := range containers {
		// 组合一个易于阅读的镜像名
		imageRef := fmt.Sprintf("%s:%s", c.ImageName, c.ImageVersion)

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\n",
			c.Name,
			c.Status,
			c.RestartCount,
			imageRef,
			c.ServiceName,
			c.NodeName,
		)
	}
}

// PrintContainerDetails 打印聚合后的容器详细信息。
func PrintContainerDetails(out io.Writer, details *clientset.ContainerInfo, history *clientset.ContainerHistoryList) {
	// --- 基础信息 ---
	fmt.Fprintf(out, "Name:           %s\n", details.Name)
	fmt.Fprintf(out, "ID:             %s\n", details.ID)
	fmt.Fprintf(out, "Task ID:        %s\n", details.TaskID)
	fmt.Fprintf(out, "Status:         %s\n", details.Status)
	fmt.Fprintf(out, "Deploy Status:  %s\n", details.DeployStatus)
	if details.FailedMessage != nil {
		fmt.Fprintf(out, "Failed Message: %s\n", *details.FailedMessage)
	}
	fmt.Fprintf(out, "\n")

	// --- 归属信息 ---
	fmt.Fprintf(out, "Controlled By:  Service/%s\n", details.ServiceName)
	fmt.Fprintf(out, "Node:           %s (%s)\n", details.NodeName, details.Address)
	fmt.Fprintf(out, "Image:          %s:%s\n", details.ImageName, details.ImageVersion)
	fmt.Fprintf(out, "\n")

	// --- 运行时信息 ---
	fmt.Fprintf(out, "Runtime Info:\n")
	uptime := time.Duration(details.Uptime) * time.Second
	fmt.Fprintf(out, "  Started At:   %s\n", details.StartedTime)
	fmt.Fprintf(out, "  Created At:   %s\n", details.CreatedTime)
	fmt.Fprintf(out, "  Uptime:       %s\n", uptime.String())
	fmt.Fprintf(out, "  Restarts:     %d\n", details.RestartCount)
	fmt.Fprintf(out, "\n")

	// --- 资源使用 ---
	fmt.Fprintf(out, "Resource Usage:\n")
	fmt.Fprintf(out, "  CPU:          %.2f%%\n", details.CPUUsage.Total)
	memUsageMiB := float64(details.MemoryUsage) / 1024 / 1024
	memLimitMiB := float64(details.MemoryLimit) / 1024 / 1024
	fmt.Fprintf(out, "  Memory:       %.2f MiB / %.2f MiB\n", memUsageMiB, memLimitMiB)
	diskUsageGiB := float64(details.SizeUsage) / 1024 / 1024 / 1024
	diskLimitGiB := float64(details.SizeLimit) / 1024 / 1024 / 1024
	fmt.Fprintf(out, "  Disk:         %.2f GiB / %.2f GiB\n", diskUsageGiB, diskLimitGiB)
	fmt.Fprintf(out, "\n")

	// --- 操作历史 ---
	if history != nil && len(history.Items) > 0 {
		fmt.Fprintf(out, "History (%d):\n", len(history.Items))
		w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  TIME\tACTION\tUSER")
		for _, h := range history.Items {
			fmt.Fprintf(w, "  %s\t%s\t%s\n", h.Time, h.Cmd, h.User)
		}
		w.Flush()
	} else {
		fmt.Fprintf(out, "No action history found.\n")
	}
}
