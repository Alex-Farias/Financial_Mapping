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
            Keys: bson.D{{"userId", 1}, {"date", -1}},
        },
        {
            Keys: bson.D{{"userId", 1}, {"description", 1}, {"date", 1}, {"amount", 1}},
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

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    // Extract user ID from the request
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }

    // Parse multipart form
    if err := r.ParseMultipartForm(32 << 20); err != nil {
        http.Error(w, "Failed to parse form", http.StatusBadRequest)
        return
    }

    // Get uploaded file
    file, _, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Failed to get file from form", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Process CSV file
    reader := csv.NewReader(file)
    
    // Skip header row
    headerRow, err := reader.Read()
    if err != nil {
        http.Error(w, "Failed to read CSV header", http.StatusBadRequest)
        return
    }
    
    log.Printf("CSV Headers: %v", headerRow)
    
    // Determine column indices based on Nubank format
    dateCol := -1
    amountCol := -1
    identifierCol := -1
    descriptionCol := -1
    
    for i, header := range headerRow {
        header = strings.ToLower(strings.TrimSpace(header))
        switch header {
        case "data":
            dateCol = i
        case "valor":
            amountCol = i
        case "identificador":
            identifierCol = i
        case "descrição", "descricao":
            descriptionCol = i
        }
    }
    
    // Check if required columns were found
    if dateCol == -1 || amountCol == -1 || descriptionCol == -1 {
        log.Printf("Required columns not found. Headers: %v", headerRow)
        http.Error(w, "CSV format not recognized. Requires Data, Valor, and Descrição columns", http.StatusBadRequest)
        return
    }
    
    // Process rows
    var transactions []Transaction
    lineCount := 1 // Header is line 1
    
    for {
        lineCount++
        row, err := reader.Read()
        if err == io.EOF {
            break
        }
        if err != nil {
            http.Error(w, "Failed to read CSV row", http.StatusBadRequest)
            return
        }
        
        log.Printf("Processing row %d: %v", lineCount, row)
        
        // Skip if we don't have enough columns
        if len(row) <= dateCol || len(row) <= amountCol || len(row) <= descriptionCol {
            log.Printf("Row %d has insufficient columns", lineCount)
            continue
        }
        
        // Parse date (Nubank format is likely DD/MM/YYYY)
        dateStr := strings.TrimSpace(row[dateCol])
        date, err := time.Parse("02/01/2006", dateStr)
        if err != nil {
            // Try alternative date formats
            formats := []string{"2006-01-02", "01/02/2006", "02-01-2006"}
            parsed := false
            
            for _, format := range formats {
                if date, err = time.Parse(format, dateStr); err == nil {
                    parsed = true
                    break
                }
            }
            
            if !parsed {
                log.Printf("Row %d: Invalid date format: %s", lineCount, dateStr)
                continue
            }
        }
        
        // Parse description
        description := strings.TrimSpace(row[descriptionCol])
        
        // Parse amount
        amountStr := strings.TrimSpace(row[amountCol])
        amountStr = strings.ReplaceAll(amountStr, ".", "") // Remove thousand separators
        amountStr = strings.ReplaceAll(amountStr, ",", ".") // Convert decimal comma to point
        amountStr = strings.ReplaceAll(amountStr, "R$", "") // Remove currency symbol
        amountStr = strings.TrimSpace(amountStr)
        
        amount, err := strconv.ParseFloat(amountStr, 64)
        if err != nil {
            log.Printf("Row %d: Invalid amount format: %s", lineCount, amountStr)
            continue
        }
        
        // Determine transaction type based on amount
        transType := "debit"
        if amount > 0 {
            transType = "credit"
        }
        amount = math.Abs(amount) // Store amount as positive
        
        // Get category from identifier or set default
        category := "Uncategorized"
        if identifierCol >= 0 && len(row) > identifierCol {
            identifier := strings.TrimSpace(row[identifierCol])
            if identifier != "" {
                category = identifier
            }
        }
        
        transaction := Transaction{
            UserID:      userID,
            Date:        date,
            Description: description,
            Category:    category,
            Amount:      amount,
            Type:        transType,
            Source:      "nubank",
        }
        
        transactions = append(transactions, transaction)
        log.Printf("Added transaction: %+v", transaction)
    }
    
    // Insert transactions into MongoDB
    if len(transactions) > 0 {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        var docs []interface{}
        for _, t := range transactions {
            docs = append(docs, t)
        }
        
        result, err := collection.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
        if err != nil {
            // Continue if some documents failed due to duplication
            log.Printf("Warning during import: %v", err)
        }
        
        insertedCount := 0
        if result != nil {
            insertedCount = len(result.InsertedIDs)
        }
        
        // Create response
        resp := Response{
            Message: fmt.Sprintf("Successfully imported %d of %d transactions", insertedCount, len(transactions)),
            Count:   insertedCount,
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
        return
    }
    
    // No transactions were processed
    http.Error(w, "No valid transactions found in CSV", http.StatusBadRequest)
}

