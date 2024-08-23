package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Record struct { //aqui se establece la estructura del record
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

type Log struct { //Aqui se establece la estructura de nuestro log
	mu      sync.Mutex
	records []Record
}

var logfile = Log{records: []Record{}}
var offsetCounter uint64 = 0

//Este handler nos permite manejar las solicitudes de escritura y lectura del Log.
func handler(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/write":
		// En este caso escribimos en el Log
		var record Record

		err := json.NewDecoder(r.Body).Decode(&record)
		if err != nil {
			http.Error(w, "Error deserializando el JSON", http.StatusBadRequest)//nos indica si hay algun error al deserializar el JSON
			return
		}

		//aqui se maneja el offset
		record.Offset = offsetCounter
		offsetCounter++

		logfile.mu.Lock()
		logfile.records = append(logfile.records, record) //se agrega el nuevo record
		logfile.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(record)

	case "/read":
		// este caso se utiliza para leer el Log
		var requestData struct {
			Offset *uint64 `json:"offset"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestData)
		if err != nil {
			http.Error(w, "Error deserializando el JSON", http.StatusBadRequest)//nos indica si hay error
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

		http.Error(w, "Registro no encontrado", http.StatusNotFound)//nos indica si el registro no se ha encontrado

	default:
		http.Error(w, "Ruta no encontrada", http.StatusNotFound)//nos indica si la ruta no dfue encontrada
	}
}}



func main() {
	http.HandleFunc("/", handler)

	log.Println("Iniciando el servidor en :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Fallo al iniciar el servidor: %v", err)
	}
}
