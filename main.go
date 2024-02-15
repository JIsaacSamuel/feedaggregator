package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"internal/database"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	DB *database.Queries
}

func main() {
	err := godotenv.Load("configs.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		log.Fatal("Unable to load PORT")
	}
	dbURL := os.Getenv("CONN")
	if dbURL == "" {
		log.Fatal("Unable to load dbURL")
	}

	db, err := sql.Open("postgres", dbURL)

	dbQueries := database.New(db)

	apiCfg := apiConfig{
		DB: dbQueries,
	}

	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	subr1 := chi.NewRouter()
	subr1.Post("/users", apiCfg.handleUserCreate)
	subr1.Get("/users", apiCfg.handleGetUserApiKey)
	subr1.Get("/readiness", handleReadiness)
	subr1.Get("/err", handleErr)
	router.Mount("/v1", subr1)

	srv := http.Server{
		Addr:    ":" + PORT,
		Handler: router,
	}

	log.Printf("Serving on port: %s\n", PORT)
	log.Fatal(srv.ListenAndServe())
}

// helper functions
func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func getApiKey(r *http.Request) (string, error) {
	apiHeader := r.Header.Get("Authorization")
	if apiHeader == "" {
		return "", errors.New("API Key not included")
	}

	tempSlice := strings.Split(apiHeader, " ")
	if len(tempSlice) < 2 || tempSlice[0] != "ApiKey" {
		return "", errors.New("Malformed Authorization header")
	}

	return tempSlice[1], nil
}

// handler functions
func (cfg *apiConfig) handleGetUserApiKey(w http.ResponseWriter, r *http.Request) {
	apiKey, err := getApiKey(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}

	user, err := cfg.DB.GetUserByApiKey(r.Context(), apiKey)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) handleUserCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, user)
}

// Dummy functions
func handleReadiness(w http.ResponseWriter, r *http.Request) {
	type Ready struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, 200, Ready{
		Status: "ok",
	})
}

func handleErr(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 500, "Internal Server Error")
}
