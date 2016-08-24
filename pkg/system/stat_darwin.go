package system

import (
	"syscall"
)

// fromStatT creates a system.StatT type from a syscall.Stat_t type
func fromStatT(s *syscall.Stat_t) (*StatT, error) {
	return &StatT{size: s.Size,
		mode: uint32(s.Mode),
		uid:  s.Uid,
		gid:  s.Gid,
		rdev: uint64(s.Rdev),
		mtim: s.Mtimespec}, nil
}

// FromStatT loads a system.StatT from a syscall.Stat_t.
func FromStatT(s *syscall.Stat_t) (*StatT, error) {
	return fromStatT(s)
}

// Stat takes a path to a file and returns
// a system.StatT type pertaining to that file.
//
// Throws an error if the file does not exist
func Stat(path string) (*StatT, error) {
	s := &syscall.Stat_t{}
	if err := syscall.Stat(path, s); err != nil {
		return nil, err
	}
	return fromStatT(s)
}
