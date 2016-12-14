package graphdriver

import (
	"errors"
	"fmt"
	"io"

	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/plugingetter"
)

type graphDriverProxy struct {
	name   string
	client pluginClient
	p      plugingetter.CompatPlugin
}

type graphDriverRequest struct {
	ID         string `json:",omitempty"`
	Parent     string `json:",omitempty"`
	MountLabel string `json:",omitempty"`
}

type graphDriverResponse struct {
	Err      string            `json:",omitempty"`
	Dir      string            `json:",omitempty"`
	Exists   bool              `json:",omitempty"`
	Status   [][2]string       `json:",omitempty"`
	Changes  []archive.Change  `json:",omitempty"`
	Size     int64             `json:",omitempty"`
	Metadata map[string]string `json:",omitempty"`
}

type graphDriverInitRequest struct {
	Home string
	Opts []string
}

func (d *graphDriverProxy) Init(home string, opts []string) error {
	if !d.p.IsV1() {
		if cp, ok := d.p.(plugingetter.CountedPlugin); ok {
			// always acquire here, it will be cleaned up on daemon shutdown
			cp.Acquire()
		}
	}
	args := &graphDriverInitRequest{
		Home: home,
		Opts: opts,
	}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.Init", args, &ret); err != nil {
		return err
	}
	if ret.Err != "" {
		return errors.New(ret.Err)
	}
	return nil
}

func (d *graphDriverProxy) String() string {
	return d.name
}

func (d *graphDriverProxy) CreateReadWrite(id, parent string, opts *CreateOpts) error {
	mountLabel := ""
	if opts != nil {
		mountLabel = opts.MountLabel
	}

	args := &graphDriverRequest{
		ID:         id,
		Parent:     parent,
		MountLabel: mountLabel,
	}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.CreateReadWrite", args, &ret); err != nil {
		return err
	}
	if ret.Err != "" {
		return errors.New(ret.Err)
	}
	return nil
}

func (d *graphDriverProxy) Create(id, parent string, opts *CreateOpts) error {
	mountLabel := ""
	if opts != nil {
		mountLabel = opts.MountLabel
	}
	args := &graphDriverRequest{
		ID:         id,
		Parent:     parent,
		MountLabel: mountLabel,
	}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.Create", args, &ret); err != nil {
		return err
	}
	if ret.Err != "" {
		return errors.New(ret.Err)
	}
	return nil
}

func (d *graphDriverProxy) Remove(id string) error {
	args := &graphDriverRequest{ID: id}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.Remove", args, &ret); err != nil {
		return err
	}
	if ret.Err != "" {
		return errors.New(ret.Err)
	}
	return nil
}

func (d *graphDriverProxy) Get(id, mountLabel string) (string, error) {
	args := &graphDriverRequest{
		ID:         id,
		MountLabel: mountLabel,
	}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.Get", args, &ret); err != nil {
		return "", err
	}
	var err error
	if ret.Err != "" {
		err = errors.New(ret.Err)
	}
	return ret.Dir, err
}

func (d *graphDriverProxy) Put(id string) error {
	args := &graphDriverRequest{ID: id}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.Put", args, &ret); err != nil {
		return err
	}
	if ret.Err != "" {
		return errors.New(ret.Err)
	}
	return nil
}

func (d *graphDriverProxy) Exists(id string) bool {
	args := &graphDriverRequest{ID: id}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.Exists", args, &ret); err != nil {
		return false
	}
	return ret.Exists
}

func (d *graphDriverProxy) Status() [][2]string {
	args := &graphDriverRequest{}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.Status", args, &ret); err != nil {
		return nil
	}
	return ret.Status
}

func (d *graphDriverProxy) GetMetadata(id string) (map[string]string, error) {
	args := &graphDriverRequest{
		ID: id,
	}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.GetMetadata", args, &ret); err != nil {
		return nil, err
	}
	if ret.Err != "" {
		return nil, errors.New(ret.Err)
	}
	return ret.Metadata, nil
}

func (d *graphDriverProxy) Cleanup() error {
	if !d.p.IsV1() {
		if cp, ok := d.p.(plugingetter.CountedPlugin); ok {
			// always release
			defer cp.Release()
		}
	}

	args := &graphDriverRequest{}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.Cleanup", args, &ret); err != nil {
		return nil
	}
	if ret.Err != "" {
		return errors.New(ret.Err)
	}
	return nil
}

func (d *graphDriverProxy) Diff(id, parent string) (io.ReadCloser, error) {
	args := &graphDriverRequest{
		ID:     id,
		Parent: parent,
	}
	body, err := d.client.Stream("GraphDriver.Diff", args)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (d *graphDriverProxy) Changes(id, parent string) ([]archive.Change, error) {
	args := &graphDriverRequest{
		ID:     id,
		Parent: parent,
	}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.Changes", args, &ret); err != nil {
		return nil, err
	}
	if ret.Err != "" {
		return nil, errors.New(ret.Err)
	}

	return ret.Changes, nil
}

func (d *graphDriverProxy) ApplyDiff(id, parent string, diff io.Reader) (int64, error) {
	var ret graphDriverResponse
	if err := d.client.SendFile(fmt.Sprintf("GraphDriver.ApplyDiff?id=%s&parent=%s", id, parent), diff, &ret); err != nil {
		return -1, err
	}
	if ret.Err != "" {
		return -1, errors.New(ret.Err)
	}
	return ret.Size, nil
}

func (d *graphDriverProxy) DiffSize(id, parent string) (int64, error) {
	args := &graphDriverRequest{
		ID:     id,
		Parent: parent,
	}
	var ret graphDriverResponse
	if err := d.client.Call("GraphDriver.DiffSize", args, &ret); err != nil {
		return -1, err
	}
	if ret.Err != "" {
		return -1, errors.New(ret.Err)
	}
	return ret.Size, nil
}
