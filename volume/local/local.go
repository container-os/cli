// Package local provides the default implementation for volumes. It
// is used to mount data volume containers and directories local to
// the host server.
package local

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/docker/volume"
)

// VolumeDataPathName is the name of the directory where the volume data is stored.
// It uses a very distintive name to avoid collisions migrating data between
// Docker versions.
const (
	VolumeDataPathName = "_data"
	volumesPathName    = "volumes"
)

var oldVfsDir = filepath.Join("vfs", "dir")

// New instantiates a new Root instance with the provided scope. Scope
// is the base path that the Root instance uses to store its
// volumes. The base path is created here if it does not exist.
func New(scope string) (*Root, error) {
	rootDirectory := filepath.Join(scope, volumesPathName)

	if err := os.MkdirAll(rootDirectory, 0700); err != nil {
		return nil, err
	}

	r := &Root{
		scope:   scope,
		path:    rootDirectory,
		volumes: make(map[string]*localVolume),
	}

	dirs, err := ioutil.ReadDir(rootDirectory)
	if err != nil {
		return nil, err
	}

	for _, d := range dirs {
		name := filepath.Base(d.Name())
		r.volumes[name] = &localVolume{
			driverName: r.Name(),
			name:       name,
			path:       r.DataPath(name),
		}
	}
	return r, nil
}

// Root implements the Driver interface for the volume package and
// manages the creation/removal of volumes. It uses only standard vfs
// commands to create/remove dirs within its provided scope.
type Root struct {
	m       sync.Mutex
	scope   string
	path    string
	volumes map[string]*localVolume
}

// DataPath returns the constructed path of this volume.
func (r *Root) DataPath(volumeName string) string {
	return filepath.Join(r.path, volumeName, VolumeDataPathName)
}

// Name returns the name of Root, defined in the volume package in the DefaultDriverName constant.
func (r *Root) Name() string {
	return volume.DefaultDriverName
}

// Create creates a new volume.Volume with the provided name, creating
// the underlying directory tree required for this volume in the
// process.
func (r *Root) Create(name string) (volume.Volume, error) {
	r.m.Lock()
	defer r.m.Unlock()

	v, exists := r.volumes[name]
	if !exists {
		path := r.DataPath(name)
		if err := os.MkdirAll(path, 0755); err != nil {
			if os.IsExist(err) {
				return nil, fmt.Errorf("volume already exists under %s", filepath.Dir(path))
			}
			return nil, err
		}
		v = &localVolume{
			driverName: r.Name(),
			name:       name,
			path:       path,
		}
		r.volumes[name] = v
	}
	v.use()
	return v, nil
}

// Remove removes the specified volume and all underlying data. If the
// given volume does not belong to this driver and an error is
// returned. The volume is reference counted, if all references are
// not released then the volume is not removed.
func (r *Root) Remove(v volume.Volume) error {
	r.m.Lock()
	defer r.m.Unlock()
	lv, ok := v.(*localVolume)
	if !ok {
		return errors.New("unknown volume type")
	}
	lv.release()
	if lv.usedCount == 0 {
		realPath, err := filepath.EvalSymlinks(lv.path)
		if err != nil {
			return err
		}
		if !r.scopedPath(realPath) {
			return fmt.Errorf("Unable to remove a directory of out the Docker root: %s", realPath)
		}

		if err := os.RemoveAll(realPath); err != nil {
			return err
		}

		delete(r.volumes, lv.name)
		return os.RemoveAll(filepath.Dir(lv.path))
	}
	return nil
}

// scopedPath verifies that the path where the volume is located
// is under Docker's root and the valid local paths.
func (r *Root) scopedPath(realPath string) bool {
	// Volumes path for Docker version >= 1.7
	if strings.HasPrefix(realPath, filepath.Join(r.scope, volumesPathName)) {
		return true
	}

	// Volumes path for Docker version < 1.7
	if strings.HasPrefix(realPath, filepath.Join(r.scope, oldVfsDir)) {
		return true
	}

	return false
}

// localVolume implements the Volume interface from the volume package and
// represents the volumes created by Root.
type localVolume struct {
	m         sync.Mutex
	usedCount int
	// unique name of the volume
	name string
	// path is the path on the host where the data lives
	path string
	// driverName is the name of the driver that created the volume.
	driverName string
}

// Name returns the name of the given Volume.
func (v *localVolume) Name() string {
	return v.name
}

// DriverName returns the driver that created the given Volume.
func (v *localVolume) DriverName() string {
	return v.driverName
}

// Path returns the data location.
func (v *localVolume) Path() string {
	return v.path
}

// Mount implements the localVolume interface, returning the data location.
func (v *localVolume) Mount() (string, error) {
	return v.path, nil
}

// Umount is for satisfying the localVolume interface and does not do anything in this driver.
func (v *localVolume) Unmount() error {
	return nil
}

func (v *localVolume) use() {
	v.m.Lock()
	v.usedCount++
	v.m.Unlock()
}

func (v *localVolume) release() {
	v.m.Lock()
	v.usedCount--
	v.m.Unlock()
}
