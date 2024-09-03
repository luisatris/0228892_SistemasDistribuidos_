package Index

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

// Aqui en el newstore, Abrimos un archivo (si no existe, se crea) y se inizializa el buffer
func NewStore(filePath string) (*store, error) {
	// Abre el archivo o lo crea si no existe
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	// Obtén el tamaño actual del archivo
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	return &store{
		File: f,
		size: uint64(stat.Size()),
		buf:  bufio.NewWriter(f),
	}, nil
}

// usamos sync.Mutex para bloquear accesos no permitidos. Luego escribe el tamaño de los datos antes de escribir los datos realies. Una vez que fue comprobado, escibimos los datos del archivo y se actualiza el tamaño de los archivos
func (s *store) Append(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Escribe el tamaño de los datos antes de escribir los datos reales
	if err := binary.Write(s.buf, enc, uint64(len(data))); err != nil {
		return err
	}

	// Escribe los datos
	if _, err := s.buf.Write(data); err != nil {
		return err
	}

	// Asegúrate de que todos los datos se escriban en el archivo
	if err := s.buf.Flush(); err != nil {
		return err
	}

	// Actualiza el tamaño del archivo
	s.size += uint64(len(data)) + lenWidth

	return nil
}

// inicia al pricnipio del archiuvo y lee los datos usando el tamaño previamente establecido. Va acumulando los datos leidos en un solo slice de bytes
func (s *store) Read() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Regresa al principio del archivo
	if _, err := s.File.Seek(0, 0); err != nil {
		return nil, err
	}

	var data []byte

	for {
		var length uint64
		if err := binary.Read(s.File, enc, &length); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		buf := make([]byte, length)
		if _, err := s.File.Read(buf); err != nil {
			return nil, err
		}

		data = append(data, buf...)
	}

	return data, nil
}

// con el read at, podemos leer datos desde una posición especifica en el archivo. Se alamcenan en un buffer especifico
func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil { //se asegura de que no hayan datos por ser escritos mientras se hace la lectura
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

// Close cierra el archivo y libera los recursos
func (s *store) Close() error {
	if err := s.buf.Flush(); err != nil {
		return err
	}
	return s.File.Close()
}
