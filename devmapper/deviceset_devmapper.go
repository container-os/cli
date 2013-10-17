package devmapper

import (
	"time"
	"encoding/json"
	"fmt"
	"github.com/dotcloud/docker/utils"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
)

var (
	DefaultDataLoopbackSize     int64  = 100 * 1024 * 1024 * 1024
	DefaultMetaDataLoopbackSize int64  = 2 * 1024 * 1024 * 1024
	DefaultBaseFsSize           uint64 = 10 * 1024 * 1024 * 1024
)

type DevInfo struct {
	Hash          string       `json:"-"`
	DeviceId      int          `json:"device_id"`
	Size          uint64       `json:"size"`
	TransactionId uint64       `json:"transaction_id"`
	Initialized   bool         `json:"initialized"`
	devices       *DeviceSetDM `json:"-"`
}

type MetaData struct {
	Devices map[string]*DevInfo `json:devices`
}

type DeviceSetDM struct {
	MetaData
	sync.Mutex
	initialized      bool
	root             string
	devicePrefix     string
	TransactionId    uint64
	NewTransactionId uint64
	nextFreeDevice   int
	activeMounts     map[string]int
}

func getDevName(name string) string {
	return "/dev/mapper/" + name
}

func (info *DevInfo) Name() string {
	hash := info.Hash
	if hash == "" {
		hash = "base"
	}
	return fmt.Sprintf("%s-%s", info.devices.devicePrefix, hash)
}

func (info *DevInfo) DevName() string {
	return getDevName(info.Name())
}

func (devices *DeviceSetDM) loopbackDir() string {
	return path.Join(devices.root, "devicemapper")
}

func (devices *DeviceSetDM) jsonFile() string {
	return path.Join(devices.loopbackDir(), "json")
}

func (devices *DeviceSetDM) getPoolName() string {
	return devices.devicePrefix + "-pool"
}

func (devices *DeviceSetDM) getPoolDevName() string {
	return getDevName(devices.getPoolName())
}

func (devices *DeviceSetDM) hasImage(name string) bool {
	dirname := devices.loopbackDir()
	filename := path.Join(dirname, name)

	_, err := os.Stat(filename)
	return err == nil
}

// ensureImage creates a sparse file of <size> bytes at the path
// <root>/devicemapper/<name>.
// If the file already exists, it does nothing.
// Either way it returns the full path.
func (devices *DeviceSetDM) ensureImage(name string, size int64) (string, error) {
	dirname := devices.loopbackDir()
	filename := path.Join(dirname, name)

	if err := os.MkdirAll(dirname, 0700); err != nil && !os.IsExist(err) {
		return "", err
	}

	if _, err := os.Stat(filename); err != nil {
		if !os.IsNotExist(err) {
			return "", err
		}
		utils.Debugf("Creating loopback file %s for device-manage use", filename)
		file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return "", err
		}

		if err = file.Truncate(size); err != nil {
			return "", err
		}
	}
	return filename, nil
}

func (devices *DeviceSetDM) allocateDeviceId() int {
	// TODO: Add smarter reuse of deleted devices
	id := devices.nextFreeDevice
	devices.nextFreeDevice = devices.nextFreeDevice + 1
	return id
}

func (devices *DeviceSetDM) allocateTransactionId() uint64 {
	devices.NewTransactionId = devices.NewTransactionId + 1
	return devices.NewTransactionId
}

func (devices *DeviceSetDM) saveMetadata() error {
	jsonData, err := json.Marshal(devices.MetaData)
	if err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}
	tmpFile, err := ioutil.TempFile(filepath.Dir(devices.jsonFile()), ".json")
	if err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	n, err := tmpFile.Write(jsonData)
	if err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}
	if n < len(jsonData) {
		return io.ErrShortWrite
	}
	if err := tmpFile.Sync(); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}
	if err := os.Rename(tmpFile.Name(), devices.jsonFile()); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	if devices.NewTransactionId != devices.TransactionId {
		if err = setTransactionId(devices.getPoolDevName(), devices.TransactionId, devices.NewTransactionId); err != nil {
			utils.Debugf("\n--->Err: %s\n", err)
			return err
		}
		devices.TransactionId = devices.NewTransactionId
	}
	return nil
}

