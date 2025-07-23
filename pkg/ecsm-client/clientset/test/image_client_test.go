// file: pkg/ecsm-client/clientset/test/image_client_test.go

package test

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/fx147/ecsm-operator/pkg/ecsm-client/clientset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- newTestClientset() 辅助函数 (已存在) ---

// TestImageClient_ReadOperations 对镜像的只读操作进行测试。
// 这个测试是安全的，因为它不会修改任何外部系统状态。
// 它依赖于你的 ECSM 环境中 "local" 仓库里至少存在一个镜像。
func TestImageClient_ReadOperations(t *testing.T) {
	// --- Setup ---
	cs := newTestClientset(t)
	imageClient := cs.Images()
	ctx := context.Background()

	// --- Test: List ---
	t.Run("List", func(t *testing.T) {
		opts := clientset.ImageListOptions{
			RegistryID: "local",
			PageNum:    1,
			PageSize:   10,
		}
		list, err := imageClient.List(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, list)
		// 核心前置条件：你的 local 仓库中必须至少有一个镜像
		require.NotEmpty(t, list.Items, "测试失败：ECSM 'local' 仓库中必须至少存在一个镜像")

		// 抽查第一个镜像的字段
		firstImage := list.Items[0]
		assert.NotEmpty(t, firstImage.ID)
		assert.NotEmpty(t, firstImage.Name)
		assert.NotZero(t, firstImage.Size)
	})

	// --- Test: GetStatistics ---
	t.Run("GetStatistics", func(t *testing.T) {
		stats, err := imageClient.GetStatistics(ctx)
		require.NoError(t, err)
		require.NotNil(t, stats)
		assert.GreaterOrEqual(t, stats.Local, 1, "本地镜像数量至少应为1")
	})

	// --- Test: GetRepositoryInfo ---
	t.Run("GetRepositoryInfo", func(t *testing.T) {
		repoList, err := imageClient.GetRepositoryInfo(ctx, clientset.RepositoryInfoOptions{})
		require.NoError(t, err)
		require.NotEmpty(t, repoList, "镜像仓库列表不应为空")

		// 检查是否至少包含 local 仓库
		foundLocal := false
		for _, repo := range repoList {
			if repo.RegistryID == "local" {
				foundLocal = true
				assert.Equal(t, "本地仓库", repo.RegistryName)
				break
			}
		}
		assert.True(t, foundLocal, "应该在仓库列表中找到 'local' 仓库")
	})

	// --- Test: GetDetails and GetConfig (dependent test) ---
	t.Run("GetDetailsAndConfig", func(t *testing.T) {
		// 1. 先 List 获取一个已知存在的镜像
		list, err := imageClient.List(ctx, clientset.ImageListOptions{RegistryID: "local", PageNum: 1, PageSize: 1})
		require.NoError(t, err)
		require.NotEmpty(t, list.Items, "无法获取任何镜像用于 Get 测试")
		imageToList := list.Items[0]
		imageID := imageToList.ID

		t.Logf("选定的测试镜像: ID=%s, Name=%s, OS=%s", imageID, imageToList.Name, imageToList.OS)

		// 2. 测试 GetDetails
		details, err := imageClient.GetDetails(ctx, "local", imageID)
		require.NoError(t, err)
		require.NotNil(t, details)
		assert.Equal(t, imageID, details.ID)
		assert.Equal(t, imageToList.Name, details.Name)
		require.NotNil(t, details.Config, "镜像详情中的 config 字段不应为 nil")

		// 3. 测试 GetConfig
		// 构造 ref 字符串: name@tag#os
		// 注意: # 需要被 URL 编码
		ref := fmt.Sprintf("%s@%s#%s", imageToList.Name, imageToList.Tag, imageToList.OS)
		// url.PathEscape 会正确处理 '#' -> '%23'，但我们的 rest client 会自动处理，这里只是演示
		t.Logf("构造的 ref 字符串: %s (编码后: %s)", ref, url.PathEscape(ref))

		config, err := imageClient.GetConfig(ctx, ref)
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.NotNil(t, config.SylixOS, "SylixOS 镜像的 config.SylixOS 字段不应为 nil")
		assert.NotNil(t, config.Platform, "SylixOS 镜像的 config.Platform 字段不应为 nil")
	})
}
