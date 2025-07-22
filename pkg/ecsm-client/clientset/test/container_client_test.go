package test

import (
	"context"
	"testing"
	"time"

	"github.com/fx147/ecsm-operator/pkg/ecsm-client/clientset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContainerClient_Get(t *testing.T) {
	clientsetInstance := newTestClientset(t)
	containerClient := clientsetInstance.Containers()
	serviceClient := clientsetInstance.Services()
	ctx := context.Background()

	// 1. 先列出所有服务，找到一个有容器的服务
	listServiceOpts := clientset.ListServicesOptions{
		PageNum:  1,
		PageSize: 100,
	}
	serviceList, err := serviceClient.List(ctx, listServiceOpts)
	require.NoError(t, err)
	require.NotEmpty(t, serviceList.Items, "无法找到任何服务")

	// 在所有服务中找到第一个有容器的服务
	var targetServiceID string
	for _, s := range serviceList.Items {
		if s.InstanceOnline > 0 { // 或者 len(s.ContainerStatusGroup) > 0
			targetServiceID = s.ID
			break
		}
	}
	require.NotEmpty(t, targetServiceID, "在所有服务中都找不到在线的容器")

	// 2. 使用该服务的 ID 列出其下的容器
	listContainerOpts := clientset.ListContainersByServiceOptions{
		PageNum:    1,
		PageSize:   10,
		ServiceIDs: []string{targetServiceID},
	}
	containerList, err := containerClient.ListByService(ctx, listContainerOpts)
	require.NoError(t, err)
	require.NotNil(t, containerList)
	require.NotEmpty(t, containerList.Items, "在服务 (ID: %s) 下找不到任何容器", targetServiceID)

	// 3. --- 核心修复 ---
	//    我们现在获取的是 TaskID，而不是 ID。
	containerToGet := containerList.Items[0]
	taskID := containerToGet.TaskID
	require.NotEmpty(t, taskID, "获取到的容器 TaskID 为空")

	// 使用 TaskID 调用 Get 方法
	containerInfo, err := containerClient.GetByTaskID(ctx, taskID)
	require.NoError(t, err) // 使用 require，出错后会立即停止测试
	require.NotNil(t, containerInfo)

	// 验证返回的详情是否与我们期望的一致
	// 注意：现在 Get 的 ID 参数是 taskID，但返回的对象的 ID 字段是 containerID
	assert.Equal(t, containerToGet.ID, containerInfo.ID, "返回的容器详情ID与列表中的ID不匹配")
	assert.Equal(t, taskID, containerInfo.TaskID, "返回的容器详情TaskID与请求的TaskID不匹配")
}

// TestContainerClient_ListByService 测试根据服务ID列出容器
func TestContainerClient_ListByService(t *testing.T) {
	// --- Setup ---
	clientsetInstance := newTestClientset(t)
	containerClient := clientsetInstance.Containers()
	serviceClient := clientsetInstance.Services()
	ctx := context.Background()

	// 1. 找到 "acc_server" 服务的 ID
	listServiceOpts := clientset.ListServicesOptions{
		PageNum:  1,
		PageSize: 100,
		Name:     "acc_server", // 直接按名称过滤，更精确
	}
	serviceList, err := serviceClient.List(ctx, listServiceOpts)
	require.NoError(t, err)
	require.Len(t, serviceList.Items, 1, "应该只找到一个名为 'acc_server' 的服务")
	accServerID := serviceList.Items[0].ID

	// --- Test ---
	t.Run("ValidServiceID", func(t *testing.T) {
		// 2. 使用该服务的 ID 列出其下的容器
		listContainerOpts := clientset.ListContainersByServiceOptions{
			PageNum:    1,
			PageSize:   10,
			ServiceIDs: []string{accServerID},
		}
		containerList, err := containerClient.ListByService(ctx, listContainerOpts)

		// --- Assertions ---
		require.NoError(t, err)
		require.NotNil(t, containerList)
		assert.NotEmpty(t, containerList.Items, "服务 'acc_server' (ID: %s) 下应该有容器", accServerID)

		// 检查返回的每个容器是否都属于 acc_server
		for _, container := range containerList.Items {
			assert.Equal(t, accServerID, container.ServiceID)
			assert.Equal(t, "acc_server", container.ServiceName)
		}
	})

	t.Run("InvalidServiceID", func(t *testing.T) {
		// 3. 使用一个无效的 Service ID 查询
		listContainerOpts := clientset.ListContainersByServiceOptions{
			PageNum:    1,
			PageSize:   10,
			ServiceIDs: []string{"invalid-service-id"},
		}
		containerList, err := containerClient.ListByService(ctx, listContainerOpts)

		// --- Assertions ---
		require.NoError(t, err)
		require.NotNil(t, containerList)
		// 期望返回一个空列表
		assert.Empty(t, containerList.Items, "使用无效 Service ID 查询时，容器列表应该为空")
		assert.Equal(t, 0, containerList.Total)
	})
}