func (devices *DeviceSetDM) registerDevice(id int, hash string, size uint64) (*DevInfo, error) {
	utils.Debugf("registerDevice(%v, %v)", id, hash)
	info := &DevInfo{
		Hash:          hash,
		DeviceId:      id,
		Size:          size,
		TransactionId: devices.allocateTransactionId(),
		Initialized:   false,
		devices:       devices,
	}

	devices.Devices[hash] = info
	if err := devices.saveMetadata(); err != nil {
		// Try to remove unused device
		delete(devices.Devices, hash)
		return nil, err
	}

	return info, nil
}

func (devices *DeviceSetDM) activateDeviceIfNeeded(hash string) error {
	utils.Debugf("activateDeviceIfNeeded(%v)", hash)
	info := devices.Devices[hash]
	if info == nil {
		return fmt.Errorf("Unknown device %s", hash)
	}

	if devinfo, _ := getInfo(info.Name()); devinfo != nil && devinfo.Exists != 0 {
		return nil
	}

	return activateDevice(devices.getPoolDevName(), info.Name(), info.DeviceId, info.Size)
}

func (devices *DeviceSetDM) createFilesystem(info *DevInfo) error {
	devname := info.DevName()

	err := exec.Command("mkfs.ext4", "-E", "discard,lazy_itable_init=0,lazy_journal_init=0", devname).Run()
	if err != nil {
		err = exec.Command("mkfs.ext4", "-E", "discard,lazy_itable_init=0", devname).Run()
	}
	if err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}
	return nil
}

func (devices *DeviceSetDM) loadMetaData() error {
	utils.Debugf("loadMetadata()")
	defer utils.Debugf("loadMetadata END")
	_, _, _, params, err := getStatus(devices.getPoolName())
	if err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	if _, err := fmt.Sscanf(params, "%d", &devices.TransactionId); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}
	devices.NewTransactionId = devices.TransactionId

	jsonData, err := ioutil.ReadFile(devices.jsonFile())
	if err != nil && !os.IsNotExist(err) {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	devices.MetaData.Devices = make(map[string]*DevInfo)
	if jsonData != nil {
		if err := json.Unmarshal(jsonData, &devices.MetaData); err != nil {
			utils.Debugf("\n--->Err: %s\n", err)
			return err
		}
	}

	for hash, d := range devices.Devices {
		d.Hash = hash
		d.devices = devices

		if d.DeviceId >= devices.nextFreeDevice {
			devices.nextFreeDevice = d.DeviceId + 1
		}

		// If the transaction id is larger than the actual one we lost the device due to some crash
		if d.TransactionId > devices.TransactionId {
			utils.Debugf("Removing lost device %s with id %d", hash, d.TransactionId)
			delete(devices.Devices, hash)
		}
	}
	return nil
}

