package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/notjacobjun/url_shortener/database"
	"github.com/notjacobjun/url_shortener/helpers"
)

// defining the request and the response structs
type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"short"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(response)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	// implement the rate limiter
	r2 := database.CreateClient(1)
	defer r2.Close()
	_, err := r2.Get(database.Ctx, c.IP()).Result()
	// error checking
	if err == redis.Nil {
		// in this case the user's IP address isn't stored in our database, therefore we have to set the IP address
		// also set the API quota
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		val, _ := r2.Get(database.Ctx, c.IP()).Result()
		// get the current user's quote right now
		valInt, _ := strconv.Atoi(val)
		// check if this user is allowed to access our service
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Rate limit exceeded", "rate_limit_reset": limit / time.Nanosecond / time.Minute})
		}
	}

	if body.XRateRemaining <= 0 {
		// check if we should reset the rate or maybe just
		// notify the user that we can't perform this operation
	} else {
		body.XRateRemaining--
	}
	// check if the url that the user entered is a valid url
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "entered an invalid url"})
	}

	// check for edge case where the url entered is localhost:3000
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "entered an invalid domain"})
	}

	// enforce https
	body.URL = helpers.EnforceHTTP(body.URL)
	// parse any custom short url that the user specified (if there exists any)
	var id string
	if body.CustomShort == "" {
		// if the user hasn't specified any specific custom short then we can just find out
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()

	val, _ := r.Get(database.Ctx, id).Result()
	// check if this custom short is already used by some other user
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "URL custom short is already being used"})
	}

	if body.Expiry == 0 {
		body.Expiry = 24
	}

	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to connect to server"})
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}

	// decrement the limit quota each time that the user accesses this service
	r2.Decr(database.Ctx, c.IP())

	// update the rate remaining value for our struct
	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	// set the remaining values for the response
	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute
	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
