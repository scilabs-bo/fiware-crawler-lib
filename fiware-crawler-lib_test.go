package fiwarecrawlerlib_test

import (
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	fcl "github.com/scilabs-bo/fiware-crawler-lib"
)

var (
	c *fcl.Crawler
)

func TestMain(m *testing.M) {
	setEnv()
	c = fcl.New()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func TestNewServiceGroup(t *testing.T) {
	sg := c.NewServiceGroup()
	if sg.EntityType != "testType" {
		t.Fail()
	}
}

func TestNewDevice(t *testing.T) {
	d := c.NewDevice()
	if d.Id != "testDevice" {
		t.Fail()
	}
}

func TestUpsertServiceGroup(t *testing.T) {
	sg := c.NewServiceGroup()
	sg.DefaultEntityNameConjunction = ":"
	c.UpsertServiceGroup(*sg)
	c.UpsertServiceGroup(*sg)
}

func TestUpsertDevice(t *testing.T) {
	d := c.NewDevice()
	c.UpsertDevice(*d)
	c.UpsertDevice(*d)
}

func TestPublishMqtt(t *testing.T) {
	data := map[string]interface{}{"test": "test"}
	c.PublishMqtt(data)
}

func TestJob(t *testing.T) {
  // This is just to stop the cron
  go func() {
		for {
      time.Sleep(time.Second * 6 )
      c.Cron.Stop()
		}
	}()
	c.Cron.LimitRunsTo(3)
  c.Cron.StopBlockingChan()
	c.StartJob(func() {
		data := map[string]interface{}{"test": "test"}
		c.PublishMqtt(data)
	})
}

func setEnv() {
	os.Setenv("CRONTAB", "*/2 * * * * *")
	os.Setenv("IOTA_HOST", "localhost")
	os.Setenv("IOTA_PORT", "4061")
	os.Setenv("SERVICE", "testservice")
	os.Setenv("SERVICE_PATH", "/test")
	os.Setenv("API_KEY", "123456")
	os.Setenv("DEVICE_ID", "testDevice")
	os.Setenv("ENTITY_TYPE", "testType")
	os.Setenv("MQTT_BROKER", "localhost")
	os.Setenv("CLIENT_ID", "testClientID")
}
func teardown() {
	err := c.Iota.DeleteServiceGroup(c.Fs, c.Conf.Resource, c.Conf.ApiKey)
	if err != nil {
		log.Error().Err(err).Send()
	}
	c.Iota.DeleteDevice(c.Fs, c.Conf.DeviceId)
}
