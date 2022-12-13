package routes

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/notjacobjun/url_shortener/database"
)

func ResolveURL(c *fiber.Ctx) error {
	// get the url from the request parameters
	url := c.Params("url")
	r := database.CreateClient(0)
	defer r.Close()

	// access the database to find out the actual url that corresponds to the shortened url
	val, err := r.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short not found in database"})
	} else if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot connect to the database"})
	}

	rInr := database.CreateClient(1)
	_ = rInr.Incr(database.Ctx, "counter")

	return c.Redirect(val, 301)
}
