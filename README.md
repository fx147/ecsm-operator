# ecsm-operator

**一个为资源受限环境设计的、不依赖 Kubernetes 的轻量化云原生编排运行时框架。**

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/fx147/ecsm-operator)
[![Go Version](https://img.shields.io/badge/go-1.18+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

`ecsm-operator` 是一个探索性的后端项目，旨在将 Kubernetes 的核心设计哲学（声明式 API、控制器模式）应用到传统、资源受限的嵌入式环境中。本项目为翼辉公司 SylixOS 嵌入式容器平台 (ECSM) 实现了一套轻量化的控制器运行时框架，为上层高级编排能力的开发构建了坚实的底层基础设施。

## 项目立意与核心挑战

在航天器等极端环境中，计算资源（CPU、内存）极其宝贵，且运维操作依赖高延迟、不稳定的星地链路。传统的命令式运维（如通过 SSH 执行脚本）不仅效率低下，且在面对网络中断或瞬时故障时极其脆弱，缺乏自我修复能力。

本项目的核心挑战在于：

1.  **环境约束**: 如何在不引入 Kubernetes 等重量级依赖的前提下，实现其强大的自动化和编排能力？
2.  **平台局限**: 如何为一个只提供底层 CRUD API 的传统平台 (ECSM)，赋予管理复杂应用所需的**编排**、**状态自愈**和**声明式管理**的能力？

## 核心架构：K8s 思想的轻量化重塑

为应对挑战，本项目对 Kubernetes 的核心架构进行了“外科手术式”的解构与重塑，在继承其设计精髓的同时，用更轻量的自研组件替换了其沉重的依赖：

| Kubernetes 组件 | 我们的轻量化替代方案 | 核心权衡与决策 |
| :--- | :--- | :--- |
| **`etcd` (分布式存储)** | **`bbolt` (嵌入式 KV 存储)** | **放弃分布式**能力，换取**极低的资源占用**和**简化的部署**。 |
| **`API Server` (网络服务)** | **`Registry` (Go 语言库)** | **放弃网络访问**，控制器通过**本地函数调用**与存储交互，实现**零网络开销**。 |
| **`Informer` (基于 WATCH)** | **自研 `Informer` (基于“发布/订阅”+轮询)** | 针对无原生 `WATCH` 的后端，通过**应用层事件通知**实现低延迟，通过**周期性全量同步**保证最终一致性。 |
| **`Cache`/`Indexer` (全量缓存)** | **“版本向量缓存” (`sync.Map`)** | **放弃全量对象缓存**，只缓存 `resourceVersion`，以**可接受的 I/O 开销**换取**数量级的内存节省**。 |

## 技术实现深度解析

#### 1. 可插拔的声明式 API 存储层 (`Registry`)
基于对 K8s 存储模式的分析，项目抽象出通用的 `Store` 接口以解耦业务逻辑与持久化。在对 `PebbleDB` 与 `bbolt` 进行审慎的技术选型后，最终基于 `bbolt` 的**原子事务 (`db.Update`)**，在业务逻辑层实现了**乐观并发控制**，通过 `resourceVersion` 检查原子性地解决了并发写入冲突。

#### 2. 混合式事件驱动的 `Informer`
为解决 `bbolt` **无原生 `WATCH` 机制**的挑战，设计了一套混合式的事件处理与同步模型。
*   **实时路径**: `Informer` 订阅 `Registry` 的“发布/订阅”事件，通过比对 `resourceVersion` 与内存中的**“版本向量缓存” (`sync.Map`)**，低延迟地处理主动变更，避免了重复事件。
*   **周期同步路径**: 定期全量 `List` 所有对象，通过将列表与版本缓存进行高效比对来计算出精确的增量变更 (`Added/Updated/Deleted`)，确保了系统的**最终一致性**并能修复任何丢失的事件。

#### 3. 标准的控制器工作流
实现了**“无全量缓存”**的调谐循环：控制器从 `WorkQueue` (`client-go` 库复用) 获取变更 `key` 后，总是直接从 `Registry` 读取最新的“期望状态 (`spec`)”，并从 `EcsmClient` 读取“现实状态”，确保了决策的实时准确性。

#### 4. 分层的外部 API 客户端库 (`EcsmClient`)
遵循 `client-go` 的分层思想，设计并实现了 `rest` (底层、链式调用) 和 `clientset` (上层、类型安全) 两层客户端，为所有控制器提供了稳定、一致的 ECSM 平台外部 API 访问入口。

## 项目组件

本项目包含两个主要的可执行文件：

*   **`ecsm-operator`**: 控制器管理器。一个长时间运行的后台服务，负责执行所有的调谐循环。
*   **`ecsm-cli`**: 一个面向管理员和开发者的命令行工具，用于直接与 ECSM 平台的 API 进行命令式交互（查询、调试等）。

## 目录结构

```
.
├── cmd/
│   ├── ecsm-cli/      # 命令行工具源码
│   └── ecsm-operator/ # 控制器源码
├── internal/          # 内部共享包 (e.g., ecsm-cli 的 printer)
├── pkg/
│   ├── apis/          # 所有声明式 API 对象的定义 (e.g., ECSMService)
│   ├── controller/    # 控制器的核心业务逻辑
│   ├── ecsm-client/   # 与 ECSM API 交互的客户端库
│   ├── informer/      # 轻量化的 Informer 实现
│   └── registry/      # 声明式 API 的存储层 (业务逻辑 + bbolt 实现)
└── ...
```

## 快速开始

**构建:**
```bash
# 构建所有组件
make build

# 或者单独构建
go build -o ./bin/ecsm-cli ./cmd/ecsm-cli
go build -o ./bin/ecsm-operator ./cmd/ecsm-operator
```

**运行测试:**
```bash
make test
```

## 使用示例 (`ecsm-cli`)

```bash
# 设置 ECSM 服务器地址 (也可以通过配置文件或环境变量)
export ECSMCLI_HOST=192.168.1.100

# 获取所有节点列表
./bin/ecsm-cli get nodes

# 获取 "default" 命名空间下的所有服务
./bin/ecsm-cli get services -n default

# 查看名为 "worker-1" 的节点的详细信息
./bin/ecsm-cli describe node worker-1
```

## 未来路线图

- [ ] **完善 `ecsm-cli`**: 实现 `create`, `delete`, `update` 等写操作命令。
- [ ] **完成 `ECSMServiceController`**: 完整实现 `Reconcile` 循环中的 `Create/Delete` 容器逻辑和滚动更新策略。
- [ ] **实现 `ECSMHpaController`**: 基于 `ECSMService.status` 和 `EcsmClient` 的监控指标，实现服务的自动水平伸缩。
- [ ] **探索更高阶的编排**: 实现基于 `dependsOn` 的服务依赖管理、基于优先级的资源抢占等高级调度功能。

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.
