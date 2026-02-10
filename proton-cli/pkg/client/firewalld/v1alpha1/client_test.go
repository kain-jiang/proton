package v1alpha1

import (
	"testing"

	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/global"
)

func TestIPSet(t *testing.T) {
	global.LoggerLevel = logrus.DebugLevel.String()

	a := require.New(t)

	name := "test-" + lo.RandomString(4, lo.NumbersCharset)

	f := &Client{}

	if f.executor == nil {
		t.Skip("Client.executor is missing")
	}

	// P: false, R: false
	requireIPSet(t, f, false, name, false)
	requireIPSet(t, f, true, name, false)

	// 创建 IPSet
	a.NoError(f.NewIPSet(name, IPSetType{Method: IPSetMethodHash, DataTypes: []IPSetDataType{IPSetDataTypeIP}}, IPSetFamilyINet))
	// P: true, R: false
	requireIPSet(t, f, false, name, false)
	requireIPSet(t, f, true, name, true)

	// 重载
	a.NoError(f.Reload())
	// P: true, R: true
	requireIPSet(t, f, false, name, true)
	requireIPSet(t, f, true, name, true)

	// P：添加 entry
	a.NoError(f.AddIPSetEntries(true, name, []string{"192.168.0.1", "192.168.0.2"}))
	requireIPSetEntries(t, f, true, name, []string{"192.168.0.1", "192.168.0.2"})
	requireIPSetEntries(t, f, false, name, nil)

	// R：添加 entry
	a.NoError(f.AddIPSetEntries(false, name, []string{"192.168.0.3", "192.168.0.4"}))
	requireIPSetEntries(t, f, true, name, []string{"192.168.0.1", "192.168.0.2"})
	requireIPSetEntries(t, f, false, name, []string{"192.168.0.3", "192.168.0.4"})

	// 重载
	a.NoError(f.Reload())
	requireIPSetEntries(t, f, true, name, []string{"192.168.0.1", "192.168.0.2"})
	requireIPSetEntries(t, f, false, name, []string{"192.168.0.1", "192.168.0.2"})

	// 删除 IPSet
	a.NoError(f.DeleteIPSet(name))
	// P: false, R: true
	requireIPSet(t, f, false, name, true)
	requireIPSet(t, f, true, name, false)

	// 重载
	a.NoError(f.Reload())
	// P: false, R: true
	requireIPSet(t, f, false, name, false)
	requireIPSet(t, f, true, name, false)
}

func requireIPSet(t testing.TB, f Interface, p bool, s string, want bool) {
	t.Helper()

	ipsets, err := f.GetIPSets(p)
	require.NoError(t, err)
	if want {
		require.Contains(t, ipsets, s)
	} else {
		require.NotContains(t, ipsets, s)
	}
}

func requireIPSetEntries(t testing.TB, f Interface, p bool, s string, want []string) {
	t.Helper()

	got, err := f.GetIPSetEntries(p, s)
	require.NoError(t, err)
	require.ElementsMatch(t, want, got)
}

func TestZone(t *testing.T) {
	global.LoggerLevel = "debug"

	z := "test-0000"
	t.Log("firewalld zone:", z)

	f := &Client{}

	if f.executor == nil {
		t.Skip("Client.executor is missing")
	}

	// P: false, R: false
	requireZone(t, f, z, false, false)

	// Create zone, P: true, R: false, Target: default, Interface[P]: empty
	require.NoError(t, f.NewZone(z))
	requireZone(t, f, z, true, false)
	requireZoneTarget(t, f, z, TargetDefault)
	requireZoneInterfaces(t, f, z, []string{}, nil)

	// Set Target, Target: ACCEPT
	require.NoError(t, f.SetZoneTarget(z, TargetAccept))
	requireZoneTarget(t, f, z, TargetAccept)

	// Reload, P: true, R: true, Interface[P|R]: empty
	require.NoError(t, f.Reload())
	requireZone(t, f, z, true, true)
	requireZoneInterfaces(t, f, z, []string{}, []string{})

	// P: Add interface v0, v1
	require.NoError(t, f.AddZoneInterfaces(true, z, []string{"v0", "v1"}))
	requireZoneInterfaces(t, f, z, []string{"v0", "v1"}, []string{})

	// R: Add interface v2, v3
	require.NoError(t, f.AddZoneInterfaces(false, z, []string{"v2", "v3"}))
	requireZoneInterfaces(t, f, z, []string{"v0", "v1"}, []string{"v2", "v3"})

	// Reload, interfaces: v0, v1
	require.NoError(t, f.Reload())
	requireZoneInterfaces(t, f, z, []string{"v0", "v1"}, []string{"v0", "v1"})

	// P: Add sources: 192.168.0.1, 192.168.0.2
	require.NoError(t, f.AddZoneSources(true, z, []string{"192.168.0.1", "192.168.0.2"}))
	requireZoneSources(t, f, z, []string{"192.168.0.1", "192.168.0.2"}, nil)

	// R: Add sources: 192.168.0.3, 192.168.0.4
	require.NoError(t, f.AddZoneSources(false, z, []string{"192.168.0.3", "192.168.0.4"}))
	requireZoneSources(t, f, z, []string{"192.168.0.1", "192.168.0.2"}, []string{"192.168.0.3", "192.168.0.4"})

	// Reload, sources: 192.168.0.1, 192.168.0.2
	require.NoError(t, f.Reload())
	requireZoneSources(t, f, z, []string{"192.168.0.1", "192.168.0.2"}, []string{"192.168.0.1", "192.168.0.2"})

	// Delete zone, P: false, R: true
	require.NoError(t, f.DeleteZone(z))
	requireZone(t, f, z, false, true)

	// Reload, P: false, R: false
	require.NoError(t, f.Reload())
	requireZone(t, f, z, false, false)
}

func requireZone(t testing.TB, f Interface, z string, permanent, runtime bool) {
	t.Helper()

	// permanent
	{
		zones, err := f.GetZones(true)
		require.NoError(t, err)
		r := lo.Ternary(permanent, require.Contains, require.NotContains)
		r(t, zones, z)
	}

	// runtime
	{
		zones, err := f.GetZones(false)
		require.NoError(t, err)
		r := lo.Ternary(runtime, require.Contains, require.NotContains)
		r(t, zones, z)
	}
}

func requireZoneTarget(t testing.TB, f Interface, z string, want Target) {
	t.Helper()

	got, err := f.GetZoneTarget(z)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func requireZoneInterfaces(t testing.TB, f Interface, z string, permanent, runtime []string) {
	t.Helper()

	// permanent
	{
		got, err := f.ListZoneInterfaces(true, z)
		require.NoError(t, err)
		require.ElementsMatch(t, permanent, got)
	}

	if runtime == nil {
		return
	}

	// runtime
	{
		got, err := f.ListZoneInterfaces(false, z)
		require.NoError(t, err)
		require.ElementsMatch(t, runtime, got)
	}
}

func requireZoneSources(t testing.TB, f Interface, z string, permanent, runtime []string) {
	t.Helper()

	// permanent
	{
		got, err := f.ListZoneSources(true, z)
		require.NoError(t, err)
		require.ElementsMatch(t, permanent, got)
	}

	if runtime == nil {
		return
	}

	// runtime
	{
		got, err := f.ListZoneSources(false, z)
		require.NoError(t, err)
		require.ElementsMatch(t, runtime, got)
	}
}
