// file: pkg/ecsm-client/clientset/test/node_client_test.go

package test

import (
	"context"
	"testing"

	"github.com/fx147/ecsm-operator/pkg/ecsm-client/clientset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- newTestClientset() 辅助函数 (已存在) ---

// TestNodeClient_ReadOperations 对节点的只读操作进行测试。
// 这个测试是安全的，因为它不会修改任何外部系统状态。
// 它依赖于你的 ECSM 环境中至少存在一个已注册的节点。
func TestNodeClient_ReadOperations(t *testing.T) {
	// --- Setup ---
	cs := newTestClientset(t)
	nodeClient := cs.Nodes()
	ctx := context.Background()

	// --- Test: List ---
	t.Run("List", func(t *testing.T) {
		opts := clientset.NodeListOptions{
			PageNum:  1,
			PageSize: 10,
		}
		list, err := nodeClient.List(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, list)
		// 核心前置条件：你的环境中必须至少有一个节点
		require.GreaterOrEqual(t, len(list.Items), 1, "测试失败：ECSM环境中必须至少存在一个节点")

		// 随机抽查第一个节点的字段是否符合预期
		firstNode := list.Items[0]
		assert.NotEmpty(t, firstNode.ID)
		assert.NotEmpty(t, firstNode.Name)
		assert.NotEmpty(t, firstNode.Status)
	})

	// --- Test: GetByID & GetByName ---
	t.Run("GetByNameAndByID", func(t *testing.T) {
		// 1. 先 List 获取一个已知存在的节点
		list, err := nodeClient.List(ctx, clientset.NodeListOptions{PageNum: 1, PageSize: 1})
		require.NoError(t, err)
		require.NotEmpty(t, list.Items, "无法获取任何节点用于Get测试")
		existingNode := list.Items[0]

		// 2. 测试 GetByName
		nodeByName, err := nodeClient.GetByName(ctx, existingNode.Name)
		require.NoError(t, err, "通过名称获取已知存在的节点不应失败")
		require.NotNil(t, nodeByName)
		assert.Equal(t, existingNode.ID, nodeByName.ID)
		assert.Equal(t, existingNode.Name, nodeByName.Name)

		// 3. 测试 GetByID
		nodeByID, err := nodeClient.GetByID(ctx, existingNode.ID)
		require.NoError(t, err, "通过ID获取已知存在的节点不应失败")
		require.NotNil(t, nodeByID)
		assert.Equal(t, existingNode.ID, nodeByID.ID)
		assert.Equal(t, existingNode.Name, nodeByID.Name)
		// 我们可以检查 Get 到的详情里，密码字段不为空（假设API会返回）
		assert.NotEmpty(t, nodeByID.Password, "GetByID返回的详情中，Password字段不应为空")
	})

	// --- Test: ListStatus ---
	t.Run("ListStatus", func(t *testing.T) {
		// 1. 获取一个已知节点的ID
		list, err := nodeClient.List(ctx, clientset.NodeListOptions{PageNum: 1, PageSize: 1})
		require.NoError(t, err)
		require.NotEmpty(t, list.Items)
		existingNodeID := list.Items[0].ID

		// 2. 获取该节点的状态
		statusList, err := nodeClient.ListStatus(ctx, []string{existingNodeID})
		require.NoError(t, err)
		require.Len(t, statusList, 1, "查询一个节点的状态应该返回一个结果")
		status := statusList[0]
		assert.Equal(t, existingNodeID, status.ID)
		assert.NotZero(t, status.MemoryTotal, "节点总内存不应为0")
	})

	// --- Test: Validation ---
	t.Run("Validation", func(t *testing.T) {
		// (这部分测试保持不变，因为它也是只读的)
		list, err := nodeClient.List(ctx, clientset.NodeListOptions{PageNum: 1, PageSize: 1})
		require.NoError(t, err)
		require.NotEmpty(t, list.Items)
		existingNode := list.Items[0]

		// 校验已存在的名字
		validateNameOpts := clientset.NodeValidateNameOptions{Name: existingNode.Name}
		nameResult, err := nodeClient.ValidateName(ctx, validateNameOpts)
		require.NoError(t, err)
		assert.False(t, nameResult.IsValid, "校验已存在的名称应该返回 IsValid=false")
	})
}

// TestNodeClient_Lifecycle 测试节点的写操作生命周期。
// 这个测试会修改外部系统，并且依赖一个可用的、未被注册的IP地址。
// 在常规单元/集成测试中可以跳过。
func TestNodeClient_Lifecycle(t *testing.T) {
	// 使用 t.Skip() 可以让测试框架记录这个测试，但会跳过它的执行。
	t.Skip("跳过生命周期测试，因为它需要一个真实、可用的、未注册的节点IP，并且会修改外部系统。")

	// ... (之前的所有生命周期测试代码可以保留在这里) ...
}