// TestContainerClient_SubmitControlActionAndGetHistory 测试控制容器状态和获取历史
func TestContainerClient_SubmitControlActionAndGetHistory(t *testing.T) {
	// --- Setup ---
	clientsetInstance := newTestClientset(t)
	containerClient := clientsetInstance.Containers()
	serviceClient := clientsetInstance.Services()
	ctx := context.Background()

	// 1. 找到 acc_server 服务下的一个容器
	// (这部分代码与 TestContainerClient_GetByTaskID 的 setup 类似)
	listServiceOpts := clientset.ListServicesOptions{PageNum: 1, PageSize: 100, Name: "acc_server"}
	serviceList, err := serviceClient.List(ctx, listServiceOpts)
	require.NoError(t, err)
	require.Len(t, serviceList.Items, 1)
	accServerID := serviceList.Items[0].ID

	listContainerOpts := clientset.ListContainersByServiceOptions{PageNum: 1, PageSize: 10, ServiceIDs: []string{accServerID}}
	containerList, err := containerClient.ListByService(ctx, listContainerOpts)
	require.NoError(t, err)
	require.NotEmpty(t, containerList.Items)

	targetContainer := containerList.Items[0]
	containerName := targetContainer.Name
	taskID := targetContainer.TaskID
	require.NotEmpty(t, containerName)
	require.NotEmpty(t, taskID)

	t.Logf("选定的测试容器: Name=%s, TaskID=%s", containerName, taskID)

	// --- Test Control Action ---
	t.Run("SubmitStopAction", func(t *testing.T) {
		// 2. 发送一个 "stop" 指令
		transaction, err := containerClient.SubmitControlActionByName(ctx, containerName, clientset.ActionStop)

		// --- Assertions ---
		require.NoError(t, err)
		require.NotNil(t, transaction)
		assert.NotEmpty(t, transaction.ID, "返回的 Transaction ID 不能为空")
		t.Logf("成功提交 'stop' 动作, Transaction ID: %s", transaction.ID)

		// 由于这是异步操作，我们等待一小段时间，让状态有机会改变
		// 注意：在真实的集成测试中，这里会使用轮询(polling)来检查事务状态，而不是硬编码等待
		time.Sleep(2 * time.Second)
	})

	// --- Test Get History ---
	t.Run("GetActionHistory", func(t *testing.T) {
		// 3. 获取该容器的操作历史
		historyOpts := clientset.ContainerHistoryOptions{
			PageNum:  1,
			PageSize: 10,
			TaskID:   taskID,
		}
		historyList, err := containerClient.GetHistory(ctx, historyOpts)

		// --- Assertions ---
		require.NoError(t, err)
		require.NotNil(t, historyList)
		assert.NotEmpty(t, historyList.Items, "容器的操作历史不应为空")

		// 检查历史记录中是否包含了我们刚刚执行的 "stop" 操作
		foundStopAction := false
		for _, history := range historyList.Items {
			assert.Equal(t, taskID, history.ID, "历史记录中的ID应该与TaskID匹配")
			if history.Cmd == string(clientset.ActionStop) {
				foundStopAction = true
				t.Logf("在历史记录中找到了 'stop' 操作，时间: %s", history.Time)
			}
		}
		assert.True(t, foundStopAction, "应该在操作历史中找到我们刚刚提交的 'stop' 动作")
	})

	// --- Cleanup (可选但推荐): 重新启动容器 ---
	t.Run("SubmitStartActionCleanup", func(t *testing.T) {
		t.Log("正在重新启动容器以清理测试状态...")
		_, err := containerClient.SubmitControlActionByName(ctx, containerName, clientset.ActionStart)
		assert.NoError(t, err, "清理步骤：重新启动容器失败")
		time.Sleep(2 * time.Second)
	})
}
