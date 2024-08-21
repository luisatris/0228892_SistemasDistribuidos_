package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type User struct { //estructura de el tipo de objeto User
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Age       int    `json:"age"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { //genera un mensaje de bienvenida en cuanto alguien ingresa al sitio web
		fmt.Fprintf(w, "Welcome to my website!")
	})

	http.HandleFunc("/encode", func(w http.ResponseWriter, r *http.Request) {
		peter := User{
			Firstname: "John",
			Lastname:  "Doe",
			Age:       25,
		}

		http.HandleFunc("/decode", func(w http.ResponseWriter, r *http.Request) {
			var user User
			json.NewDecoder(r.Body).Decode(&user)

			fmt.Fprintf(w, "%s %s is %d years old!", user.Firstname, user.Lastname, user.Age)
		})

		json.NewEncoder(w).Encode(peter)
	})

	fs := http.FileServer(http.Dir("static/"))                //permite servir imagenes , css y javascript
	http.Handle("/static/", http.StripPrefix("/static/", fs)) //apuntya al URL

	http.ListenAndServe(":80", nil) //permite aceptar conecciones a nuestro servidor
}
