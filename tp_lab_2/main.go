package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	api "example.com/tpmod/Api/v1" // Asegúrate de tener la ruta correcta
	logpkg "example.com/tpmod/log" // Cambia a tu paquete real
)

type LogHandler struct {
	logfile *logpkg.Log
	mu      sync.Mutex // Mutex para proteger el acceso concurrente
}

func (lh *LogHandler) writeHandler(w http.ResponseWriter, r *http.Request) {
	lh.mu.Lock()         // Bloquea el mutex antes de escribir
	defer lh.mu.Unlock() // Asegúrate de desbloquear al final

	var record api.Record
	err := json.NewDecoder(r.Body).Decode(&record)
	if err != nil {
		http.Error(w, "Error deserializando el JSON", http.StatusBadRequest)
		return
	}

	// Usar el método Append del log
	offset, err := lh.logfile.Append(&record)
	if err != nil {
		http.Error(w, "Error escribiendo en el log", http.StatusInternalServerError)
		return
	}

	record.Offset = offset
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

func (lh *LogHandler) readHandler(w http.ResponseWriter, r *http.Request) {
	lh.mu.Lock()         // Bloquea el mutex antes de leer
	defer lh.mu.Unlock() // Asegúrate de desbloquear al final

	var requestData struct {
		Offset uint64 `json:"offset"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Error deserializando el JSON", http.StatusBadRequest)
		return
	}

	record, err := lh.logfile.Read(requestData.Offset)
	if err != nil {
		http.Error(w, "Registro no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]api.Record{"record": *record})
}

func main() {
	// Crear una nueva instancia de Log
	logConfig := logpkg.Config{
		Segment: logpkg.SegmentConfig{
			MaxStoreBytes: 1024, // Configura según tus necesidades
			MaxIndexBytes: 1024,
			InitialOffset: 0,
		},
	}

	logfile, err := logpkg.NewLog("log_directory", logConfig)
	if err != nil {
		log.Fatalf("Error al crear el log: %v", err)
	}
	defer logfile.Close()

	logHandler := &LogHandler{logfile: logfile}

	http.HandleFunc("/write", logHandler.writeHandler)
	http.HandleFunc("/read", logHandler.readHandler)

	log.Println("Iniciando el servidor en :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Fallo al iniciar el servidor: %v", err)
	}
}
