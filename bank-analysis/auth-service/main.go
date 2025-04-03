package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// User represents the user model stored in MongoDB
type User struct {
	ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Email    string             `json:"email" bson:"email"`
	Password string             `json:"password" bson:"password"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest represents the register request body
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response represents the HTTP response
type Response struct {
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	Email   string `json:"email,omitempty"`
}

var client *mongo.Client
var collection *mongo.Collection
var jwtKey []byte

func main() {
	// MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your_secret_key_change_in_production"
	}
	jwtKey = []byte(jwtSecret)

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("bank_analysis").Collection("users")

	// Create a unique index on the email field
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Fatal(err)
	}

	// HTTP server
	router := mux.NewRouter()
	
	// Add debug endpoint
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok", 
			"service": "auth",
		})
	}).Methods("GET")
	
	// IMPORTANT: Routes must NOT have the /api/auth prefix since the API gateway adds it
	router.HandleFunc("/register", registerHandler).Methods("POST")
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/validate", validateTokenHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Auth service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Bad request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Could not hash password: %v", err)
		http.Error(w, "Could not hash password", http.StatusInternalServerError)
		return
	}

	user := User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	// Insert user into MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fix: Properly handle or ignore the insertion result
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("User creation error: %v", err)
		http.Error(w, "Email already exists or server error", http.StatusInternalServerError)
		return
	}

	// Generate JWT
	token, err := generateToken(req.Email)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Create response
	resp := Response{
		Message: "User registered successfully",
		Token:   token,
		Email:   req.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Bad request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find user in MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user User
	err = collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		log.Printf("User not found: %v", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		log.Printf("Password mismatch: %v", err)
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate JWT
	token, err := generateToken(req.Email)
	if err != nil {
		log.Printf("Token generation error: %v", err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	// Create response
	resp := Response{
		Message: "Login successful",
		Token:   token,
		Email:   req.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func validateTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		log.Printf("Missing Authorization header")
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Remove "Bearer " prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		log.Printf("Invalid token: %v", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("Invalid token claims format")
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	
	email, ok := claims["email"].(string)
	if !ok {
		log.Printf("Email claim missing from token")
		http.Error(w, "Invalid token: missing email claim", http.StatusUnauthorized)
		return
	}

	// Token is valid
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Message: "Token is valid", Email: email})
}

func generateToken(email string) (string, error) {
	// Create claims with user email and expiry time
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Token valid for 24 hours
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		log.Printf("Error signing token: %v", err)
		return "", err
	}

	log.Printf("Generated token for user: %s", email)
	return tokenString, nil
}