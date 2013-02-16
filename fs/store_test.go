package fs

import (
	"testing"
	"io/ioutil"
	"github.com/dotcloud/docker/fake"
	"os"
	"errors"
	"fmt"
)

func TestInit(t *testing.T) {
	store, err := TempStore("testinit")
	if err != nil {
		t.Fatal(err)
	}
	defer nuke(store)
	paths, err := store.Paths()
	if err != nil {
		t.Fatal(err)
	}
	if l := len(paths); l != 0 {
		t.Fatal("Fresh store should be empty after init (len=%d)", l)
	}
}

func TestCreate(t *testing.T) {
	store, err := TempStore("testcreate")
	if err != nil {
		t.Fatal(err)
	}
	defer nuke(store)
	archive, err := fake.FakeTar()
	if err != nil {
		t.Fatal(err)
	}
	image, err := store.Create(archive, nil, "foo", "Testing")
	if err != nil {
		t.Fatal(err)
	}
	if images, err := store.Images(); err != nil {
		t.Fatal(err)
	} else if l := len(images); l != 1 {
		t.Fatalf("Wrong number of images. Should be %d, not %d", 1, l)
	}
	if images, err := store.List("foo"); err != nil {
		t.Fatal(err)
	} else if l := len(images); l != 1 {
		t.Fatalf("Path foo has wrong number of images (should be %d, not %d)", 1, l)
	} else if images[0].Id != image.Id {
		t.Fatalf("Imported image should be listed at path foo (%s != %s)", images[0], image)
	}
}

// Copy an image to a new path
func TestCopyNewPath(t *testing.T) {
	store, err := TempStore("testcopynewpath")
	if err != nil {
		t.Fatal(err)
	}
	defer nuke(store)
	archive, err := fake.FakeTar()
	if err != nil {
		t.Fatal(err)
	}
	src, err := store.Create(archive, nil, "foo", "Testing")
	if err != nil {
		t.Fatal(err)
	}
	dst, err := src.Copy("bar")
	if err != nil {
		t.Fatal(err)
	}
	// ID should be the same
	if src.Id != dst.Id {
		t.Fatal("Different IDs")
	}
	// Check number of images at source path
	if images, err := store.List("foo"); err != nil {
		t.Fatal(err)
	} else if l := len(images); l != 1 {
		t.Fatal("Wrong number of images at source path (should be %d, not %d)", 1, l)
	}
	// Check number of images at destination path
	if images, err := store.List("bar"); err != nil {
		t.Fatal(err)
	} else if l := len(images); l != 1 {
		t.Fatal("Wrong number of images at destination path (should be %d, not %d)", 1, l)
	}
	if err := healthCheck(store); err != nil {
		t.Fatal(err)
	}
}

// Copying an image to the same path twice should fail
func TestCopySameName(t *testing.T) {
	store, err := TempStore("testcopysamename")
	if err != nil {
		t.Fatal(err)
	}
	defer nuke(store)
	archive, err := fake.FakeTar()
	if err != nil {
		t.Fatal(err)
	}
	src, err := store.Create(archive, nil, "foo", "Testing")
	if err != nil {
		t.Fatal(err)
	}
	_, err = src.Copy("foo")
	if err == nil {
		t.Fatal("Copying an image to the same patch twice should fail.")
	}
}

/*
func TestMount(t *testing.T) {
	store, err := TempStore()
	if err != nil {
		t.Fatal(err)
	}
	defer nuke(store)
	archive, err := fake.FakeTar()
	if err != nil {
		t.Fatal(err)
	}
	image, err := store.Create(archive, nil, "foo", "Testing")
	if err != nil {
		t.Fatal(err)
	}
	// Create mount targets
	root, err := ioutil.TempDir("", "docker-fs-test")
	if err != nil {
		t.Fatal(err)
	}
	rw, err := ioutil.TempDir("", "docker-fs-test")
	if err != nil {
		t.Fatal(err)
	}
	mountpoint, err := image.Mount(root, rw)
	if err != nil {
		t.Fatal(err)
	}
	defer mountpoint.Umount()
	// Mountpoint should be marked as mounted
	if !mountpoint.Mounted() {
		t.Fatal("Mountpoint not mounted")
	}
	// There should be one mountpoint registered
	if l := len(image.Mountpoints()); l != 1 {
		t.Fatal("Wrong number of mountpoints registered (should be %d, not %d)", 1, l)
	}
	// Unmounting should work
	if err := mountpoint.Umount(); err != nil {
		t.Fatal(err)
	}
	// De-registering should work
	if err := mountpoint.Deregister(); err != nil {
		t.Fatal(err)
	}
	if l := len(image.Mountpoints()); l != 0 {
		t.Fatal("Wrong number of mountpoints registered (should be %d, not %d)", 0, l)
	}
	// General health check
	if err := healthCheck(); err != nil {
		t.Fatal(err)
	}
}
*/

func TempStore(prefix string) (*Store, error) {
	dir, err := ioutil.TempDir("", "docker-fs-test-" + prefix)
	if err != nil {
		return nil, err
	}
	return New(dir)
}

func nuke(store *Store) error {
	return os.RemoveAll(store.Root)
}

// Look for inconsistencies in a store.
func healthCheck(store *Store) error {
	parents := make(map[string]bool)
	paths, err := store.Paths()
	if err != nil {
		return err
	}
	for _, path := range paths {
		images, err := store.List(path)
		if err != nil {
			return err
		}
		IDs := make(map[string]bool) // All IDs for this path
		for _, img := range images {
			// Check for duplicate IDs per path
			if _, exists := IDs[img.Id]; exists {
				return errors.New(fmt.Sprintf("Duplicate ID: %s", img.Id))
			} else {
				IDs[img.Id] = true
			}
			// Store parent for 2nd pass
			if parent := img.Parent; parent != "" {
				parents[parent] = true
			}
		}
	}
	// Check non-existing parents
	for parent := range parents {
		if _, exists := parents[parent]; !exists {
			return errors.New("Reference to non-registered parent: " + parent)
		}
	}
	return nil
}