func (devices *DeviceSetDM) setupBaseImage() error {
	oldInfo := devices.Devices[""]
	if oldInfo != nil && oldInfo.Initialized {
		return nil
	}

	if oldInfo != nil && !oldInfo.Initialized {
		utils.Debugf("Removing uninitialized base image")
		if err := devices.removeDevice(""); err != nil {
			utils.Debugf("\n--->Err: %s\n", err)
			return err
		}
	}

	utils.Debugf("Initializing base device-manager snapshot")

	id := devices.allocateDeviceId()

	// Create initial device
	if err := createDevice(devices.getPoolDevName(), id); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	utils.Debugf("Registering base device (id %v) with FS size %v", id, DefaultBaseFsSize)
	info, err := devices.registerDevice(id, "", DefaultBaseFsSize)
	if err != nil {
		_ = deleteDevice(devices.getPoolDevName(), id)
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	utils.Debugf("Creating filesystem on base device-manager snapshot")

	if err = devices.activateDeviceIfNeeded(""); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	if err := devices.createFilesystem(info); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	info.Initialized = true
	if err = devices.saveMetadata(); err != nil {
		info.Initialized = false
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	return nil
}

func setCloseOnExec(name string) {
	fileInfos, _ := ioutil.ReadDir("/proc/self/fd")
	if fileInfos != nil {
		for _, i := range fileInfos {
			link, _ := os.Readlink(filepath.Join("/proc/self/fd", i.Name()))
			if link == name {
				fd, err := strconv.Atoi(i.Name())
				if err == nil {
					syscall.CloseOnExec(fd)
				}
			}
		}
	}
}

func (devices *DeviceSetDM) log(level int, file string, line int, dmError int, message string) {
	if level >= 7 {
		return // Ignore _LOG_DEBUG
	}

	utils.Debugf("libdevmapper(%d): %s:%d (%d) %s", level, file, line, dmError, message)
}

func (devices *DeviceSetDM) initDevmapper() error {
	logInit(devices)

	// Make sure the sparse images exist in <root>/devicemapper/data and
	// <root>/devicemapper/metadata

	createdLoopback := !devices.hasImage("data") || !devices.hasImage("metadata")
	data, err := devices.ensureImage("data", DefaultDataLoopbackSize)
	if err != nil {
		utils.Debugf("Error device ensureImage (data): %s\n", err)
		return err
	}
	metadata, err := devices.ensureImage("metadata", DefaultMetaDataLoopbackSize)
	if err != nil {
		utils.Debugf("Error device ensureImage (metadata): %s\n", err)
		return err
	}

	// Set the device prefix from the device id and inode of the data image

	st, err := os.Stat(data)
	if err != nil {
		return fmt.Errorf("Error looking up data image %s: %s", data, err)
	}
	sysSt := st.Sys().(*syscall.Stat_t)
	// "reg-" stands for "regular file".
	// In the future we might use "dev-" for "device file", etc.
	devices.devicePrefix = fmt.Sprintf("docker-reg-%d-%d", sysSt.Dev, sysSt.Ino)


	// Check for the existence of the device <prefix>-pool
	utils.Debugf("Checking for existence of the pool '%s'", devices.getPoolName())
	info, err := getInfo(devices.getPoolName())
	if info == nil {
		utils.Debugf("Error device getInfo: %s", err)
		return err
	}

	// It seems libdevmapper opens this without O_CLOEXEC, and go exec will not close files
	// that are not Close-on-exec, and lxc-start will die if it inherits any unexpected files,
	// so we add this badhack to make sure it closes itself
	setCloseOnExec("/dev/mapper/control")

	// If the pool doesn't exist, create it
	if info.Exists == 0 {
		utils.Debugf("Pool doesn't exist. Creating it.")
		dataFile, err := AttachLoopDevice(data)
		if err != nil {
			utils.Debugf("\n--->Err: %s\n", err)
			return err
		}
		defer dataFile.Close()

		metadataFile, err := AttachLoopDevice(metadata)
		if err != nil {
			utils.Debugf("\n--->Err: %s\n", err)
			return err
		}
		defer metadataFile.Close()

		if err := createPool(devices.getPoolName(), dataFile, metadataFile); err != nil {
			utils.Debugf("\n--->Err: %s\n", err)
			return err
		}
	}

	// If we didn't just create the data or metadata image, we need to
	// load the metadata from the existing file.
	if !createdLoopback {
		if err = devices.loadMetaData(); err != nil {
			utils.Debugf("\n--->Err: %s\n", err)
			return err
		}
	}

	// Setup the base image
	if err := devices.setupBaseImage(); err != nil {
		utils.Debugf("Error device setupBaseImage: %s\n", err)
		return err
	}

	return nil
}

func (devices *DeviceSetDM) AddDevice(hash, baseHash string) error {
	devices.Lock()
	defer devices.Unlock()

	if err := devices.ensureInit(); err != nil {
		utils.Debugf("Error init: %s\n", err)
		return err
	}

	if devices.Devices[hash] != nil {
		return fmt.Errorf("hash %s already exists", hash)
	}

	baseInfo := devices.Devices[baseHash]
	if baseInfo == nil {
		utils.Debugf("Base Hash not found")
		return fmt.Errorf("Unknown base hash %s", baseHash)
	}

	deviceId := devices.allocateDeviceId()

	if err := devices.createSnapDevice(devices.getPoolDevName(), deviceId, baseInfo.Name(), baseInfo.DeviceId); err != nil {
		utils.Debugf("Error creating snap device: %s\n", err)
		return err
	}

	if _, err := devices.registerDevice(deviceId, hash, baseInfo.Size); err != nil {
		deleteDevice(devices.getPoolDevName(), deviceId)
		utils.Debugf("Error registering device: %s\n", err)
		return err
	}
	return nil
}

func (devices *DeviceSetDM) removeDevice(hash string) error {
	info := devices.Devices[hash]
	if info == nil {
		return fmt.Errorf("hash %s doesn't exists", hash)
	}

	devinfo, _ := getInfo(info.Name())
	if devinfo != nil && devinfo.Exists != 0 {
		if err := removeDevice(info.Name()); err != nil {
			utils.Debugf("Error removing device: %s\n", err)
			return err
		}
	}

	if info.Initialized {
		info.Initialized = false
		if err := devices.saveMetadata(); err != nil {
			utils.Debugf("Error saving meta data: %s\n", err)
			return err
		}
	}

	if err := deleteDevice(devices.getPoolDevName(), info.DeviceId); err != nil {
		utils.Debugf("Error deleting device: %s\n", err)
		return err
	}

	devices.allocateTransactionId()
	delete(devices.Devices, info.Hash)

	if err := devices.saveMetadata(); err != nil {
		devices.Devices[info.Hash] = info
		utils.Debugf("Error saving meta data: %s\n", err)
		return err
	}

	return nil
}

func (devices *DeviceSetDM) RemoveDevice(hash string) error {
	devices.Lock()
	defer devices.Unlock()

	if err := devices.ensureInit(); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	return devices.removeDevice(hash)
}

func (devices *DeviceSetDM) deactivateDevice(hash string) error {
	utils.Debugf("[devmapper] deactivateDevice(%s)", hash)
	defer utils.Debugf("[devmapper] deactivateDevice END")
	var devname string
	// FIXME: shouldn't we just register the pool into devices?
	devname, err := devices.byHash(hash)
	if err != nil {
		return err
	}
	devinfo, err := getInfo(devname)
	if err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}
	if devinfo.Exists != 0 {
		if err := removeDevice(devname); err != nil {
			utils.Debugf("\n--->Err: %s\n", err)
			return err
		}
		if err := devices.waitRemove(hash); err != nil {
			return err
		}
	}

	return nil
}

// waitRemove blocks until either:
// a) the device registered at <device_set_prefix>-<hash> is removed,
// or b) the 1 second timeout expires.
func (devices *DeviceSetDM) waitRemove(hash string) error {
	utils.Debugf("[deviceset %s] waitRemove(%s)", devices.devicePrefix, hash)
	defer utils.Debugf("[deviceset %s] waitRemove END", devices.devicePrefix, hash)
	devname, err := devices.byHash(hash)
	if err != nil {
		return err
	}
	i := 0
	for ; i<1000; i+=1 {
		devinfo, err := getInfo(devname)
		if err != nil {
			// If there is an error we assume the device doesn't exist.
			// The error might actually be something else, but we can't differentiate.
			return nil
		}
		utils.Debugf("Waiting for removal of %s: exists=%d", devname, devinfo.Exists)
		if devinfo.Exists == 0 {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	if i == 1000 {
		return fmt.Errorf("Timeout while waiting for device %s to be removed", devname)
	}
	return nil
}

// waitClose blocks until either:
// a) the device registered at <device_set_prefix>-<hash> is closed,
// or b) the 1 second timeout expires.
func (devices *DeviceSetDM) waitClose(hash string) error {
	devname, err := devices.byHash(hash)
	if err != nil {
		return err
	}
	i := 0
	for ; i<1000; i+=1 {
		devinfo, err := getInfo(devname)
		if err != nil {
			return err
		}
		utils.Debugf("Waiting for unmount of %s: opencount=%d", devname, devinfo.OpenCount)
		if devinfo.OpenCount == 0 {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}
	if i == 1000 {
		return fmt.Errorf("Timeout while waiting for device %s to close", devname)
	}
	return nil
}

// byHash is a hack to allow looking up the deviceset's pool by the hash "pool".
// FIXME: it seems probably cleaner to register the pool in devices.Devices,
// but I am afraid of arcane implications deep in the devicemapper code,
// so this will do.
func (devices *DeviceSetDM) byHash(hash string) (devname string, err error) {
	if hash == "pool" {
		return devices.getPoolDevName(), nil
	}
	info := devices.Devices[hash]
	if info == nil {
		return "", fmt.Errorf("hash %s doesn't exists", hash)
	}
	return info.Name(), nil
}


func (devices *DeviceSetDM) Shutdown() error {
	utils.Debugf("[deviceset %s] shutdown()", devices.devicePrefix)
	defer utils.Debugf("[deviceset %s] shutdown END", devices.devicePrefix)
	devices.Lock()
	utils.Debugf("[devmapper] Shutting down DeviceSet: %s", devices.root)
	defer devices.Unlock()

	if !devices.initialized {
		return nil
	}

	for path, count := range devices.activeMounts {
		for i := count; i > 0; i-- {
			if err := syscall.Unmount(path, 0); err != nil {
				utils.Debugf("Shutdown unmounting %s, error: %s\n", path, err)
			}
		}
		delete(devices.activeMounts, path)
	}

	for _, d := range devices.Devices {
		if err := devices.waitClose(d.Hash); err != nil {
			utils.Errorf("Warning: error waiting for device %s to unmount: %s\n", d.Hash, err)
		}
		if err := devices.deactivateDevice(d.Hash); err != nil {
			utils.Debugf("Shutdown deactivate %s , error: %s\n", d.Hash, err)
		}
	}

	pool := devices.getPoolDevName()
	if devinfo, err := getInfo(pool); err == nil && devinfo.Exists != 0 {
		if err := devices.deactivateDevice("pool"); err != nil {
			utils.Debugf("Shutdown deactivate %s , error: %s\n", pool, err)
		}
	}

	return nil
}

func (devices *DeviceSetDM) MountDevice(hash, path string) error {
	devices.Lock()
	defer devices.Unlock()

	if err := devices.ensureInit(); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	if err := devices.activateDeviceIfNeeded(hash); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	info := devices.Devices[hash]

	err := syscall.Mount(info.DevName(), path, "ext4", syscall.MS_MGC_VAL, "discard")
	if err != nil && err == syscall.EINVAL {
		err = syscall.Mount(info.DevName(), path, "ext4", syscall.MS_MGC_VAL, "")
	}
	if err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	count := devices.activeMounts[path]
	devices.activeMounts[path] = count + 1

	return nil
}

func (devices *DeviceSetDM) UnmountDevice(hash, path string, deactivate bool) error {
	utils.Debugf("[devmapper] UnmountDevice(hash=%s path=%s)", hash, path)
	defer utils.Debugf("[devmapper] UnmountDevice END")
	devices.Lock()
	defer devices.Unlock()

	utils.Debugf("[devmapper] Unmount(%s)", path)
	if err := syscall.Unmount(path, 0); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}
	utils.Debugf("[devmapper] Unmount done")
	// Wait for the unmount to be effective,
	// by watching the value of Info.OpenCount for the device
	if err := devices.waitClose(hash); err != nil {
		return err
	}

	if count := devices.activeMounts[path]; count > 1 {
		devices.activeMounts[path] = count - 1
	} else {
		delete(devices.activeMounts, path)
	}

	if deactivate {
		devices.deactivateDevice(hash)
	}

	return nil
}

func (devices *DeviceSetDM) HasDevice(hash string) bool {
	devices.Lock()
	defer devices.Unlock()

	if err := devices.ensureInit(); err != nil {
		return false
	}
	return devices.Devices[hash] != nil
}

func (devices *DeviceSetDM) HasInitializedDevice(hash string) bool {
	devices.Lock()
	defer devices.Unlock()

	if err := devices.ensureInit(); err != nil {
		return false
	}

	info := devices.Devices[hash]
	return info != nil && info.Initialized
}

func (devices *DeviceSetDM) HasActivatedDevice(hash string) bool {
	devices.Lock()
	defer devices.Unlock()

	if err := devices.ensureInit(); err != nil {
		return false
	}

	info := devices.Devices[hash]
	if info == nil {
		return false
	}
	devinfo, _ := getInfo(info.Name())
	return devinfo != nil && devinfo.Exists != 0
}

func (devices *DeviceSetDM) SetInitialized(hash string) error {
	devices.Lock()
	defer devices.Unlock()

	if err := devices.ensureInit(); err != nil {
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	info := devices.Devices[hash]
	if info == nil {
		return fmt.Errorf("Unknown device %s", hash)
	}

	info.Initialized = true
	if err := devices.saveMetadata(); err != nil {
		info.Initialized = false
		utils.Debugf("\n--->Err: %s\n", err)
		return err
	}

	return nil
}

func (devices *DeviceSetDM) ensureInit() error {
	if !devices.initialized {
		devices.initialized = true
		if err := devices.initDevmapper(); err != nil {
			utils.Debugf("\n--->Err: %s\n", err)
			return err
		}
	}
	return nil
}

func NewDeviceSetDM(root string) *DeviceSetDM {
	SetDevDir("/dev")

	return &DeviceSetDM{
		initialized:  false,
		root:         root,
		MetaData:     MetaData{Devices: make(map[string]*DevInfo)},
		activeMounts: make(map[string]int),
	}
}
