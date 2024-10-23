package routes

import (
	_ "github.com/go-redis/redis/v8"
	_ "github.com/gofiber/fiber/v2"
	"url_short_v2/api/database"
	"github.com/gin-gonic/gin"
	"net/http"
)


func ResolveURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		var url string
		url = c.Param("url")

		r := database.CreateClient(0)
		defer r.Close()

		value, err := r.Get(database.Ctx, url).Result()
		if value == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't connect to the DB"})
			return
		}

		c.Redirect(301, value)
		return
	}
}


