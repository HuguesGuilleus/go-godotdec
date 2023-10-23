package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
)

type Package struct {
	EngineVerion   uint32
	EngineMajor    uint32
	EngineMinor    uint32
	EngineRevision uint32

	Files []File
}

type File struct {
	Path   string
	Offset int64
	Size   int64
	MD5    [md5.Size]byte
	Data   []byte
}

func ReadPackage(r io.ReadSeeker) (*Package, error) {
	// Read meta
	headBytes := [4 + 4*4 + 16*4 + 4]byte{}
	if _, err := io.ReadFull(r, headBytes[:]); err != nil {
		return nil, fmt.Errorf("Can not read godot file head: %w", err)
	}

	if headBytes[0] != 0x47 || headBytes[1] != 0x44 || headBytes[2] != 0x50 || headBytes[3] != 0x43 {
		return nil, fmt.Errorf("Not a valid godot magic number, get 0x%X", headBytes[:4])
	}

	meta := &Package{}
	meta.EngineVerion = binary.LittleEndian.Uint32(headBytes[4:8])
	meta.EngineMajor = binary.LittleEndian.Uint32(headBytes[8:12])
	meta.EngineMinor = binary.LittleEndian.Uint32(headBytes[12:16])
	meta.EngineRevision = binary.LittleEndian.Uint32(headBytes[16:20])
	fileLen := int(binary.LittleEndian.Uint32(headBytes[21*4:]))

	// Read file infos
	meta.Files = make([]File, fileLen)
	for i := range meta.Files {
		info, err := readFileInfo(r)
		if err != nil {
			return nil, fmt.Errorf("Read file info %d: %w", i, err)
		}
		meta.Files[i] = *info
	}

	// Read file content
	if err := meta.readFiles(r); err != nil {
		return nil, err
	}

	// Return
	sort.Slice(meta.Files, func(i, j int) bool {
		return meta.Files[i].Path < meta.Files[j].Path
	})
	return meta, nil
}

func readFileInfo(r io.Reader) (*File, error) {
	pathLenBuffer := [4]byte{}
	if _, err := io.ReadFull(r, pathLenBuffer[:]); err != nil {
		return nil, fmt.Errorf("Read file path size: %w", err)
	}
	pathLen := binary.LittleEndian.Uint32(pathLenBuffer[:4])
	path := make([]byte, int(pathLen))
	if _, err := io.ReadFull(r, path); err != nil {
		return nil, fmt.Errorf("Read file path: %w", err)
	}
	path = bytes.TrimPrefix(path, []byte("res://"))
	path = bytes.TrimRightFunc(path, func(r rune) bool { return r == '\x00' })

	buff := [8*2 + md5.Size]byte{}
	if _, err := io.ReadFull(r, buff[:]); err != nil {
		return nil, fmt.Errorf("Read file info: %w", err)
	}

	info := &File{
		Path:   string(path),
		Offset: int64(binary.LittleEndian.Uint64(buff[0:8])),
		Size:   int64(binary.LittleEndian.Uint64(buff[8:16])),
		MD5:    [md5.Size]byte(buff[16:]),
	}

	return info, nil
}

func (meta *Package) readFiles(r io.ReadSeeker) error {
	sort.Slice(meta.Files, func(i, j int) bool {
		return meta.Files[i].Offset < meta.Files[j].Offset
	})

	for i, info := range meta.Files {
		if _, err := r.Seek(info.Offset, io.SeekStart); err != nil {
			return fmt.Errorf("Read file %q: %w", info.Path, err)
		}
		data := make([]byte, info.Size)
		if _, err := io.ReadFull(r, data); err != nil {
			return fmt.Errorf("Read file %q: %w", info.Path, err)
		} else if hash := md5.Sum(data); !bytes.Equal(hash[:], info.MD5[:]) {
			return fmt.Errorf("File %q has wrong hash 0x%X", info.Path, info.MD5)
		}
		meta.Files[i].Data = data
	}

	return nil
}
