/*
 * Copyright © Siemens 2026 - 2026. ALL RIGHTS RESERVED.
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package utils

import (
	"io"
	"os"
)

// FileSystemOperations is an interface to bring functionality of required file system operations
// Note: missing functionalities can be added and implemented on demand, functions must be one-to-one mapping from `os` package
type FileSystemOperations interface {

	// Open os.WriteFile
	WriteFile(path string, data []byte, perm os.FileMode) error
	// Open os.RemoveAll
	RemoveAll(path string) error

	// Open os.Chmod
	Chmod(name string, mode os.FileMode) error

	// Open os.Chown
	Chown(name string, uid, gid int) error

	// Open os.Stat
	Stat(path string) error

	// Open os.MkdirAll
	MkdirAll(path string, perm os.FileMode) error
}

// FileIO is an interface to bring functionality of required file IO operations
// Note: missing functionalities can be added and implemented on demand
type FileIO interface {
	io.Reader
	io.Writer
	io.Closer

	Sync() error
}

// OsFileSystemOperations FileSystemOperations implementation, wrapper for functionalities from `os` package
type OsFileSystemOperations struct{}

func (f *OsFileSystemOperations) WriteFile(path string, data []byte, perm os.FileMode) error {
	if err := os.WriteFile(path, data, perm); err != nil {
		return err
	}
	return nil
}

func (f *OsFileSystemOperations) ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}
	return data, nil
}

func (f *OsFileSystemOperations) Chown(name string, uid, gid int) error {
	if err := os.Chown(name, uid, gid); err != nil {
		return err
	}
	return nil
}

func (f *OsFileSystemOperations) Chmod(name string, mode os.FileMode) error {
	if err := os.Chmod(name, mode); err != nil {
		return err
	}
	return nil
}

func (f *OsFileSystemOperations) RemoveAll(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

func (f *OsFileSystemOperations) Stat(path string) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}

func (f *OsFileSystemOperations) MkdirAll(path string, perm os.FileMode) error {
	if err := os.MkdirAll(path, perm); err != nil {
		return err
	}
	return nil
}

// OsFile wrapper type for os.File
type OsFile struct {
	file *os.File
}

func (o *OsFile) Read(p []byte) (n int, err error) {
	return o.file.Read(p)
}

func (o *OsFile) Write(p []byte) (n int, err error) {
	return o.file.Write(p)
}

func (o *OsFile) Close() error {
	return o.file.Close()
}

func (o *OsFile) Sync() error {
	return o.file.Sync()
}
