package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// LogEntry represents a log entry structure
type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

// MongoDBLogger represents the MongoDB logger
type MongoDBLogger struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewMongoDBLogger creates a new MongoDB logger
func NewMongoDBLogger(connectionString, dbName, collectionName string) (*MongoDBLogger, error) {
	clientOptions := options.Client().ApplyURI(connectionString)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection(collectionName)

	return &MongoDBLogger{
		client:     client,
		collection: collection,
	}, nil
}

// Log writes a log entry to MongoDB
func (logger *MongoDBLogger) Log(level, message string) error {
	entry := LogEntry{
		Level:   level,
		Message: message,
	}

	_, err := logger.collection.InsertOne(context.Background(), entry)
	return err
}

// Close closes the MongoDB logger connection
func (logger *MongoDBLogger) Close() {
	if logger.client != nil && logger.client.Ping(context.Background(), nil) == nil {
		err := logger.client.Disconnect(context.Background())
		if err != nil {
			log.Println("Error closing MongoDB connection:", err)
		}
	}
}

// LogHandler handles incoming log entries via HTTP POST requests
func LogHandler(logger *MongoDBLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var entry LogEntry
		err := json.NewDecoder(r.Body).Decode(&entry)
		if err != nil {
			log.Printf("Error decoding request body: %v\n", err)
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		err = logger.Log(entry.Level, entry.Message)
		if err != nil {
			log.Printf("Error logging message: %v\n", err)
			http.Error(w, "Error logging message", http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "Log entry added to MongoDB.")
	}
}

func main() {
	// Replace the following values with your MongoDB connection details
	connectionString := "mongodb://admin:password@mongodb:27017"
	dbName := "logs"
	collectionName := "log_entries"

	logger, err := NewMongoDBLogger(connectionString, dbName, collectionName)
	if err != nil {
		log.Fatal("Error creating MongoDB logger:", err)
	}

	defer logger.Close()

	// Create a new router using go-chi/chi
	mux := chi.NewRouter()

	// Use cors middleware
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Register the LogHandler as the handler for the "/log" endpoint
	mux.Post("/log", LogHandler(logger))

	// Start the HTTP server
	port := 8083
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	err2 := srv.ListenAndServe()
	if err2 != nil {
		log.Panic(err2)
	}
}