func scanFolderHandler(w http.ResponseWriter, r *http.Request) {
    // Extract user ID from the request
    userID := r.Header.Get("X-User-ID")
    if userID == "" {
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }
    
    // Parse request for folder path
    var req struct {
        FolderPath string `json:"folderPath"`
        Source     string `json:"source"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    if req.FolderPath == "" {
        http.Error(w, "Folder path is required", http.StatusBadRequest)
        return
    }
    
    // Use default source if not provided
    if req.Source == "" {
        req.Source = "import"
    }
    
    // List CSV files in the directory
    files, err := filepath.Glob(filepath.Join(req.FolderPath, "*.csv"))
    if err != nil {
        http.Error(w, "Failed to scan folder", http.StatusInternalServerError)
        return
    }
    
    if len(files) == 0 {
        http.Error(w, "No CSV files found in the specified folder", http.StatusBadRequest)
        return
    }
    
    // Process each CSV file
    totalProcessed := 0
    totalFiles := 0
    
    for _, filePath := range files {
        file, err := os.Open(filePath)
        if err != nil {
            log.Printf("Error opening file %s: %v", filePath, err)
            continue
        }
        
        // Process CSV file
        reader := csv.NewReader(file)
        
        // Skip header row
        _, err = reader.Read()
        if err != nil {
            file.Close()
            continue
        }
        
        // Process rows
        var transactions []Transaction
        
        for {
            row, err := reader.Read()
            if err == io.EOF {
                break
            }
            if err != nil {
                break
            }
            
            // Skip empty rows
            if len(row) < 3 {
                continue
            }
            
            // Parse transaction data from CSV (similar to uploadHandler)
            date, err := time.Parse("2006-01-02", strings.TrimSpace(row[0]))
            if err != nil {
                // Try alternative date format
                date, err = time.Parse("01/02/2006", strings.TrimSpace(row[0]))
                if err != nil {
                    continue
                }
            }
            
            description := strings.TrimSpace(row[1])
            
            amountStr := strings.TrimSpace(row[2])
            amountStr = strings.ReplaceAll(amountStr, ",", "")
            amountStr = strings.ReplaceAll(amountStr, "$", "")
            amount, err := strconv.ParseFloat(amountStr, 64)
            if err != nil {
                continue
            }
            
            transType := "debit"
            if amount > 0 {
                transType = "credit"
            }
            amount = math.Abs(amount)
            
            category := "Uncategorized"
            if len(row) > 3 {
                category = strings.TrimSpace(row[3])
            }
            
            transaction := Transaction{
                UserID:      userID,
                Date:        date,
                Description: description,
                Category:    category,
                Amount:      amount,
                Type:        transType,
                Source:      req.Source,
            }
            
            transactions = append(transactions, transaction)
        }
        
        file.Close()
        
        // Insert transactions into MongoDB
        if len(transactions) > 0 {
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            
            var docs []interface{}
            for _, t := range transactions {
                docs = append(docs, t)
            }
            
            result, err := collection.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
            cancel()
            
            if err != nil {
                log.Printf("Warning during import of file %s: %v", filePath, err)
            }
            
            insertedCount := 0
            if result != nil {
                insertedCount = len(result.InsertedIDs)
            }
            
            totalProcessed += insertedCount
            totalFiles++
        }
    }
    
    // Create response
    resp := Response{
        Message: fmt.Sprintf("Successfully processed %d files and imported %d transactions", totalFiles, totalProcessed),
        Count:   totalProcessed,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}