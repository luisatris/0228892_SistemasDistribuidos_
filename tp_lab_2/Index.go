package Index

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/edsrzf/mmap-go"
)

const (
	OffWidth   uint64 = 10
	PossWidth  uint64 = 20
	EntryWidth uint64 = OffWidth + PossWidth
)

type Index struct {
	file *os.File
	mmap mmap.MMap
	size uint64
}

func NewIndex(filePath string, width uint64) (*Index, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}

	data, err := mmap.Map(file, mmap.RDWR, 0)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("error mapping file: %v", err)
	}

	return &Index{
		file: file,
		mmap: data,
		size: width,
	}, nil
}

func (i *Index) Read(in int64) (uint32, uint64, error) {
	if i.size == 0 {
		return 0, 0, fmt.Errorf("index not initialized")
	}

	if in == -1 {
		return 0, 0, fmt.Errorf("index out of range")
	}

	pos := uint64(in) * EntryWidth
	if i.size < pos+EntryWidth {
		return 0, 0, fmt.Errorf("index out of range")
	}

	offBytes := i.mmap[pos : pos+OffWidth]
	posBytes := i.mmap[pos+OffWidth : pos+EntryWidth]

	off := binary.LittleEndian.Uint32(offBytes)
	p := binary.LittleEndian.Uint64(posBytes)

	return off, p, nil
}

func (i *Index) Write(in int64, off uint32, pos uint64) error {
	if i.size == 0 {
		return fmt.Errorf("index not initialized")
	}

	if in == -1 {
		return fmt.Errorf("index out of range")
	}

	idxPos := uint64(in) * EntryWidth
	if i.size < idxPos+EntryWidth {
		return fmt.Errorf("index out of range")
	}

	offBytes := make([]byte, OffWidth)
	posBytes := make([]byte, PossWidth)

	binary.LittleEndian.PutUint32(offBytes, off)
	binary.LittleEndian.PutUint64(posBytes, pos)

	copy(i.mmap[idxPos:idxPos+OffWidth], offBytes)
	copy(i.mmap[idxPos+OffWidth:idxPos+EntryWidth], posBytes)

	return nil
}

func (i *Index) Close() error {
	if err := i.mmap.Unmap(); err != nil {
		return fmt.Errorf("error unmapping file: %v", err)
	}
	return i.file.Close()
}
