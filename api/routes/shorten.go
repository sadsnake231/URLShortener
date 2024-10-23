package routes

import (
	"github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"os"
	"strconv"
	"net/http"
	"url_short_v2/api/database"
	"url_short_v2/api/helpers"
	"github.com/gin-gonic/gin"
	"time"
	"fmt"
)

type request struct {
	URL         	string 		`json:"url"`
	CustomShort 	string 		`json:"short"`
}

type response struct {
	URL         	string 				`json:"url"`
	CustomShort 	string 				`json:"short"`
	UsagesLeft		int 				`json:"left"`
	RefreshTime 	time.Duration 		`json:"refresh"`
}


func ShortenURL() gin.HandlerFunc {
	return func(c *gin.Context){
		body := new(request)
		resp := new(response)
		const API_QUOTA = 10

		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "can't parse JSON"})
			return
		}
		
		
		r2 := database.CreateClient(1) //opening second database for IP checking
		defer r2.Close()

		val, err := r2.Get(database.Ctx, c.ClientIP()).Result() //how many usages left
		valInt, _ := strconv.Atoi(val)
		fmt.Println(val)
		limit, _ := r2.TTL(database.Ctx, c.ClientIP()).Result() //how many time is left before limit refreshing
		if val == ""{ //if there isn't any record of this IP
			fmt.Println("first")
			_ = r2.Set(database.Ctx, c.ClientIP(), API_QUOTA - 1, 30*60*time.Second).Err() //putting IPislimit, how many usages are left (MAX - 1) and time left before refreshing
			valInt = API_QUOTA
			limit = time.Minute * 30
		} else { // if there is
			fmt.Println("second")
			if valInt <= 0{ // if no usages left
				c.JSON(http.StatusForbidden, gin.H{
					"error": "No usages left",
					"Time_left": limit / time.Nanosecond / time.Minute,
				})
				return
			}
		}


		if !govalidator.IsURL(body.URL) { //checking if the input is an actual url
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL"})
			return
		}


		if !helpers.RemoveDomainErr(body.URL) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "access restricted"})
			return
		}

		body.URL = helpers.EnforceHTTP(body.URL)
		var id string
		if body.CustomShort == "" {
			id = uuid.New().String()[:6]
		} else {
			id = body.CustomShort
		}

		r := database.CreateClient(0)
		defer r.Close()

		val, _ = r.Get(database.Ctx, id).Result()
		if val != "" {
			c.JSON(http.StatusForbidden,gin.H{"error": "already taken"})
			return
		}
		
		err = r.Set(database.Ctx, id, body.URL, 30*24*60*60*time.Second).Err() //expiration date of custom url is 30 days
		if err != nil {
			c.JSON(http.StatusInternalServerError,gin.H{"error": err.Error()})
			return
		}

		resp.URL = body.URL
		resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
		err = r2.Set(database.Ctx, c.ClientIP(), valInt - 1, limit).Err() //decreasing the number of usages left
		if err != nil{
			fmt.Println(err.Error())
		}
		resp.UsagesLeft = valInt - 1
		resp.RefreshTime = limit / time.Nanosecond / time.Minute
		c.JSON(http.StatusOK, resp)
	}
}