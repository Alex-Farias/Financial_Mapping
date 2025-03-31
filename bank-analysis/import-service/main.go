package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Transaction represents a financial transaction
type Transaction struct {
	UserID      string    `json:"userId" bson:"userId"`
	Date        time.Time `json:"date" bson:"date"`
	Description string    `json:"description" bson:"description"`
	Category    string    `json:"category" bson:"category"`
	Amount      float64   `json:"amount" bson:"amount"`
	Type        string    `json:"type" bson:"type"` // "credit" or "debit"
	Source      string    `json:"source" bson:"source"` // "checking" or "credit_card"
}

// Response represents the HTTP response
type Response struct {
	Message string `json:"message"`
	Count   int    `json:"count,omitempty"`
}

var client *mongo.Client
var collection *mongo.Collection

func main() {
	// MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("bank_analysis").Collection("transactions")

	// Create indexes for better query performance
	indexModels := []mongo.IndexModel{
		{
			Keys: bson.M{"userId": 1, "date": -1},
		},
		{
			Keys: bson.M{"userId": 1, "description": 1, "date": 1, "amount": 1},
			Options: options.Index().SetUnique(true), // Prevent duplicate entries
		},
	}

	_, err = collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		log.Fatal(err)
	}

	// HTTP server
	router := mux.NewRouter()
	router.HandleFunc("/upload", uploadHandler).Methods("POST")
	router.HandleFunc("/scan", scanFolderHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	log.Printf("Import service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}