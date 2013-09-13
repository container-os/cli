package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type ChangeType int

const (
	ChangeModify = iota
	ChangeAdd
	ChangeDelete
)

type Change struct {
	Path string
	Kind ChangeType
}

func (change *Change) String() string {
	var kind string
	switch change.Kind {
	case ChangeModify:
		kind = "C"
	case ChangeAdd:
		kind = "A"
	case ChangeDelete:
		kind = "D"
	}
	return fmt.Sprintf("%s %s", kind, change.Path)
}

func ChangesAUFS(layers []string, rw string) ([]Change, error) {
	var changes []Change
	err := filepath.Walk(rw, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Rebase path
		path, err = filepath.Rel(rw, path)
		if err != nil {
			return err
		}
		path = filepath.Join("/", path)

		// Skip root
		if path == "/" {
			return nil
		}

		// Skip AUFS metadata
		if matched, err := filepath.Match("/.wh..wh.*", path); err != nil || matched {
			return err
		}

		change := Change{
			Path: path,
		}

		// Find out what kind of modification happened
		file := filepath.Base(path)
		// If there is a whiteout, then the file was removed
		if strings.HasPrefix(file, ".wh.") {
			originalFile := file[len(".wh."):]
			change.Path = filepath.Join(filepath.Dir(path), originalFile)
			change.Kind = ChangeDelete
		} else {
			// Otherwise, the file was added
			change.Kind = ChangeAdd

			// ...Unless it already existed in a top layer, in which case, it's a modification
			for _, layer := range layers {
				stat, err := os.Stat(filepath.Join(layer, path))
				if err != nil && !os.IsNotExist(err) {
					return err
				}
				if err == nil {
					// The file existed in the top layer, so that's a modification

					// However, if it's a directory, maybe it wasn't actually modified.
					// If you modify /foo/bar/baz, then /foo will be part of the changed files only because it's the parent of bar
					if stat.IsDir() && f.IsDir() {
						if f.Size() == stat.Size() && f.Mode() == stat.Mode() && f.ModTime() == stat.ModTime() {
							// Both directories are the same, don't record the change
							return nil
						}
					}
					change.Kind = ChangeModify
					break
				}
			}
		}

		// Record change
		changes = append(changes, change)
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return changes, nil
}

type FileInfo struct {
	parent *FileInfo
	name string
	stat syscall.Stat_t
	children map[string]*FileInfo
}

func (root *FileInfo) LookUp(path string) *FileInfo {
	parent := root
	if path == "/" {
		return root
	}

	pathElements := strings.Split(path, "/")
	for _, elem := range pathElements {
		if elem != "" {
			child := parent.children[elem]
			if child == nil {
				return nil
			}
			parent = child
		}
	}
	return parent
}

func (info *FileInfo)path() string {
	if info.parent == nil {
		return "/"
	}
	return filepath.Join(info.parent.path(), info.name)
}

func (info *FileInfo)unlink() {
	if info.parent != nil {
		delete(info.parent.children, info.name)
	}
}

func (info *FileInfo)Remove(path string) bool {
	child := info.LookUp(path)
	if child != nil {
		child.unlink()
		return true
	}
	return false
}

func (info *FileInfo)isDir() bool {
	return info.parent == nil || info.stat.Mode&syscall.S_IFDIR == syscall.S_IFDIR
}


func (info *FileInfo)addChanges(oldInfo *FileInfo, changes *[]Change) {
	if oldInfo == nil {
		// add
		change := Change{
			Path: info.path(),
			Kind: ChangeAdd,
		}
		*changes = append(*changes, change)
	}

	// We make a copy so we can modify it to detect additions
	// also, we only recurse on the old dir if the new info is a directory
	// otherwise any previous delete/change is considered recursive
	oldChildren := make(map[string]*FileInfo)
	if oldInfo != nil && info.isDir() {
		for k, v := range oldInfo.children {
			oldChildren[k] = v
		}
	}

	for name, newChild := range info.children {
		oldChild, _ := oldChildren[name]
		if oldChild != nil {
			// change?
			oldStat := &oldChild.stat
			newStat := &newChild.stat
			if oldStat.Ino != newStat.Ino ||
				oldStat.Mode != newStat.Mode ||
				oldStat.Uid != newStat.Uid ||
				oldStat.Gid != newStat.Gid ||
				oldStat.Rdev != newStat.Rdev ||
				oldStat.Size != newStat.Size ||
				oldStat.Blocks != newStat.Blocks ||
				oldStat.Mtim != newStat.Mtim ||
				oldStat.Ctim != newStat.Ctim {
				change := Change{
					Path: newChild.path(),
					Kind: ChangeModify,
				}
				*changes = append(*changes, change)
			}

			// Remove from copy so we can detect deletions
			delete(oldChildren, name)
		}

		newChild.addChanges(oldChild, changes)
	}
	for _, oldChild := range oldChildren {
		// delete
		change := Change{
			Path: oldChild.path(),
			Kind: ChangeDelete,
		}
		*changes = append(*changes, change)
	}


}

func (info *FileInfo)Changes(oldInfo *FileInfo) []Change {
	var changes []Change

	info.addChanges(oldInfo, &changes)

	return changes
}


func collectFileInfo(sourceDir string) (*FileInfo, error) {
	root := &FileInfo {
		name: "/",
		children: make(map[string]*FileInfo),
	}

	err := filepath.Walk(sourceDir, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Rebase path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.Join("/", relPath)

		if relPath == "/" {
			return nil
		}

		parent := root.LookUp(filepath.Dir(relPath))
		if parent == nil {
			return fmt.Errorf("collectFileInfo: Unexpectedly no parent for %s", relPath)
		}

		info := &FileInfo {
			name: filepath.Base(relPath),
			children: make(map[string]*FileInfo),
			parent: parent,
		}

		if err := syscall.Lstat(path, &info.stat); err != nil {
			return err
		}

		parent.children[info.name] = info

		return nil
	})
	if err != nil {
		return nil, err
	}
	return root, nil
}

func ChangesDirs(newDir, oldDir string) ([]Change, error) {
	oldRoot, err := collectFileInfo(oldDir)
	if err != nil {
		return nil, err
	}
	newRoot, err := collectFileInfo(newDir)
	if err != nil {
		return nil, err
	}

	// Ignore changes in .docker-id
	_ = newRoot.Remove("/.docker-id")
	_ = oldRoot.Remove("/.docker-id")

	return newRoot.Changes(oldRoot), nil
}
