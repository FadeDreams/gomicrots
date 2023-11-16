package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"bytes"
	"encoding/json"
	"errors"
	"io"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"logs_topic", // name
		"topic",      // type
		true,         // durable?
		false,        // auto-deleted?
		false,        // internal?
		false,        // no-wait?
		nil,          // arguments?
	)
}

type Config struct{}

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	// specify who is allowed to connect
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	mux.Use(middleware.Heartbeat("/ping"))

	mux.Post("/", app.Service1)
	mux.Post("/push", app.PushRMQ)
	mux.Post("/log", app.LogH)

	return mux
}

func (app *Config) Service1(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for this handler
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

	payload := jsonResponse{
		Error:   false,
		Message: "Hit the Service1",
	}
	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) PushRMQ(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for this handler
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
	// Connect to RabbitMQ
	rabbitConn, err := amqp.Dial("amqp://guest:guest@rabbitmq")
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConn.Close()

	// Create a channel
	channel, err := rabbitConn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer channel.Close()

	// Declare the exchange
	err = declareExchange(channel)
	if err != nil {
		log.Fatal(err)
	}

	// Publish a message
	message := Payload{
		Name: "log",
		Data: "Your log message data here",
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
	}

	err = channel.Publish(
		"logs_topic", // exchange
		"log.INFO",   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Message published successfully")

	payload := jsonResponse{
		Error:   false,
		Message: "Message published successfully",
	}
	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) LogH(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for this handler
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")

	payload := map[string]string{
		"level":   "data",
		"message": "xx",
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	// Make the POST request
	url := "http://loghandler:8083/log"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error making POST request:", err)
		return
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected response status: %d\n", resp.StatusCode)
		return
	}

	fmt.Println("POST request successful!")
}

type jsonResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// readJSON tries to read the body of a request and converts it into JSON
func (app *Config) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576 // one megabyte

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must have only a single JSON value")
	}

	return nil
}

// writeJSON takes a response status code and arbitrary data and writes a json response to the client
func (app *Config) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

// errorJSON takes an error, and optionally a response status code, and generates and sends
// a json error response
func (app *Config) errorJSON(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload jsonResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}

const webPort = "8081"

func main() {
	app := Config{}

	log.Printf("Starting service1 on port %s\n", webPort)

	// define http server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	// start the server
	err := srv.ListenAndServe()
	if err != nil {
		log.Panic(err)
	}
}
