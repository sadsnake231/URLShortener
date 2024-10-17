package routes

import (
	"github.com/asaskevich/govalidator"
	_ "github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"os"
	_ "strconv"
	"url_short_v2/api/database"
	"url_short_v2/api/helpers"
)

type request struct {
	URL         string `json:"url"`
	CustomShort string `json:"short"`
}

type response struct {
	URL         string `json:"url"`
	CustomShort string `json:"short"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "can't parse JSON"})
	}

	//check if the input is an actual url

	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	//check for domain err

	if !helpers.RemoveDomainErr(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "access restricted"})
	}

	//enforce https, ssl

	body.URL = helpers.EnforceHTTP(body.URL)
	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)]
	defer r.Close()

	val, err := r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "already taken"})
	}

	err = r.Set(database.Ctx, id, body.URL, 0).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "couldnt put"})
	}


	resp := response{
		URL:         body.URL,
		CustomShort: "",
	}
	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id
	return c.Status(fiber.StatusOK).JSON(resp)

}
