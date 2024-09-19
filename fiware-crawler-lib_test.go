package fiwarecrawlerlib_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/docker/compose/v2/pkg/api"
	fcl "github.com/scilabs-bo/fiware-crawler-lib"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	c      *fcl.Crawler
	ctx    context.Context
	cancel context.CancelFunc
	cmp    tc.ComposeStack
	iP     string
	mP     string
)

func TestMain(m *testing.M) {
	defer teardown()

	if err := createContainer(); err != nil {
		panic(err)
	}
	setEnv()
	c = fcl.New()
	code := m.Run()
	os.Exit(code)
}

func createContainer() error {
	ctx, cancel = context.WithCancel(context.Background())
	compose, err := tc.NewDockerComposeWith(tc.WithStackFiles("./testdata/docker-compose.yml"), tc.StackIdentifier("tests"))
	cmp = compose
	if err != nil {
		return err
	}

	compose.
		WaitForService("orion", wait.ForLog("msg=Startup completed")).
		WaitForService("iot-agent", wait.ForLog("$share/ul//+/+/attrs/+")).
		WaitForService("mosquitto", wait.ForListeningPort("1883/tcp")).
		WaitForService("iot-agent", wait.ForHTTP("/iot/about").WithStatusCodeMatcher(func(status int) bool { return status == 200 }).WithPort("4061/tcp")).
		Up(ctx, tc.WithRecreate(api.RecreateNever), tc.Wait(false))
	iota, err := cmp.ServiceContainer(ctx, "iot-agent")
	if err != nil {
		return err
	}
	iotaP, err := iota.MappedPort(ctx, "4061")
	if err != nil {
		return err
	}
	iP = iotaP.Port()
	mosquitto, err := cmp.ServiceContainer(ctx, "mosquitto")
	if err != nil {
		return err
	}
	mosquittoP, err := mosquitto.MappedPort(ctx, "1883")
	if err != nil {
		return err
	}
	mP = mosquittoP.Port()

	return nil

}

func TestNewServiceGroup(t *testing.T) {
	sg := c.NewConfigGroup()
	if sg.EntityType != "testType" {
		t.Error("Could not create config group")
	}
}

func TestNewDevice(t *testing.T) {
	d := c.NewDevice()
	if d.Id != "testDevice" {
		t.Error("Could not create config group")
	}
}

func TestUpsertServiceGroup(t *testing.T) {
	sg := c.NewConfigGroup()
	sg.DefaultEntityNameConjunction = ":"
	if err := c.UpsertConfigGroup(*sg); err != nil {
		t.Error(err)
	}

	if err := c.UpsertConfigGroup(*sg); err != nil {
		t.Error(err)
	}

}

func TestUpsertDevice(t *testing.T) {
	d := c.NewDevice()
	if err := c.UpsertDevice(*d); err != nil {
		t.Error(err)
	}

	if err := c.UpsertDevice(*d); err != nil {
		t.Error(err)
	}

}

func TestPublishMqtt(t *testing.T) {
	data := map[string]interface{}{"test": "test"}
	if err := c.PublishMqtt(data); err != nil {
		t.Error(err)
	}
	if err := c.PublishMqttWithDeviceId(data, "Test"); err != nil {
		t.Error(err)
	}

}

func TestJob(t *testing.T) {
	t.Log("Starting testing job")
	// This is just to stop the cron
	go func() {
		for {
			time.Sleep(time.Second * 6)
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

	os.Setenv("IOTA_PORT", iP)
	os.Setenv("SERVICE", "testservice")
	os.Setenv("SERVICE_PATH", "/test")
	os.Setenv("API_KEY", "123456")
	os.Setenv("DEVICE_ID", "testDevice")
	os.Setenv("ENTITY_TYPE", "testType")
	os.Setenv("MQTT_BROKER", "localhost")
	os.Setenv("MQTT_PORT", mP)
	os.Setenv("USERNAME", "weathercrawler")
	os.Setenv("PASSWORD", "test")
	os.Setenv("CLIENT_ID", "testClientID")
}
func teardown() {
	if cmp != nil {
		defer cmp.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal)
	}
	defer c.Iota.DeleteConfigGroup(c.Fs, c.Conf.Resource, c.Conf.ApiKey)

	defer c.Iota.DeleteDevice(c.Fs, c.Conf.DeviceId)
}
