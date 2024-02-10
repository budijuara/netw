package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// options
	portListen := flag.String("port", "33445", "Port to listen")
	httpTarget := flag.String("target", "http://localhost:34567", "Target URL")
	flag.Parse()
	// Fiber instance
	app := fiber.New()

	// Logger middleware with custom configuration for verbose logging
	app.Use(logger.New(logger.Config{
		Format:     "${pid} ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006 15:04:05",
		TimeZone:   "Local",
	}))

	app.All("/*", func(c *fiber.Ctx) error {
		// Destination URL
		url := *httpTarget + c.OriginalURL()

		// Forwarding the request to the destination URL
		resp, err := forwardRequest(c.Method(), url, c, c.Body())
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		defer resp.Body.Close()

		// Reading the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return c.Status(500).SendString("Failed to read response body")
		}

		// Setting the response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Response().Header.Add(key, value)
			}
		}

		// Returning the response from the destination
		return c.Status(resp.StatusCode).Send(body)
	})

	fmt.Printf("Listening on port %s, forwarded to :%s\n", *portListen, *httpTarget)
	log.Fatal(app.Listen(":" + *portListen))
}

func forwardRequest(method, url string, ctx *fiber.Ctx, body []byte) (*http.Response, error) {
	// Convert fasthttp headers to net/http headers
	netHeaders := http.Header{}
	ctx.Request().Header.VisitAll(func(key, value []byte) {
		netHeaders.Add(string(key), string(value))
	})

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header = netHeaders

	return client.Do(req)
}
