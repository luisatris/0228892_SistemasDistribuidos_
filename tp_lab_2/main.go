package main

import (
	"log"

	//api "example.com/tpmod/Api/v1"
	logpkg "example.com/tpmod/Log"
	Server "example.com/tpmod/server"
)

func main() {
	logConfig := logpkg.Config{
		Segment: logpkg.SegmentConfig{
			MaxStoreBytes: 1024,
			MaxIndexBytes: 1024,
			InitialOffset: 0,
		},
	}

	logfile, err := logpkg.NewLog("log_directory", logConfig)
	if err != nil {
		log.Fatalf("Error al crear el log: %v", err)
	}
	defer logfile.Close()

	// Crear el servidor HTTP
	httpServer := Server.NewHTTPServer(":8080", logfile)

	log.Println("Iniciando el servidor en :8080...")
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatalf("Fallo al iniciar el servidor: %v", err)
	}
}
