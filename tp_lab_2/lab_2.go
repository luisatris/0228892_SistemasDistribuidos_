package Index

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

type grpcServer struct {
	api.UnimplementedLogServer
	*CommitLog
}

type CommitLog struct {
	mu         sync.Mutex
	records    map[uint64]Record // Usamos un mapa para acceso rápido por offset
	nextOffset uint64
}

/*
type Log struct {
	mu      sync.Mutex
	records []Record
}
*/

var logfile = Log{records: []Record{}}
var offsetCounter uint64 = 0

var _ api.LogServer = (*grpcServer)(nil)

func writeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

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
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

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
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Ruta no encontrada", http.StatusNotFound)
}

func newgrpcServer(commitlog *CommitLog) (srv *grpcServer, err error) {
	srv = &grpcServer{
		CommitLog: commitlog,
	}
	return srv, nil
}

func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (*api.ProduceResponse, error) {
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.ProduceResponse{Offset: offset}, nil
}

func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (*api.ConsumeResponse, error) {
	// Leemos el registro desde el CommitLog basado en el offset proporcionado en la solicitud.
	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, err
	}

	// Creamos la respuesta utilizando el registro obtenido.
	return &api.ConsumeResponse{
		Record: record, // Asumimos que `ConsumeResponse` tiene un campo `Record` de tipo `*Record`
	}, nil
}

func main() {
	http.HandleFunc("/write", writeHandler)
	http.HandleFunc("/read", readHandler)
	http.HandleFunc("/", notFoundHandler)

	log.Println("Iniciando el servidor en :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Fallo al iniciar el servidor: %v", err)
	}
}
