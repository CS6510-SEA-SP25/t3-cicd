package main

import (
	"cicd/pipeci/worker/db"
	"cicd/pipeci/worker/queue"
	"cicd/pipeci/worker/storage"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	// Force log's color
	gin.ForceConsoleColor()
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		log.Printf("endpoint %v %v %v %v\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}

	// router := gin.Default()
	router := gin.New()

	// Global middleware
	// Logger middleware will write the logs to gin.DefaultWriter even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	router.Use(gin.Logger())

	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	router.Use(gin.Recovery())

	// Ping
	router.GET("/", ping)

	// Consime messages in the background
	go queue.Consume()

	return router
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("No .env file, use default env variables.")
	} else {
		log.Printf("Loading .env file.")
	}

	// Init database
	db.Init()

	// Init artifact storage
	storage.Init()

	router := setupRouter()

	// Expose
	_ = router.Run("0.0.0.0:8081")
}

func ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"success": true})
}
