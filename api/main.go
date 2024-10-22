package main

import (
	"fmt"
	_ "github.com/gofiber/fiber/v2"
	_ "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"url_short_v2/api/routes"
)


func Routes(incomingRoutes *gin.Engine){
	incomingRoutes.GET("/:url", routes.ResolveURL())
	incomingRoutes.POST("/api/v1", routes.ShortenURL())
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}
	
	router := gin.New()
	router.Use(gin.Logger())
	Routes(router)


	log.Fatal(router.Run(os.Getenv("APP_PORT")))
}
