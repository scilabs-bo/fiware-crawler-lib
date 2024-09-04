// Package fiwarecrawlerlib provides utility functions for regularly crawling data
// and creating configuration groups and devices for a FIWARE IoT-Agent UL (Ultralight).
//
// The main functionalities include configuring a crawler with scheduling, handling
// MQTT communication, and managing FIWARE service groups and devices. It is designed
// to simplify the integration of IoT devices with FIWARE IoT Agents using Ultralight
// protocol for data communication.
package fiwarecrawlerlib

import (
	"fmt"
	"strings"
	"time"

	env "github.com/Netflix/go-env"
	i "github.com/fbuedding/iota-admin/pkg/iot-agent-sdk"
	"github.com/go-co-op/gocron"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Config represents the configuration parameters for the fiwarecrawlerlib package.
// It is populated using environment variables.
type Config struct {
	Crontab string `env:"CRONTAB,required=true"`

	IotAHost string `env:"IOTA_HOST,required=true"`
	IotAPort int    `env:"IOTA_PORT,required=true"`

	Service     string `env:"SERVICE,required=true"`
	ServicePath string `env:"SERVICE_PATH,required=true"`

	ApiKey   i.Apikey   `env:"API_KEY,required=true"`
	Resource i.Resource `env:"RESOURCE,default=/iot/d"`

	DeviceId   i.DeciveId `env:"DEVICE_ID, required=false"`
	EntityType string     `env:"ENTITY_TYPE, required=true"`

	LogLevel string `env:"LOG_LEVEL,default=DEBUG"`

	MqttBroker string `env:"MQTT_BROKER,default=mosquitto"`
	MqttPort   int    `env:"MQTT_PORT,default=1883"`
	ClientId   string `env:"CLIENT_ID,required=true"`
	Username   string `env:"USERNAME,required=false"`
	Password   string `env:"PASSWORD,required=false"`
}

// Crawler represents the main structure for the fiwarecrawlerlib package,
// encapsulating the configuration, FIWARE IoT-Agent, FiwareService, and the cron scheduler.
type Crawler struct {
	Conf Config
	Iota i.IoTA
	Fs   i.FiwareService
	Cron *gocron.Scheduler
}

// New creates a new Crawler instance and initializes it with the configuration
// parameters from environment variables.
// Return a pointer to the Crawler instance
func New() *Crawler {
	c := &Crawler{}
	_, err := env.UnmarshalFromEnviron(&c.Conf)
	if err != nil {
		log.Fatal().Err(err).Msg("Error while reading envs")
	}
	setLogLevel(c.Conf.LogLevel)
	c.Iota = i.IoTA{Host: c.Conf.IotAHost, Port: c.Conf.IotAPort}
	c.Fs = i.FiwareService{Service: c.Conf.Service, ServicePath: c.Conf.ServicePath}
	c.Cron = gocron.NewScheduler(time.Local)
	c.Cron = c.Cron.CronWithSeconds(c.Conf.Crontab)
	return c
}

// StartJob takes a function jobFunc, adds it to the cron scheduler and starts it blocking.
func (c *Crawler) StartJob(jobFunc interface{}) {
	c.Cron.Do(jobFunc)
	c.Cron.StartBlocking()
}

// PublishMqtt publishes data to an MQTT broker based on the configuration parameters.
// data items will be joind to a ul payload, nested struct will be parsed to json.
func (c *Crawler) PublishMqtt(data map[string]interface{}) {
	payloadArr := make([]string, 0)
	for k, v := range data {
		payloadArr = append(payloadArr, fmt.Sprintf("%v|%v", k, v))
	}
	payload := strings.Join(payloadArr, "|")
	log.Debug().Str("payload", payload).Msg("Publishing payload...")
	options := mqtt.NewClientOptions().AddBroker("mqtt://" + c.Conf.MqttBroker + ":" + fmt.Sprint(c.Conf.MqttPort))
	options.ClientID = c.Conf.ClientId
	if c.Conf.Username != "" {
		options.Username = c.Conf.Username
	}
	if c.Conf.Password != "" {
		options.Password = c.Conf.Password
	}
	client := mqtt.NewClient(options)
	t := client.Connect()
	_ = t.Wait()
	if t.Error() != nil {
		log.Error().Err(t.Error()).Msg("Error connecting")
		return
	}
	if c.Conf.DeviceId == "" {
		log.Error().Msg("For publishing without device id a device id has to be set")
		return
	}
	topic := fmt.Sprintf("/ul/%v/%v/attrs", c.Conf.ApiKey, c.Conf.DeviceId)
	log.Debug().Str("topic", topic).Msg("Publishing on topic")
	t = client.Publish(topic, options.WillQos, false, payload)
	_ = t.Wait()
	if t.Error() != nil {
		log.Error().Err(t.Error()).Msg("Error publishing message")
	}
	client.Disconnect(100)
	log.Info().Msg("Published data")
}

func (c *Crawler) PublishMqttWithDeviceId(data map[string]interface{}, deviceId i.DeciveId) {
	payloadArr := make([]string, 0)
	for k, v := range data {
		payloadArr = append(payloadArr, fmt.Sprintf("%v|%v", k, v))
	}
	payload := strings.Join(payloadArr, "|")
	log.Debug().Str("payload", payload).Msg("Publishing payload...")
	options := mqtt.NewClientOptions().AddBroker("mqtt://" + c.Conf.MqttBroker + ":" + fmt.Sprint(c.Conf.MqttPort))
	options.ClientID = c.Conf.ClientId
	if c.Conf.Username != "" {
		options.Username = c.Conf.Username
	}
	if c.Conf.Password != "" {
		options.Password = c.Conf.Password
	}
	client := mqtt.NewClient(options)
	t := client.Connect()
	_ = t.Wait()
	if t.Error() != nil {
		log.Error().Err(t.Error()).Msg("Error connecting")
		return
	}
	if deviceId == "" {
		log.Error().Msg("Device id can not be empty")
		return
	}
	topic := fmt.Sprintf("/ul/%v/%v/attrs", c.Conf.ApiKey, deviceId)
	log.Debug().Str("topic", topic).Msg("Publishing on topic")
	t = client.Publish(topic, options.WillQos, false, payload)
	_ = t.Wait()
	if t.Error() != nil {
		log.Error().Err(t.Error()).Msg("Error publishing message")
	}
	client.Disconnect(100)
	log.Info().Msg("Published data")
}

// NewConfigGroup creates a new ConfigGroup instance based on the configuration.
// It returns a pointer to the newly created ConfigGroup.
func (c *Crawler) NewConfigGroup() *i.ConfigGroup {
	sg := &i.ConfigGroup{Apikey: c.Conf.ApiKey, Resource: c.Conf.Resource, EntityType: c.Conf.EntityType}
	return sg
}

// NewDevice creates a new Device instance based on the configuration.
// It returns a pointer to the newly created Device.
func (c *Crawler) NewDevice() *i.Device {
	d := &i.Device{Id: c.Conf.DeviceId, Transport: "MQTT"}
	return d
}

// UpsertConfigGroup ensures the existence of a service group in the FIWARE IoT-Agent.
// It takes a ConfigGroup sg as input and returns no values.
func (c *Crawler) UpsertConfigGroup(cg i.ConfigGroup) {
	c.Iota.UpsertConfigGroup(c.Fs, cg)
}

// UpsertDevice ensures the existence of a device in the FIWARE IoT-Agent.
// It takes a Device d as input and returns no values.
func (c *Crawler) UpsertDevice(d i.Device) {
	c.Iota.UpsertDevice(c.Fs, d)
}

func setLogLevel(ll string) {
	ll = strings.ToLower(ll)
	switch ll {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		log.Fatal().Msg("Log level need to be one of this: [TRACE DEBUG INFO WARNING ERROR FATAL PANIC]")
	}
}
