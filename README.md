# ecsm-operator
为翼辉公司容器管理平台ECSM设计的调谐循环控制器

## 阶段一

目标： 实现最基础的调谐循环，证明核心逻辑可行。

实现内容：

1. **定义 API 对象**：用 Go struct 定义你的核心资源，比如 ECSMApp，包含 Spec（期望状态）和 Status（实际状态）字段。
2. **编写 ECSM 客户端**：封装对 ECSM HTTP API 的调用（ListApps, GetApp, CreateContainer, DeleteContainer 等）。
3. **实现 reconcile 函数**：业务逻辑核心，比较 desired 和 actual 的差异，然后调用客户端执行操作。
4. **构建简单的轮询控制器**：
   - 一个 main 函数或 Run 函数。
   - 使用 time.Ticker 设置一个固定的轮询周期（例如 30 秒）。
   - 在循环中：
     a. 从某个地方读取期望状态（为了简单，可以先硬编码在代码里，或从一个本地 YAML 文件读取）。
     b. 调用 ECSM 客户端 List 全量实际状态。
     c. 调用 reconcile 函数。
     d. 处理错误并打印详细日志。
### 阶段性任务
1. **pkg/apis/... (API 定义包)**
    
    - **职责**: 存放我们已经定义好的所有 API 结构体，如 ECSMService, ObjectMeta 等。
        
    - **状态**: **已完成 95%**。这是我们的地基。
        
2. **pkg/registry (ECSM Registry 实现包)**
    
    - **职责**: 实现对我们“声明式世界”的存储和检索。它需要提供一个接口，让其他模块可以 Get, List, Create, Update, Delete 我们的 ECSMService 对象。
        
    - **初期实现**: 我们将从一个简单的**基于文件的存储 (FileStore)** 开始。它会把每个 ECSMService 对象作为一个单独的 JSON 或 YAML 文件存储在磁盘上。
        
3. **pkg/ecsm_client (ECSM API 客户端包)**
    
    - **职责**: 封装所有与“现实世界”（ECSM 平台）的 HTTP REST API 交互。它将提供类型安全的方法，如 ListECSMContainers(labels map[string]string) ([]Container, error) 和 CreateECSMContainer(payload *CreateContainerRequest) error。
        
    - **作用**: 将控制器与底层的 http.Client 和 JSON 序列化/反序列化逻辑解耦。控制器只需要调用这个客户端的方法，而不需要关心具体的 HTTP 细节。
        
4. **pkg/controller (核心控制器逻辑包)**
    
    - **职责**: 这是项目的“大脑”。它包含核心的**调谐循环 (Reconciliation Loop)**。
        
    - **逻辑**:  
        a. 从 ECSM Registry 获取一个 ECSMService 对象。  
        b. 使用 ecsm_client 获取 ECSM 平台上的现实状态。  
        c. 比较“期望”与“现实”的差异。  
        d. 使用 ecsm_client 来执行必要的创建/删除操作，以弥合差异。  
        e. 更新 ECSMService 对象的 status 字段，并将其写回 ECSM Registry。
        
5. **cmd/ecsm-operator/main.go (主程序入口)**
    
    - **职责**: 这是 Operator 的启动器。
        
    - **任务**: 初始化所有模块（创建 registry 实例，创建 ecsm_client 实例，创建 controller 实例），然后启动一个无限循环，定期触发控制器的调谐逻辑。
        
6. **cmd/ecsmctl/main.go (命令行工具)**
    
    - **职责**: 为用户提供一个与 ECSM Registry 交互的工具。这是用户将他们的 YAML “意图”送入我们系统的唯一方式。
        
    - **核心功能**: 实现一个 ecsmctl apply -f <filename.yaml> 命令。该命令会读取 YAML 文件，将其解析为 ECSMService 结构体，然后调用 registry 包的功能将其保存到存储中。

### 开发顺序

**第一步（也是下一步）：实现 ECSM Registry (文件存储版)**

- **为什么是它？**:
    
    - **它是基础**: 控制器 (ecsm-operator) 和命令行 (ecsmctl) 都依赖它。没有它，我们寸步难行。
        
    - **它最简单且独立**: 我们可以不依赖任何其他模块，快速地实现一个 FileStore，它只需要能把 ECSMService 对象序列化成 JSON 并写入文件即可。
        
- **产出物**: 一个 pkg/registry/filestore.go 文件，提供 Get, Save 等方法。
    

**第二步：实现 ecsmctl 的 apply 命令**

