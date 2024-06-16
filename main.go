package main

import (
	"context"
	"encoding/json"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	"firebase.google.com/go"
	"firebase.google.com/go/db"
	"google.golang.org/api/option"
)

type Note struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var client *db.Client

func main() {
	// Cargar variables de entorno desde el archivo .env
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	ctx := context.Background()
	creds := map[string]string{
		"type":                        os.Getenv("FIREBASE_TYPE"),
		"project_id":                  os.Getenv("FIREBASE_PROJECT_ID"),
		"private_key_id":              os.Getenv("FIREBASE_PRIVATE_KEY_ID"),
		"private_key":                 os.Getenv("FIREBASE_PRIVATE_KEY"),
		"client_email":                os.Getenv("FIREBASE_CLIENT_EMAIL"),
		"client_id":                   os.Getenv("FIREBASE_CLIENT_ID"),
		"auth_uri":                    os.Getenv("FIREBASE_AUTH_URI"),
		"token_uri":                   os.Getenv("FIREBASE_TOKEN_URI"),
		"auth_provider_x509_cert_url": os.Getenv("FIREBASE_AUTH_PROVIDER_X509_CERT_URL"),
		"client_x509_cert_url":        os.Getenv("FIREBASE_CLIENT_X509_CERT_URL"),
	}

	credsJSON, err := json.Marshal(creds)
	if err != nil {
		log.Fatalf("error marshalling creds: %v", err)
	}

	opt := option.WithCredentialsJSON(credsJSON)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err = app.DatabaseWithURL(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("error initializing database client: %v\n", err)
	}

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowHeaders = []string{"*"}
	router.Use(cors.New(config))

	router.POST("/notes", createNote)
	router.GET("/notes", getNotes)
	router.GET("/notes/:id", getNote)
	router.PUT("/notes/:id", updateNote)
	router.DELETE("/notes/:id", deleteNote)

	err = router.Run(":8000")
	if err != nil {
		return
	}
}

func createNote(c *gin.Context) {
	var note Note
	if err := c.ShouldBindJSON(&note); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("Received note: %+v", note)

	ref, err := client.NewRef("notes").Push(context.Background(), nil)
	if err != nil {
		log.Printf("Error creating reference: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	note.ID = ref.Key
	log.Printf("Generated ID: %s", note.ID)

	if err := ref.Set(context.Background(), note); err != nil {
		log.Printf("Error setting note in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, note)
}

func getNotes(c *gin.Context) {
	var notes map[string]Note
	if err := client.NewRef("notes").Get(context.Background(), &notes); err != nil {
		log.Printf("Error getting notes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
}

func getNote(c *gin.Context) {
	id := c.Param("id")
	var note Note
	if err := client.NewRef("notes/"+id).Get(context.Background(), &note); err != nil {
		log.Printf("Error getting note: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	c.JSON(http.StatusOK, note)
}

func updateNote(c *gin.Context) {
	id := c.Param("id")
	var note Note
	if err := c.ShouldBindJSON(&note); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := client.NewRef("notes/"+id).Set(context.Background(), note); err != nil {
		log.Printf("Error setting note in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, note)
}

func deleteNote(c *gin.Context) {
	id := c.Param("id")
	if err := client.NewRef("notes/" + id).Delete(context.Background()); err != nil {
		log.Printf("Error deleting note: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
