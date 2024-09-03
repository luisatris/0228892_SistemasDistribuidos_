package Index

// SegmentConfig contiene configuraciones específicas para los segmentos.
type SegmentConfig struct {
	MaxStoreBytes uint64 // Tamaño máximo del archivo de almacenamiento (store)
	MaxIndexBytes uint64 // Tamaño máximo del archivo de índice (index)
	InitialOffset uint64 // Offset inicial para nuevos segmentos
	IndexWidth    uint64
}

// Config es una configuración para el índice.
type Config struct {
	Segment    SegmentConfig // Configuración de los segmentos
	IndexWidth uint64
}
