package hbconfig_test

import (
	"log"
	"os"
	"path"
	"runtime"
	"testing"
	"time"

	"github.com/bludot/dynamic_config/config"
	"github.com/bludot/dynamic_config/hbconfig"
	"github.com/stretchr/testify/assert"
)

func TestNewDynamicConfig(t *testing.T) {
	t.Run("should return a new DynamicConfigService", func(t *testing.T) {
		a := assert.New(t)
		cfg := config.Config{}
		dynamicConfigService := hbconfig.NewDynamicConfig(&cfg)
		err := dynamicConfigService.LoadConfig()
		a.NoError(err)
		a.NotNil(dynamicConfigService)
	})
}

func TestGetDynamicConfig(t *testing.T) {
	t.Run("should return the config", func(t *testing.T) {
		cfg := config.Config{}
		a := assert.New(t)
		dynamicConfigService := hbconfig.NewDynamicConfig(&cfg)
		a.NotNil(dynamicConfigService)
		err := dynamicConfigService.LoadConfig()
		a.NoError(err)
		conf := hbconfig.GetDynamicConfig()
		a.NotNil(conf)
	})
	t.Run("verify change of config", func(t *testing.T) {
		cfg := config.Config{}
		a := assert.New(t)
		dynamicConfigService := hbconfig.NewDynamicConfig(&cfg, "vault")
		a.NotNil(dynamicConfigService)
		err := dynamicConfigService.LoadConfig()
		a.NoError(err)
		var conf *config.Config
		c := hbconfig.GetDynamicConfig()
		conf = c.(*config.Config)
		time.Sleep(time.Second * 5)
		// open file
		configFileValue := `
{

  "DB": {
    "Host": "derp",
    "User": "derp",
    "Password": "derp",
    "Name": "derp",
    "Port": 1111
  }
}
`

		_, filename, _, _ := runtime.Caller(0)
		log.Println(path.Join(path.Dir(filename), "../vault") + "/config.json")

		os.WriteFile(path.Join(path.Dir(filename), "../vault")+"/config.json", []byte(configFileValue), 0644)

		time.Sleep(time.Second * 2)

		c2 := hbconfig.GetDynamicConfig()
		conf = c2.(*config.Config)
		a.Equal("derp", conf.DB.Host)
		configFileDefault := `
{

  "DB": {
    "Host": "localhost",
    "User": "dbuser",
    "Password": "dbpass",
    "Name": "derp",
    "Port": 3306
  }
}
`
		// _, filename, _, _ := runtime.Caller(0)
		os.WriteFile(path.Join(path.Dir(filename), "../vault")+"/config.json", []byte(configFileDefault), 0644)
	})

	t.Run("verify change of config and trigger", func(t *testing.T) {
		triggeredValue := "nil"
		cfg := config.Config{}
		a := assert.New(t)
		dynamicConfigService := hbconfig.NewDynamicConfig(&cfg, "vault")
		hbconfig.RegisterAutoloadCallback(func() {
			var conf *config.Config
			c := hbconfig.GetDynamicConfig()
			conf = c.(*config.Config)
			triggeredValue = conf.Reload.Trigger
		})
		a.NotNil(dynamicConfigService)
		err := dynamicConfigService.LoadConfig()
		a.NoError(err)
		var conf *config.Config
		c := hbconfig.GetDynamicConfig()
		conf = c.(*config.Config)
		time.Sleep(time.Second * 5)
		// open file
		configFileValue := `
{
  "reload": {
    "trigger": "triggered"
  }
}
`

		_, filename, _, _ := runtime.Caller(0)

		os.WriteFile(path.Join(path.Dir(filename), "../vault")+"/config.callback.json", []byte(configFileValue), 0644)

		time.Sleep(time.Second * 2)

		c2 := hbconfig.GetDynamicConfig()
		conf = c2.(*config.Config)
		a.Equal("triggered", conf.Reload.Trigger)
		a.Equal("triggered", triggeredValue)
		configFileDefault := `
{
  "reload": {
    "trigger": "NA"
  }
}
`
		// _, filename, _, _ := runtime.Caller(0)
		os.WriteFile(path.Join(path.Dir(filename), "../vault")+"/config.callback.json", []byte(configFileDefault), 0644)
	})
}
