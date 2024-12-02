// client.go
package main

import (
	"fmt"
	"log"

	"github.com/hashicorp/consul/api"
)

// Función para descubrir servicios desde Consul
func discoverService(serviceName string) {
	// Crear cliente Consul y especificar la dirección y el puerto de Consul
	client, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500", // Dirección de Consul en tu máquina local
	})
	if err != nil {
		log.Fatalf("Error creando cliente Consul: %v", err)
	}

	// Buscar el servicio registrado en Consul
	services, _, err := client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		log.Fatalf("Error buscando el servicio: %v", err)
	}

	if len(services) == 0 {
		fmt.Printf("No se encontraron servicios con el nombre: %s\n", serviceName)
	} else {
		for _, service := range services {
			fmt.Printf("Servicio encontrado: %s, Dirección: %s:%d\n",
				service.Service.ID, service.Service.Address, service.Service.Port)
		}
	}
}

// Llama a la función para descubrir el servicio
func startDiscovery() {
	// Descubrir el servicio 'my-service' registrado en Consul
	discoverService("my-service")
}
