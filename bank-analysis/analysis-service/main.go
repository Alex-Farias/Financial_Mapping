package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// MonthlySpending represents monthly spending aggregation
type MonthlySpending struct {
	Month            string             `json:"month"`
	Year             int                `json:"year"`
	TotalIncome      float64            `json:"totalIncome"`
	TotalExpenses    float64            `json:"totalExpenses"`
	NetCashflow      float64            `json:"netCashflow"`
	CategoryBreakdown map[string]float64 `json:"categoryBreakdown"`
}

// TransactionList represents a paginated list of transactions
type TransactionList struct {
	Total        int           `json:"total"`
	Transactions []Transaction `json:"transactions"`
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
	
	// Add debug endpoint
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok", 
			"service": "analysis", 
			"userID": userID,
		})
	}).Methods("GET")
	
	// Main routes - notice these are WITHOUT the /api/analysis prefix
	router.HandleFunc("/monthly", getMonthlyAnalysisHandler).Methods("GET")
	router.HandleFunc("/transactions", getTransactionsHandler).Methods("GET")
	router.HandleFunc("/transactions/search", searchTransactionsHandler).Methods("GET")
	router.HandleFunc("/categories", updateCategoryHandler).Methods("PUT")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	log.Printf("Analysis service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func getTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the request
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get query parameters
	limit := 50 // Default limit
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if val, err := strconv.Atoi(limitParam); err == nil && val > 0 {
			limit = val
		}
	}

	offset := 0 // Default offset
	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if val, err := strconv.Atoi(offsetParam); err == nil && val >= 0 {
			offset = val
		}
	}

	// Query transactions from MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count total transactions
	filter := bson.M{"userId": userID}
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("Error counting documents: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get transactions with pagination
	findOptions := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"date": -1}) // Sort by date descending

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("Error finding documents: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Parse results
	var transactions []Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		log.Printf("Error parsing results: %v", err)
		http.Error(w, "Error parsing results", http.StatusInternalServerError)
		return
	}

	// Create response
	result := TransactionList{
		Total:        int(total),
		Transactions: transactions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func searchTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the request
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get search query
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}

	// Create search filter
	filter := bson.M{
		"userId": userID,
		"$or": []bson.M{
			{"description": bson.M{"$regex": query, "$options": "i"}}, // Case-insensitive search
			{"category": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	// Query transactions from MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	findOptions := options.Find().
		SetLimit(100). // Limit search results
		SetSort(bson.M{"date": -1})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("Error searching documents: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Parse results
	var transactions []Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		log.Printf("Error parsing search results: %v", err)
		http.Error(w, "Error parsing results", http.StatusInternalServerError)
		return
	}

	// Create response
	result := TransactionList{
		Total:        len(transactions),
		Transactions: transactions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func updateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the request
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	
	// Parse request body
	var req struct {
		TransactionIDs []string `json:"transactionIds"`
		Category       string   `json:"category"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Invalid request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Category == "" || len(req.TransactionIDs) == 0 {
		http.Error(w, "Category and at least one transaction ID are required", http.StatusBadRequest)
		return
	}
	
	// Convert transaction IDs to ObjectIDs
	var objectIDs []primitive.ObjectID
	for _, id := range req.TransactionIDs {
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			log.Printf("Invalid transaction ID format: %s - %v", id, err)
			http.Error(w, "Invalid transaction ID format", http.StatusBadRequest)
			return
		}
		objectIDs = append(objectIDs, objectID)
	}
	
	// Update transactions
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	filter := bson.M{
		"_id": bson.M{"$in": objectIDs},
		"userId": userID, // Ensure user can only update their own transactions
	}
	
	update := bson.M{
		"$set": bson.M{"category": req.Category},
	}
	
	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Printf("Database error updating categories: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	
	// Create response
	resp := struct {
		Message     string `json:"message"`
		UpdatedCount int64  `json:"updatedCount"`
	}{
		Message:     "Categories updated successfully",
		UpdatedCount: result.ModifiedCount,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getMonthlyAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from the request
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		log.Printf("Missing X-User-ID header for monthly analysis")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("Processing monthly analysis for user: %s", userID)

	// Get start date and end date
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	// Default to last 6 months if not specified
	start := time.Now().AddDate(0, -6, 0)
	end := time.Now()

	if startStr != "" {
		if parsedStart, err := time.Parse("2006-01-02", startStr); err == nil {
			start = parsedStart
		}
	}

	if endStr != "" {
		if parsedEnd, err := time.Parse("2006-01-02", endStr); err == nil {
			end = parsedEnd
		}
	}

	// MongoDB aggregation pipeline
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Group by year and month
	pipeline := mongo.Pipeline{
		// Match user's transactions within date range
		{{Key: "$match", Value: bson.M{
			"userId": userID,
			"date": bson.M{
				"$gte": start,
				"$lte": end,
			},
		}}},
		// Group by year, month, and category
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"year":     bson.M{"$year": "$date"},
				"month":    bson.M{"$month": "$date"},
				"category": "$category",
			},
			"totalAmount": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$eq": []interface{}{"$type", "credit"}},
					"$amount",
					bson.M{"$multiply": []interface{}{"$amount", -1}},
				},
			}},
		}}},
		// Group by year and month to get category breakdown
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"year":  "$_id.year",
				"month": "$_id.month",
			},
			"categories": bson.M{"$push": bson.M{
				"category": "$_id.category",
				"amount":   "$totalAmount",
			}},
			"totalIncome": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$gt": []interface{}{"$totalAmount", 0}},
					"$totalAmount",
					0,
				},
			}},
			"totalExpenses": bson.M{"$sum": bson.M{
				"$cond": []interface{}{
					bson.M{"$lt": []interface{}{"$totalAmount", 0}},
					bson.M{"$abs": "$totalAmount"},
					0,
				},
			}},
		}}},
		// Sort by year and month
		{{Key: "$sort", Value: bson.M{
			"_id.year":  1,
			"_id.month": 1,
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Error in aggregation: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Process results
	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		log.Printf("Error parsing aggregation results: %v", err)
		http.Error(w, "Error parsing results", http.StatusInternalServerError)
		return
	}

	// Convert to MonthlySpending format
	monthNames := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	var monthlySpending []MonthlySpending
	for _, result := range results {
		id := result["_id"].(bson.M)
		year := id["year"].(int32)
		month := id["month"].(int32)
		
		// Create category breakdown map
		categoryBreakdown := make(map[string]float64)
		categories := result["categories"].(bson.A)
		for _, cat := range categories {
			category := cat.(bson.M)
			categoryBreakdown[category["category"].(string)] = category["amount"].(float64)
		}
		
		// Calculate net cashflow
		totalIncome := result["totalIncome"].(float64)
		totalExpenses := result["totalExpenses"].(float64)
		netCashflow := totalIncome - totalExpenses
		
		// Create monthly spending object
		spending := MonthlySpending{
			Month:            monthNames[month-1],
			Year:             int(year),
			TotalIncome:      totalIncome,
			TotalExpenses:    totalExpenses,
			NetCashflow:      netCashflow,
			CategoryBreakdown: categoryBreakdown,
		}
		
		monthlySpending = append(monthlySpending, spending)
	}
	
	log.Printf("Returning %d months of analysis data", len(monthlySpending))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(monthlySpending)
}