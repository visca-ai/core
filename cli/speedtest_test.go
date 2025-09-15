package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/agent/agenttest"
	"github.com/wirtualdev/wirtualdev/v2/cli"
	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/testutil"
	"github.com/wirtualdev/wirtualdev/v2/wirtuald/wirtualdtest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk"
)

func TestSpeedtest(t *testing.T) {
	t.Parallel()
	t.Skip("Flaky test - see https://github.com/wirtualdev/wirtualdev/issues/6321")
	if testing.Short() {
		t.Skip("This test takes a minimum of 5ms per a hardcoded value in Tailscale!")
	}
	client, workspace, agentToken := setupWorkspaceForAgent(t)
	_ = agenttest.New(t, client.URL, agentToken)
	wirtualdtest.AwaitWorkspaceAgents(t, client, workspace.ID)

	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
	defer cancel()

	require.Eventually(t, func() bool {
		ws, err := client.Workspace(ctx, workspace.ID)
		if !assert.NoError(t, err) {
			return false
		}
		a := ws.LatestBuild.Resources[0].Agents[0]
		return a.Status == wirtualsdk.WorkspaceAgentConnected &&
			a.LifecycleState == wirtualsdk.WorkspaceAgentLifecycleReady
	}, testutil.WaitLong, testutil.IntervalFast, "agent is not ready")

	inv, root := clitest.New(t, "speedtest", workspace.Name)
	clitest.SetupConfig(t, client, root)
	pty := ptytest.New(t)
	inv.Stdout = pty.Output()
	inv.Stderr = pty.Output()

	ctx, cancel = context.WithTimeout(context.Background(), testutil.WaitLong)
	defer cancel()

	inv.Logger = testutil.Logger(t).Named("speedtest")
	cmdDone := tGo(t, func() {
		err := inv.WithContext(ctx).Run()
		assert.NoError(t, err)
	})
	<-cmdDone
}

func TestSpeedtestJson(t *testing.T) {
	t.Parallel()
	t.Skip("Potentially flaky test - see https://github.com/wirtualdev/wirtualdev/issues/6321")
	if testing.Short() {
		t.Skip("This test takes a minimum of 5ms per a hardcoded value in Tailscale!")
	}
	client, workspace, agentToken := setupWorkspaceForAgent(t)
	_ = agenttest.New(t, client.URL, agentToken)
	wirtualdtest.AwaitWorkspaceAgents(t, client, workspace.ID)

	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
	defer cancel()

	require.Eventually(t, func() bool {
		ws, err := client.Workspace(ctx, workspace.ID)
		if !assert.NoError(t, err) {
			return false
		}
		a := ws.LatestBuild.Resources[0].Agents[0]
		return a.Status == wirtualsdk.WorkspaceAgentConnected &&
			a.LifecycleState == wirtualsdk.WorkspaceAgentLifecycleReady
	}, testutil.WaitLong, testutil.IntervalFast, "agent is not ready")

	inv, root := clitest.New(t, "speedtest", "--output=json", workspace.Name)
	clitest.SetupConfig(t, client, root)
	out := bytes.NewBuffer(nil)
	inv.Stdout = out
	ctx, cancel = context.WithTimeout(context.Background(), testutil.WaitLong)
	defer cancel()

	inv.Logger = testutil.Logger(t).Named("speedtest")
	cmdDone := tGo(t, func() {
		err := inv.WithContext(ctx).Run()
		assert.NoError(t, err)
	})
	<-cmdDone

	var result cli.SpeedtestResult
	require.NoError(t, json.Unmarshal(out.Bytes(), &result))
	require.Len(t, result.Intervals, 5)
}
