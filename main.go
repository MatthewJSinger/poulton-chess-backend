package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {

	// connect to the database
	cfg := mysql.Config{
		User:   "root",
		Passwd: "mtZ5PwnHSBqU",
		Net:    "tcp",
		Addr:   "192.168.1.163:3306",
		DBName: "chess_club",
	}

	var err error
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


func getClubChampionshipPlayers(w http.ResponseWriter, r *http.Request) {
	// get list of all players names that are in the club championship sql table
	rows, err := db.Query("SELECT name, rating FROM club_championship")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	players := []map[string]interface{}{}
	for rows.Next() {
		var player_name string
		var player_rating int
		err := rows.Scan(&player_name, &player_rating)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		player := map[string]interface{}{
			"name":   player_name,
			"rating": player_rating,
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
	if err := json.NewDecoder(r.Body).Decode(&player); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// insert the player into the club championship table
	_, err := db.Exec("INSERT INTO club_championship (name, rating) VALUES (?, ?)", player.Name, player.Rating)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)

}	