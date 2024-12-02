// service.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/consul/api"
)

// Handler para la verificación de salud
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Healthy"))
}

// Función para registrar el servicio en Consul
func registerService() {
	client, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500", // Dirección de Consul en tu máquina local
	})
	if err != nil {
		log.Fatalf("Error creando cliente Consul: %v", err)
	}

	serviceRegistration := &api.AgentServiceRegistration{
		ID:      "my-service-1",
		Name:    "my-service",
		Address: "127.0.0.1",
		Port:    8080,
		Check: &api.AgentServiceCheck{
			HTTP:     "http://127.0.0.1:8080/health", // URL de verificación de salud
			Interval: "10s",                          // Intervalo de comprobación de salud
		},
	}

	// Registrar el servicio en Consul
	err = client.Agent().ServiceRegister(serviceRegistration)
	if err != nil {
		log.Fatalf("Error registrando el servicio: %v", err)
	}

	fmt.Println("Servicio registrado con éxito en Consul")
}

// Función para iniciar el servidor HTTP
func startHTTPServer() {
	http.HandleFunc("/health", healthCheckHandler)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}
