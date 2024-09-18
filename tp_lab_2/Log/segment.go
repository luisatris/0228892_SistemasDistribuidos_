package Log

import (
	"fmt"
	"os"
	"path"
	"strings"

	api "example.com/tpmod/api/v1"
)

// segment maneja un segmento de almacenamiento y su índice.
type segment struct {
	store                  *store
	index                  *Index
	baseOffset, nextOffset uint64
	config                 Config
}

// newSegment crea un nuevo segmento.
func newSegment(dir string, baseOffset uint64, c Config) (*segment, error) {
	// Inicializar el segmento
	s := &segment{
		baseOffset: baseOffset,
		config:     c,
	}

	// Abrir el archivo de almacenamiento
	storeFilePath := path.Join(dir, fmt.Sprintf("%d.store", baseOffset))
	storeFile, err := os.OpenFile(storeFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Crear una nueva instancia de store usando el archivo abierto
	if s.store, err = NewStore(storeFile); err != nil { // Cambiado a storeFile
		storeFile.Close() // Cerrar el archivo en caso de error
		return nil, err
	}

	// Abrir el archivo del índice
	indexFilePath := path.Join(dir, fmt.Sprintf("%d.index", baseOffset))
	indexFile, err := os.OpenFile(indexFilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		s.store.Close() // Cerrar el archivo store en caso de error
		return nil, err
	}

	// Crear una nueva instancia de index usando el archivo abierto
	if s.index, err = NewIndex(indexFile, c); err != nil { // Cambiado para pasar c
		indexFile.Close() // Cerrar el archivo en caso de error
		s.store.Close()   // Cerrar el archivo store en caso de error
		return nil, err
	}

	// Leer el índice para determinar el siguiente offset
	off, _, err := s.index.Read(-1)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.nextOffset = baseOffset
		} else {
			return nil, err
		}
	} else {
		s.nextOffset = baseOffset + uint64(off) + 1
	}

	return s, nil
}

// IsMaxed verifica si el segmento ha alcanzado la capacidad máxima.
func (s *segment) IsMaxed() bool {
	// Define el tamaño máximo en bytes
	const maxSize = 1 << 20 // 1 MB

	// Verifica si el tamaño del archivo ha alcanzado o superado la capacidad máxima
	stat, err := s.store.Stat()
	if err != nil {
		return false
	}
	return stat.Size() >= maxSize
}

func (s *segment) Append(record *api.Record) (uint64, error) {
	// Agregar datos al store y capturar los valores devueltos
	off, n, err := s.store.Append(record.Data) // Captura los tres valores
	if err != nil {
		return 0, err
	}

	// Calcular el nuevo offset
	dataSize := n // Usa n para el tamaño de los datos escritos
	newOffset := off

	// Actualizar el índice
	if err := s.index.Write(uint32(newOffset-s.baseOffset), newOffset); err != nil { // Cambiado para pasar solo dos argumentos
		return 0, err
	}

	// Actualizar el próximo offset
	s.nextOffset += uint64(dataSize) + 1
	return newOffset, nil
}

func (s *segment) Read(offset uint64) (*api.Record, error) {
	// Leer del índice para obtener la posición de los datos
	_, pos, err := s.index.Read(int64(offset - s.baseOffset))
	if err != nil {
		return nil, err
	}

	// Leer los datos desde la posición obtenida
	data := make([]byte, pos) // Aquí deberías usar el tamaño correcto
	if _, err := s.store.ReadAt(data, int64(pos)); err != nil {
		return nil, err
	}

	return &api.Record{Data: data}, nil
}

// Remove elimina los archivos del segmento.
func (s *segment) Remove() error {
	// Construir las rutas de los archivos
	storeFilePath := path.Join("segment_dir", fmt.Sprintf("%d.store", s.baseOffset))
	indexFilePath := path.Join("segment_dir", fmt.Sprintf("%d.index", s.baseOffset))

	// Eliminar el archivo de datos (store)
	if err := os.Remove(storeFilePath); err != nil {
		return fmt.Errorf("error al eliminar archivo de datos: %v", err)
	}

	// Eliminar el archivo de índice (index)
	if err := os.Remove(indexFilePath); err != nil {
		return fmt.Errorf("error al eliminar archivo de índice: %v", err)
	}

	return nil
}

// Close cierra el segmento y libera los recursos asociados.
func (s *segment) Close() error {
	var err error

	// Cerrar el archivo de store
	if s.store != nil {
		if closeErr := s.store.Close(); closeErr != nil {
			err = closeErr
		}
	}

	// Cerrar el archivo de índice
	if s.index != nil {
		if closeErr := s.index.Close(); closeErr != nil {
			err = closeErr
		}
	}

	return err
}
