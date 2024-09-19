package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
)

type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

type Log struct {
	mu      sync.Mutex
	records []Record
}

var logfile = Log{records: []Record{}}
var offsetCounter uint64 = 0

// Este handler maneja las solicitudes de escritura y lectura del Log.
func handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/write":
		var record Record

		err := json.NewDecoder(r.Body).Decode(&record)
		if err != nil {
			http.Error(w, "Error deserializando el JSON", http.StatusBadRequest)
			return
		}

		record.Offset = offsetCounter
		offsetCounter++

		logfile.mu.Lock()
		logfile.records = append(logfile.records, record)
		logfile.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(record)

	case "/read":
		var requestData struct {
			Offset *uint64 `json:"offset"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Error deserializando el JSON", http.StatusBadRequest)
			return
		}

		logfile.mu.Lock()
		defer logfile.mu.Unlock()

		for _, record := range logfile.records {
			if record.Offset == *requestData.Offset {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]Record{"record": record})
				return
			}
		}

		http.Error(w, "Registro no encontrado", http.StatusNotFound)

	default:
		http.Error(w, "Ruta no encontrada", http.StatusNotFound)
	}
}

func main() {
	logDir := "log_directory"
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatalf("Error al crear el directorio de logs: %v", err)
	}

	logFile, err := os.OpenFile(logDir+"/logfile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Error al abrir el archivo de log: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	http.HandleFunc("/", handler)

	log.Println("Iniciando el servidor en :8081...")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Fallo al iniciar el servidor: %v", err)
	}
}
