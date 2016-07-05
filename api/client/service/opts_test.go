package service

import (
	"testing"
	"time"

	"github.com/docker/docker/pkg/testutil/assert"
	"github.com/docker/engine-api/types/swarm"
)

func TestMemBytesString(t *testing.T) {
	var mem memBytes = 1048576
	assert.Equal(t, mem.String(), "1 MiB")
}

func TestMemBytesSetAndValue(t *testing.T) {
	var mem memBytes
	assert.NilError(t, mem.Set("5kb"))
	assert.Equal(t, mem.Value(), int64(5120))
}

func TestNanoCPUsString(t *testing.T) {
	var cpus nanoCPUs = 6100000000
	assert.Equal(t, cpus.String(), "6.100")
}

func TestNanoCPUsSetAndValue(t *testing.T) {
	var cpus nanoCPUs
	assert.NilError(t, cpus.Set("0.35"))
	assert.Equal(t, cpus.Value(), int64(350000000))
}

func TestDurationOptString(t *testing.T) {
	dur := time.Duration(300 * 10e8)
	duration := DurationOpt{value: &dur}
	assert.Equal(t, duration.String(), "5m0s")
}

func TestDurationOptSetAndValue(t *testing.T) {
	var duration DurationOpt
	assert.NilError(t, duration.Set("300s"))
	assert.Equal(t, *duration.Value(), time.Duration(300*10e8))
}

func TestUint64OptString(t *testing.T) {
	value := uint64(2345678)
	opt := Uint64Opt{value: &value}
	assert.Equal(t, opt.String(), "2345678")

	opt = Uint64Opt{}
	assert.Equal(t, opt.String(), "none")
}

func TestUint64OptSetAndValue(t *testing.T) {
	var opt Uint64Opt
	assert.NilError(t, opt.Set("14445"))
	assert.Equal(t, *opt.Value(), uint64(14445))
}

func TestMountOptString(t *testing.T) {
	mount := MountOpt{
		values: []swarm.Mount{
			{
				Type:   swarm.MountType("BIND"),
				Source: "/home/path",
				Target: "/target",
			},
			{
				Type:   swarm.MountType("VOLUME"),
				Source: "foo",
				Target: "/target/foo",
			},
		},
	}
	expected := "BIND /home/path /target, VOLUME foo /target/foo"
	assert.Equal(t, mount.String(), expected)
}

func TestMountOptSetNoError(t *testing.T) {
	var mount MountOpt
	assert.NilError(t, mount.Set("type=bind,target=/target,source=/foo"))

	mounts := mount.Value()
	assert.Equal(t, len(mounts), 1)
	assert.Equal(t, mounts[0], swarm.Mount{
		Type:   swarm.MountType("BIND"),
		Source: "/foo",
		Target: "/target",
	})
}

func TestMountOptSetErrorNoType(t *testing.T) {
	var mount MountOpt
	assert.Error(t, mount.Set("target=/target,source=/foo"), "type is required")
}

func TestMountOptSetErrorNoTarget(t *testing.T) {
	var mount MountOpt
	assert.Error(t, mount.Set("type=VOLUME,source=/foo"), "target is required")
}

func TestMountOptSetErrorInvalidKey(t *testing.T) {
	var mount MountOpt
	assert.Error(t, mount.Set("type=VOLUME,bogus=foo"), "unexpected key 'bogus'")
}

func TestMountOptSetErrorInvalidField(t *testing.T) {
	var mount MountOpt
	assert.Error(t, mount.Set("type=VOLUME,bogus"), "invalid field 'bogus'")
}

func TestMountOptSetErrorInvalidWritable(t *testing.T) {
	var mount MountOpt
	assert.Error(t, mount.Set("type=VOLUME,readonly=no"), "invalid value for readonly: no")
}

func TestMountOptDefaultEnableWritable(t *testing.T) {
	var m MountOpt
	assert.NilError(t, m.Set("type=bind,target=/foo,source=/foo"))
	assert.Equal(t, m.values[0].ReadOnly, false)

	m = MountOpt{}
	assert.NilError(t, m.Set("type=bind,target=/foo,source=/foo,readonly"))
	assert.Equal(t, m.values[0].ReadOnly, true)

	m = MountOpt{}
	assert.NilError(t, m.Set("type=bind,target=/foo,source=/foo,readonly=1"))
	assert.Equal(t, m.values[0].ReadOnly, true)

	m = MountOpt{}
	assert.NilError(t, m.Set("type=bind,target=/foo,source=/foo,readonly=0"))
	assert.Equal(t, m.values[0].ReadOnly, false)
}

func TestMountOptVolumeNoCopy(t *testing.T) {
	var m MountOpt
	assert.Error(t, m.Set("type=volume,target=/foo,volume-nocopy"), "source is required")

	m = MountOpt{}
	assert.NilError(t, m.Set("type=volume,target=/foo,source=foo"))
	assert.Equal(t, m.values[0].VolumeOptions == nil, true)

	m = MountOpt{}
	assert.NilError(t, m.Set("type=volume,target=/foo,source=foo,volume-nocopy=true"))
	assert.Equal(t, m.values[0].VolumeOptions != nil, true)
	assert.Equal(t, m.values[0].VolumeOptions.NoCopy, true)

	m = MountOpt{}
	assert.NilError(t, m.Set("type=volume,target=/foo,source=foo,volume-nocopy"))
	assert.Equal(t, m.values[0].VolumeOptions != nil, true)
	assert.Equal(t, m.values[0].VolumeOptions.NoCopy, true)

	m = MountOpt{}
	assert.NilError(t, m.Set("type=volume,target=/foo,source=foo,volume-nocopy=1"))
	assert.Equal(t, m.values[0].VolumeOptions != nil, true)
	assert.Equal(t, m.values[0].VolumeOptions.NoCopy, true)
}

func TestMountOptTypeConflict(t *testing.T) {
	var m MountOpt
	assert.Error(t, m.Set("type=bind,target=/foo,source=/foo,volume-nocopy=true"), "cannot mix")
	assert.Error(t, m.Set("type=volume,target=/foo,source=/foo,bind-propagation=rprivate"), "cannot mix")
}
