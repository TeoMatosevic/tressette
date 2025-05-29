package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"tressette-game/internal/database"
)

func HandleRoutes(db *database.Service) {
	http.HandleFunc("/api/results/player/{name}", func(w http.ResponseWriter, r *http.Request) {
		GetResultsByPlayerHandler(db, w, r)
	})

	log.Println("Registered route: /api/results/player/{name}")

	http.HandleFunc("/api/results", func(w http.ResponseWriter, r *http.Request) {
		GetResultsHandler(db, w, r)
	})

	log.Println("Registerd route: /api/results")
}

func GetResultsByPlayerHandler(db *database.Service, w http.ResponseWriter, r *http.Request) {
	// Logic to handle fetching results by player ID

	player := r.PathValue("name")
	if player == "" {
		http.Error(w, "Player name is required", http.StatusBadRequest)
		return
	}

	results, err := db.GetByPlayer(player)
	if err != nil {
		// Not found
		if err == sql.ErrNoRows {
			http.Error(w, "No results found for player", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to fetch results", http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func GetResultsHandler(db *database.Service, w http.ResponseWriter, r *http.Request) {
	// Logic to handle fetching game results
	results, err := db.GetAll()
	if err != nil {
		http.Error(w, "Failed to fetch results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
