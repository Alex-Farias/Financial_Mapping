package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
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
	Source      string    `json:"source" bson:"source"`
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

	// HTTP server
	router := mux.NewRouter()
	router.HandleFunc("/export/csv", exportCSVHandler).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	log.Printf("Export service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func exportCSVHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the request
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get optional date range
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	// Create filter
	filter := bson.M{"userId": userID}

	// Add date range if provided
	if startStr != "" || endStr != "" {
		dateFilter := bson.M{}

		if startStr != "" {
			start, err := time.Parse("2006-01-02", startStr)
			if err == nil {
				dateFilter["$gte"] = start
			}
		}

		if endStr != "" {
			end, err := time.Parse("2006-01-02", endStr)
			if err == nil {
				dateFilter["$lte"] = end
			}
		}

		if len(dateFilter) > 0 {
			filter["date"] = dateFilter
		}
	}

	// Query transactions from MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	findOptions := options.Find().SetSort(bson.M{"date": 1}) // Sort by date ascending
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Parse results
	var transactions []Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		http.Error(w, "Error parsing results", http.StatusInternalServerError)
		return
	}

	// Set response headers for CSV download
	filename := fmt.Sprintf("bank_transactions_%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Create CSV writer
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write CSV header
	header := []string{"Date", "Description", "Category", "Amount", "Type", "Source"}
	if err := csvWriter.Write(header); err != nil {
		http.Error(w, "Error writing CSV", http.StatusInternalServerError)
		return
	}

	// Write transactions
	for _, t := range transactions {
		// Format amount based on transaction type
		amount := t.Amount
		if t.Type == "debit" {
			amount = -amount
		}

		row := []string{
			t.Date.Format("2006-01-02"),
			t.Description,
			t.Category,
			fmt.Sprintf("%.2f", amount), // Format to 2 decimal places
			t.Type,
			t.Source,
		}

		if err := csvWriter.Write(row); err != nil {
			http.Error(w, "Error writing CSV", http.StatusInternalServerError)
			return
		}
	}
}