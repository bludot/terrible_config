package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/bludot/dynamic_config/config"
	"github.com/bludot/dynamic_config/hbconfig"
	"github.com/gofiber/fiber/v2"
)

func loadConfigOrPanic() config.Config {
	cfg := config.Config{}
	_ = hbconfig.NewDynamicConfig(&cfg, os.Getenv("VAULT_AGENT_SECRETS_PATH"))

	c := hbconfig.GetDynamicConfig()
	conf := c.(*config.Config)

	return *conf
}
func logFile() {
	conf := hbconfig.GetDynamicConfig()
	confString, _ := json.Marshal(conf)
	log.Println(string(confString))
	time.Sleep(time.Second * 10)
	logFile()
}

func main() {

	_ = loadConfigOrPanic()
	log.Println("Test")
	go logFile()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		conf := hbconfig.GetDynamicConfig()
		confString, _ := json.Marshal(conf)
		return c.SendString(string(confString))
	})

	err := app.Listen(":3001")
	if err != nil {
		log.Fatal(err)
	}
}
