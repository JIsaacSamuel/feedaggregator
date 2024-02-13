module github.com/JIsaacSamuel/feedaggregator

go 1.21.5

require (
	github.com/go-chi/chi/v5 v5.0.11
	github.com/go-chi/cors v1.2.1
	github.com/joho/godotenv v1.5.1
)

require github.com/lib/pq v1.10.9

require internal/database v1.0.0

require github.com/google/uuid v1.6.0

replace internal/database => ./internal/database
