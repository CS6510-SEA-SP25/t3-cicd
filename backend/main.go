package main

import (
	"cicd/pipeci/backend/db"
	"cicd/pipeci/backend/routes"
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

	// Report endpoints
	router.POST("/report/local", routes.ReportPastExecutionsLocal)
	router.POST("/report/local/query", routes.QueryPastExecutionsLocal)

	// Execute endpoints
	router.POST("/execute/local", routes.ExecuteLocal)

	return router
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("No .env file, use default env variables.")
	} else {
		log.Printf("Loading .env file.")
	}

	db.Init()

	router := setupRouter()
	// Expose
	_ = router.Run("0.0.0.0:8080")
}

func ping(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"success": true})
}
