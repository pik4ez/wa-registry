// +build kubeall kubernetes

package test

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"

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

	deploymentResourcePath, err := filepath.Abs("../k8s/registry-deployment.yaml")
	require.NoError(t, err)
	serviceResourcePath, err := filepath.Abs("../k8s/registry-service.yaml")
	require.NoError(t, err)
	ingressResourcePath, err := filepath.Abs("../k8s/registry-ingress.yaml")
	require.NoError(t, err)

	namespaceName := fmt.Sprintf("wolt-assignment-%s", strings.ToLower(random.UniqueId()))
	options := k8s.NewKubectlOptions("", "", namespaceName)
	k8s.CreateNamespace(t, options, namespaceName)
	defer k8s.DeleteNamespace(t, options, namespaceName)

	defer k8s.KubectlDelete(t, options, deploymentResourcePath)
	defer k8s.KubectlDelete(t, options, serviceResourcePath)
	defer k8s.KubectlDelete(t, options, ingressResourcePath)

	k8s.KubectlApply(t, options, deploymentResourcePath)
	k8s.KubectlApply(t, options, serviceResourcePath)
	k8s.KubectlApply(t, options, ingressResourcePath)

	service := k8s.GetService(t, options, "registry-service")
	require.Equal(t, service.Name, "registry-service")

	k8s.WaitUntilServiceAvailable(t, options, "registry-service", 10, 1*time.Second)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)

	ctx := context.Background()
	pullResp, err := cli.ImagePull(ctx, "hello-world", types.ImagePullOptions{})
	require.NoError(t, err)
	assertNoDockerError(t, pullResp)

	err = cli.ImageTag(ctx, "hello-world", "localhost:5000/hello-world:0.1")
	require.NoError(t, err)

	var authConfig = types.AuthConfig{
		ServerAddress: "http://localhost:5000/v2/",
	}
	authConfigBytes, _ := json.Marshal(authConfig)
	authConfigEncoded := base64.URLEncoding.EncodeToString(authConfigBytes)
	opts := types.ImagePushOptions{RegistryAuth: authConfigEncoded}
	pushResp, err := cli.ImagePush(ctx, "localhost:5000/hello-world:0.1", opts)
	require.NoError(t, err)
	assertNoDockerError(t, pushResp)
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
