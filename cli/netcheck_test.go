package cli_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wirtualdev/wirtualdev/v2/cli/clitest"
	"github.com/wirtualdev/wirtualdev/v2/pty/ptytest"
	"github.com/wirtualdev/wirtualdev/v2/wirtualsdk/healthsdk"
)

func TestNetcheck(t *testing.T) {
	t.Parallel()

	pty := ptytest.New(t)
	config := login(t, pty)

	var out bytes.Buffer
	inv, _ := clitest.New(t, "netcheck", "--global-config", string(config))
	inv.Stdout = &out

	clitest.StartWithWaiter(t, inv).RequireSuccess()

	b := out.Bytes()
	t.Log(string(b))
	var report healthsdk.ClientNetcheckReport
	require.NoError(t, json.Unmarshal(b, &report))

	// We do not assert that the report is healthy, just that
	// it has the expected number of reports per region.
	require.Len(t, report.DERP.Regions, 1+1) // 1 built-in region + 1 test-managed STUN region
	for _, v := range report.DERP.Regions {
		require.Len(t, v.NodeReports, len(v.Region.Nodes))
	}
}
