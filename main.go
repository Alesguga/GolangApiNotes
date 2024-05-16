package main

import (
	"context"
	"github.com/gin-contrib/cors"
	"log"
	"net/http"

	"firebase.google.com/go"
	"firebase.google.com/go/db"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type Note struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var client *db.Client

func main() {
	ctx := context.Background()
	sa := option.WithCredentialsFile("notesasus-firebase-adminsdk-86cpl-d4a6fc676d.json")

	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err = app.DatabaseWithURL(ctx, "https://notesasus-default-rtdb.firebaseio.com/")
	if err != nil {
		log.Fatalf("error initializing database client: %v\n", err)
	}

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // Cambia esto según sea necesario
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowHeaders = []string{"Origin", "Content-Type"}
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