- **为什么是它？**:
    
    - **完成输入闭环**: 一旦 Registry 完成，我们就可以立刻构建 ecsmctl。这让我们拥有了将用户的 YAML 文件存入我们系统的能力。这个流程一打通，我们就可以为后续的控制器准备好“测试数据”了。
        
- **产出物**: 一个可以运行的 ecsmctl 二进制文件，能成功将 ECSMService 的 YAML 存为磁盘上的文件。
    

**第三步：定义 ecsm_client 的接口和 Mock 实现**

- **为什么是它？**:
    
    - **解耦开发**: 控制器的逻辑依赖于 ecsm_client。但我们不希望在开发控制器时，因为网络问题或 ECSM 环境问题而受阻。
        
    - **接口先行**: 我们先定义一个 ECSMClient 的 **Go 接口**，比如 type Client interface { ListContainers(...) ... }。
        
    - **创建 Mock**: 然后我们创建一个**假的、用于测试的 Mock 实现** (MockClient)，它不发送任何 HTTP 请求，只是返回一些硬编码的假数据。
        
- **产出物**: pkg/ecsm_client/client.go (定义接口) 和 pkg/ecsm_client/mock_client.go (用于测试的假客户端)。

- **子任务 3a: 定义 Client 接口**
    
    - **为什么**: 这是软件工程的最佳实践。我们先在 pkg/ecsm_client/client.go 中定义一个 Client **接口**，清晰地列出我们需要 ECSM 平台提供哪些能力。
        
    - **示例**:
        ```go
        package ecsm_client
        
        type Client interface {
            // ListServices 根据标签列出 ECSM 上的服务
            ListServices(labels map[string]string) ([]ECSMServiceInfo, error)
            // CreateService 在 ECSM 平台上创建一个新服务
            CreateService(payload *CreateServiceRequest) (*ECSMServiceInfo, error)
            // DeleteService 删除一个服务
            DeleteService(serviceID string) error
            // ... 其他需要的方法
        }
        ```
        
    - **好处**: 接口是**契约**。它让我们的控制器可以依赖于这个稳定的契约，而不是某个具体的实现。这为我们未来的测试工作打下了坚实的基础。
        
- **子任务 3b: 实现 httpClient 结构体，发起真实 HTTP 请求**
    
    - **为什么**: 这就是实现你“价值展示”目标的关键一步。
        
    - **任务**: 创建一个 httpClient 结构体，它**实现**我们上面定义的 Client 接口。这个结构体将包含 ECSM 服务器的地址、http.Client 实例，以及所有方法的真实实现——拼接 URL、构造请求体、发送 HTTP 请求、处理响应和错误。
        
    - **目标**: 能够通过调用 myClient.CreateService(...)，真正在 ECSM 平台上创建一个容器/服务。

**第四步：编写控制器最核心的调谐骨架**

- **理由**: 现在，我们拥有了创建“意图”的工具 (ecsmctl)、存放意图的仓库 (Registry)，以及与“现实”交互的桥梁 (ecsm_client)。是时候构建我们的大脑了。
    
- **任务**:
    
    1. 在 main.go 中，初始化真实的 FileStore 和真实的 httpClient。
        
    2. 将这两个实例**注入**到我们的控制器 Reconciler 中。
        
    3. 编写 Reconcile 函数的骨架：
        
        - 从 Registry 获取一个 ECSMService。
            
        - 调用**真实的 httpClient** 的 ListServices 方法，获取平台上的现有服务。
            
        - 进行比较。如果发现期望的 ECSMService 在平台不存在，就调用**真实的 httpClient** 的 CreateService 方法。
            
- **目标**: **实现一个完整的、端到端的场景**：运行 ecsmctl apply -f service.yaml，然后启动 ecsm-operator，能亲眼看到 ECSM 平台上真的多出了一个服务！
    
第三步提前进行集成：
1. **快速验证核心价值**: 完成第四步后，你就拥有了一个可以向任何人演示的核心功能闭环。这极大地增强了项目信心。
    
2. **提前暴露集成问题**: 与 ECSM API 的集成是最可能出现意外的地方（文档不符、认证复杂、响应奇怪等）。我们把这个最大的风险点提到了最前面，一旦解决，后续开发会非常顺利。
    
3. **兼顾良好设计**: 我们通过坚持“接口优先”的原则（3a），保留了未来轻松编写单元测试和 Mock 的能力。当我们想要为控制器编写快速、可靠的测试时，我们只需创建一个 MockClient 来实现同一个 Client 接口，而不需要修改任何一行控制器代码。