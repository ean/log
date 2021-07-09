// Package mmap implements functionality for keeping a files data in memory.
// It is used by log control to have direct access to the content of the control file.
package mmap

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

type MMap struct {
	f     *os.File
	Data  []byte
	prot  int
	flags int
}

const (
	PROT_READ  = syscall.PROT_READ
	PROT_WRITE = syscall.PROT_WRITE

	MAP_PRIVATE = unix.MAP_PRIVATE
	MAP_SHARED  = unix.MAP_SHARED
)

// Map creates a new memory mapping for the given file handle
func Map(path string, prot int, flags int) (*MMap, error) {
	fileFlags := os.O_RDONLY | os.O_CREATE
	if prot&PROT_WRITE != 0 {
		fileFlags = os.O_CREATE | os.O_RDWR
	}
	f, err := os.OpenFile(path, fileFlags, 0600)
	if err != nil {
		return nil, err
	}
	m := &MMap{
		f:     f,
		Data:  nil,
		prot:  prot,
		flags: flags,
	}
	err = m.Extend()
	return m, err
}

// Flush ensures written data is synced to permanent storage
func (m *MMap) Flush() error {
	if m != nil {
		return unix.Msync(m.Data, unix.MS_SYNC)
	}
	return nil
}

// Unmap removes the current mapping for m
func (m *MMap) Unmap() error {
	if m != nil {
		m.f.Close()
		return syscall.Munmap(m.Data)
	}
	return nil
}

// Extend remaps m to the current size of the underlying file handle
func (m *MMap) Extend() error {
	if m.Data != nil {
		if err := syscall.Munmap(m.Data); err != nil {
			return err
		}
	}
	s, err := m.f.Stat()
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	l := int(s.Size())
	if l == 0 {
		m.Data = nil
		return nil
	}

	data, err := syscall.Mmap(int(m.f.Fd()), 0, l, m.prot, m.flags)
	if err != nil {
		return fmt.Errorf("mmap: %w", err)
	}
	m.Data = data
	return nil
}
