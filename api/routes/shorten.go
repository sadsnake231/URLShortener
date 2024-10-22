package routes

import (
	"github.com/asaskevich/govalidator"
	_ "github.com/go-redis/redis/v8"
	_ "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"os"
	_ "strconv"
	"net/http"
	"url_short_v2/api/database"
	"url_short_v2/api/helpers"
	"github.com/gin-gonic/gin"
)

type request struct {
	URL         string `json:"url"`
	CustomShort string `json:"short"`
}

type response struct {
	URL         string `json:"url"`
	CustomShort string `json:"short"`
}


func ShortenURL() gin.HandlerFunc {
	return func(c *gin.Context){
		body := new(request)

		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "can't parse JSON"})
			return
		}

	//check if the input is an actual url

		if !govalidator.IsURL(body.URL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
			return
		}

	//check for domain err

		if !helpers.RemoveDomainErr(body.URL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "access restricted"})
			return
		}

	//enforce https, ssl

		body.URL = helpers.EnforceHTTP(body.URL)
		var id string
		if body.CustomShort == "" {
			id = uuid.New().String()[:6]
		} else {
			id = body.CustomShort
		}

		r := database.CreateClient(0)
		defer r.Close()

		val, err := r.Get(database.Ctx, id).Result()
		if val != "" {
			c.JSON(http.StatusForbidden,gin.H{"error": "already taken"})
			return
		}
		
		err = r.Set(database.Ctx, id, body.URL, 0).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{"error": err.Error()})
			return
		}

		resp := response{
			URL:         body.URL,
			CustomShort: "",
		}
		resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
		c.JSON(http.StatusOK, resp)
	}
}