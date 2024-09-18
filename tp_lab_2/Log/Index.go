package Log

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

func NewIndex(pf *os.File, c Config) (*Index, error) {
	// Usar el archivo recibido directamente sin abrirlo de nuevo
	data, err := mmap.Map(pf, mmap.RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("error mapping file: %v", err)
	}

	return &Index{
		file: pf,
		mmap: data,
		size: c.IndexWidth, // Suponiendo que 'Width' es un campo de la estructura 'Config'
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

func (i *Index) Write(off uint32, pos uint64) error {
	if i.size == 0 {
		return fmt.Errorf("index not initialized")
	}

	// Calcular la posición del índice basado en 'off'
	idxPos := uint64(off) * EntryWidth
	if idxPos+EntryWidth > uint64(i.size) {
		return fmt.Errorf("index out of range")
	}

	// Crear los bytes para 'off' y 'pos'
	offBytes := make([]byte, OffWidth)
	posBytes := make([]byte, PossWidth)

	binary.LittleEndian.PutUint32(offBytes, off)
	binary.LittleEndian.PutUint64(posBytes, pos)

	// Copiar los bytes en la memoria mapeada
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

func (i *Index) Name() string {
	return i.file.Name()
}
