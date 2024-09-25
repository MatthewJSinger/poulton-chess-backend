package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	var err error

	password := os.Getenv("SECRET")
	db_host := os.Getenv("DB_IP") + ":3306"

	// connect to the database
	cfg := mysql.Config{
		User:   "root",
		Passwd: password,
		Net:    "tcp",
		Addr:   db_host,
		DBName: "chess_club",
	}


	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		panic(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		panic(pingErr)
	}
	fmt.Println("Connected to database")

	// create a new router
	r := chi.NewRouter()
	r.Use(corsMiddleware)
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})


	r.Route("/club-championship", func(r chi.Router) {
		r.Get("/players", getClubChampionshipPlayers)
		r.Post("/add-player", addClubChampionshipPlayer)
	});

	http.ListenAndServe(":3000", r)
	fmt.Println("Server started on port 3000")

}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}


func getClubChampionshipPlayers(w http.ResponseWriter, r *http.Request) {
	// get list of all players names that are in the club championship sql table
	rows, err := db.Query("SELECT name, rating, points FROM club_championship")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	players := []map[string]interface{}{}
	for rows.Next() {
		var player_name string
		var player_rating int
		var player_points int
		err := rows.Scan(&player_name, &player_rating, &player_points)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		player := map[string]interface{}{
			"name":   player_name,
			"rating": player_rating,
			"points": player_points,
		}
		players = append(players, player)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(players); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func addClubChampionshipPlayer(w http.ResponseWriter, r *http.Request) {
	// add a new player to the club championship sql table

	// get the player name and rating from the request body
	var player struct {
		Name   string `json:"name"`
		Rating int    `json:"rating"`
	}
	fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	fmt.Println(player)

	// Validate input
	if player.Name == "" || player.Rating < 0 || player.Rating > 3000 {
		http.Error(w, "Invalid player data", http.StatusBadRequest)
		return
	}
	rows, err := db.Query("SELECT name, rating FROM club_championship")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var existingPlayerName string
		var existingPlayerRating int
		err := rows.Scan(&existingPlayerName, &existingPlayerRating)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if existingPlayerName == player.Name {
			http.Error(w, "Player already exists", http.StatusConflict)
			return
		}
	}


	// Use prepared statements to prevent SQL injection
	stmt, err := db.Prepare("INSERT INTO club_championship (name, rating) VALUES (?, ?)")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(player.Name, player.Rating)
	if err != nil {
		http.Error(w, "Failed to add player", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Player added successfully"))
}