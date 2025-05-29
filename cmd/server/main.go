package main

import (
	"log"
	"net/http"

	"tressette-game/internal/database"
	"tressette-game/internal/server"
)

func main() {
	log.Println("Starting Tressette server...")

	db := database.New()
	defer db.Close()

	hub := server.NewHub(&db)
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.ServeWs(hub, w, r)
	})

	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/", fs)

	server.HandleRoutes(&db)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

