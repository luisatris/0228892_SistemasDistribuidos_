// main.go
package main

import (
	"time"
)

func main() {
	// Iniciar servidor HTTP en un goroutine
	go startHTTPServer()

	// Esperar un momento para asegurar que el servidor HTTP esté listo
	time.Sleep(2 * time.Second)

	// Registrar el servicio en Consul
	registerService()

	// Iniciar el descubrimiento de servicios
	// Esto invoca automáticamente `startDiscovery()` desde client.go
	startDiscovery()

	// Mantener el programa en ejecución
	select {}
}
