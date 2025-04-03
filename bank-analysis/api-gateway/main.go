package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var (
	authServiceURL      = getEnv("AUTH_SERVICE_URL", "http://auth-service:8081")
	importServiceURL    = getEnv("IMPORT_SERVICE_URL", "http://import-service:8082")
	analysisServiceURL  = getEnv("ANALYSIS_SERVICE_URL", "http://analysis-service:8083")
	exportServiceURL    = getEnv("EXPORT_SERVICE_URL", "http://export-service:8084")
	jwtKey              = []byte(getEnv("JWT_SECRET", "your_secret_key"))
)

func main() {
	router := mux.NewRouter()

	// Setup CORS middleware
	router.Use(corsMiddleware)

	// Public routes (no authentication required)
	publicRouter := router.PathPrefix("/api").Subrouter()
	publicRouter.PathPrefix("/auth").Handler(createReverseProxy(authServiceURL))

	// Protected routes (authentication required)
	protectedRouter := router.PathPrefix("/api").Subrouter()
	protectedRouter.Use(authMiddleware)

	// Route to services based on path prefix
	protectedRouter.PathPrefix("/import").Handler(createReverseProxy(importServiceURL))
	protectedRouter.PathPrefix("/analysis").Handler(createReverseProxy(analysisServiceURL))
	protectedRouter.PathPrefix("/export").Handler(createReverseProxy(exportServiceURL))
	protectedRouter.PathPrefix("/transactions").Handler(createReverseProxy(analysisServiceURL))
	protectedRouter.PathPrefix("/categories").Handler(createReverseProxy(analysisServiceURL))

	// Static file server for frontend
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static")))

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("API Gateway listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

// createReverseProxy creates a reverse proxy to the specified target URL
func createReverseProxy(targetURL string) http.Handler {
	url, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	// Modify request to strip the /api/{service} prefix and add X-User-ID header
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		
		// Save original URL for logging
		originalPath := req.URL.Path
		
		// Remove /api prefix
		path := strings.TrimPrefix(originalPath, "/api")
		
		// Remove service prefix (keep only the endpoint part)
		parts := strings.SplitN(path, "/", 3)
		if len(parts) >= 3 {
			// If path is like /import/upload, we want just /upload
			req.URL.Path = "/" + parts[2]
		} else if len(parts) == 2 {
			// If path is like /import/?, we want just /?
			if strings.Contains(parts[1], "?") {
				req.URL.Path = "/"
			} else {
				// If path is like /import, we want just /
				req.URL.Path = "/"
			}
		}
		
		log.Printf("Forwarding request: %s -> %s%s", originalPath, targetURL, req.URL.Path)
		
		// Get user ID from context
		if userID, ok := req.Context().Value("userID").(string); ok {
			req.Header.Set("X-User-ID", userID)
			log.Printf("Added X-User-ID header: %s", userID)
		}
	}

	return proxy
}

// authMiddleware validates JWT tokens and adds userID to request context
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("Missing Authorization header for: %s", r.URL.Path)
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Remove Bearer prefix if present
		tokenString := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		}

		log.Printf("Processing token: %s... for path: %s", tokenString[:10], r.URL.Path)

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil {
			log.Printf("Token parsing error: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Check if token is valid
		if !token.Valid {
			log.Printf("Token is invalid")
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

		// Set user ID in context
		userID, ok := claims["email"].(string)
		if !ok {
			log.Printf("Invalid user ID in token")
			http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
			return
		}

		log.Printf("User authenticated: %s for path: %s", userID, r.URL.Path)

		// Create a new request context with the user ID
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin
		w.Header().Set("Access-Control-Allow-Origin", "*")
		
		// Allow common HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		
		// Allow common headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}