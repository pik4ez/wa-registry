// +build kubeall kubernetes

package test

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
)

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

func TestRegistry(t *testing.T) {
	t.Parallel()

	kubectlOptions, clean := createNamespace(t, "wolt-assignment")
	defer clean()

	kubernetesResources := []string{
		"k8s/persistent-volume.yaml",

		"k8s/secret-auth.yaml",

		"../k8s/redis-deployment.yaml",
		"../k8s/redis-service.yaml",

		"../k8s/registry-deployment.yaml",
		"../k8s/registry-service.yaml",
		"../k8s/registry-ingress.yaml",
		"../k8s/registry-pvc.yaml",
	}
	for _, res := range kubernetesResources {
		applyKubernetesResource(t, kubectlOptions, res)
	}

	storageDir := "/tmp/mock-repository-storage/docker"
	assert.NoDirExists(t, storageDir)
	defer os.RemoveAll(storageDir)

	service := k8s.GetService(t, kubectlOptions, "registry-service")
	require.Equal(t, service.Name, "registry-service")

	k8s.WaitUntilServiceAvailable(t, kubectlOptions, "registry-service", 10, 10*time.Second)

	validateFunc := func(code int, body string) bool {
		return code == 200
	}
	http_helper.HttpGetWithRetryWithCustomValidation(t, "http://localhost:5000/metrics", nil, 10, 20*time.Second, validateFunc)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	ctx := context.Background()
	pullResp, err := cli.ImagePull(ctx, "hello-world", types.ImagePullOptions{})
	require.NoError(t, err)
	assertNoDockerError(t, pullResp)

	err = cli.ImageTag(ctx, "hello-world", "localhost:5000/hello-world:0.1")
	require.NoError(t, err)

	var authConfig = types.AuthConfig{
		Username:      "testuser",
		Password:      "testpassword",
		ServerAddress: "http://localhost:5000/v2/",
	}
	authConfigBytes, _ := json.Marshal(authConfig)
	authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)
	opts := types.ImagePushOptions{RegistryAuth: authConfigEncoded}
	pushResp, err := cli.ImagePush(ctx, "localhost:5000/hello-world:0.1", opts)
	require.NoError(t, err)
	assertNoDockerError(t, pushResp)

	assert.DirExists(t, storageDir)
}

func createNamespace(t *testing.T, prefix string) (*k8s.KubectlOptions, func()) {
	namespaceName := fmt.Sprintf("%s-%s", prefix, strings.ToLower(random.UniqueId()))
	options := k8s.NewKubectlOptions("", "", namespaceName)
	k8s.CreateNamespace(t, options, namespaceName)
	clean := func() {
		k8s.DeleteNamespace(t, options, namespaceName)
	}
	return options, clean
}

func applyKubernetesResource(t *testing.T, options *k8s.KubectlOptions, relPath string) {
	absPath, err := filepath.Abs(relPath)
	require.NoError(t, err)
	k8s.KubectlApply(t, options, absPath)
}

func assertNoDockerError(t *testing.T, rd io.Reader) {
	var lastLine string

	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	errLine := &ErrorLine{}
	json.Unmarshal([]byte(lastLine), errLine)
	require.Empty(t, errLine.Error)
	require.NoError(t, scanner.Err())
}
