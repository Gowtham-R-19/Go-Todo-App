package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Todo structure for JSON
type Todo struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Status  string `json:"status"`
	Date    string `json:"date"` // Add date field for reminder
}

// Metric structure for container metrics
type Metric struct {
	Name        string  `json:"name"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage string  `json:"memory_usage"`
}

var client *mongo.Client

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI not set in .env file")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}

	r := gin.Default()
	r.Static("/public", "./public") // Serve static files from the public folder
	r.StaticFile("/", "./public/index.html")

	r.POST("/todos", createTodo)
	r.GET("/todos", fetchTodos)
	r.PATCH("/todos/:id", updateTodoStatus) // New route for updating status
	r.PUT("/todos/:id", updateTodo)
	r.DELETE("/todos/:id", deleteTodo)

	r.GET("/metrics", getMetrics) // Metrics endpoint

	// Handle dynamic port for Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default for local
	}
	r.Run(":" + port)
}

// Create a new todo
func createTodo(c *gin.Context) {
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	todo.ID = primitive.NewObjectID().Hex()
	todo.Status = "incomplete" // Default status
	_, err := client.Database("testdb").Collection("todos").InsertOne(context.TODO(), todo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create todo"})
		return
	}

	c.JSON(http.StatusCreated, todo)
}

// Fetch all todos
func fetchTodos(c *gin.Context) {
	cursor, err := client.Database("testdb").Collection("todos").Find(context.TODO(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch todos"})
		return
	}
	defer cursor.Close(context.TODO())

	var todos []Todo
	for cursor.Next(context.TODO()) {
		var todo Todo
		if err := cursor.Decode(&todo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not decode todo"})
			return
		}
		todos = append(todos, todo)
	}

	if err := cursor.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cursor error"})
		return
	}

	c.JSON(http.StatusOK, todos)
}

// Update a todo's status
func updateTodoStatus(c *gin.Context) {
	id := c.Param("id")
	var update struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := client.Database("testdb").Collection("todos").UpdateOne(context.TODO(),
		bson.M{"id": id}, bson.M{"$set": bson.M{"status": update.Status}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update todo status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo status updated"})
}

// Update a todo
func updateTodo(c *gin.Context) {
	id := c.Param("id")
	var todo Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	_, err := client.Database("testdb").Collection("todos").UpdateOne(context.TODO(),
		bson.M{"id": id}, bson.M{"$set": bson.M{"content": todo.Content, "date": todo.Date}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update todo"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

// Delete a todo
func deleteTodo(c *gin.Context) {
	id := c.Param("id")
	_, err := client.Database("testdb").Collection("todos").DeleteOne(context.TODO(), bson.M{"id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete todo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
}

// Get container metrics with mock data
func getMetrics(c *gin.Context) {
	// Mock data for container metrics
	metrics := []Metric{
		{Name: "Container1", CPUUsage: 23.5, MemoryUsage: "512MiB / 1GiB"},
		{Name: "Container2", CPUUsage: 45.0, MemoryUsage: "256MiB / 1GiB"},
		{Name: "Container3", CPUUsage: 12.3, MemoryUsage: "128MiB / 512MiB"},
		{Name: "Container4", CPUUsage: 67.8, MemoryUsage: "1GiB / 2GiB"},
	}

	c.JSON(http.StatusOK, metrics)
}
