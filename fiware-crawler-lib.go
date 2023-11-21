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

type Config struct {
	Crontab string `env:"CRONTAB,required=true"`

	IotAHost string `env:"IOTA_HOST,required=true"`
	IotAPort int    `env:"IOTA_PORT,required=true"`

	Service     string `env:"SERVICE,required=true"`
	ServicePath string `env:"SERVICE_PATH,required=true"`

	ApiKey   i.Apikey   `env:"API_KEY,required=true"`
	Resource i.Resource `env:"RESOURCE,default=/iot/d"`

	DeviceId   i.DeciveId `env:"DEVICE_ID, required=true"`
	EntityType string     `env:"ENTITY_TYPE, required=true"`

	LogLevel string `env:"LOG_LEVEL,default=DEBUG"`

	MqttBroker string `env:"MQTT_BROKER,default=mosquitto"`
	MqttPort   int    `env:"MQTT_PORT,default=1883"`
	ClientId   string `env:"CLIENT_ID,required=true"`
	Username   string `env:"USERNAME,required=false"`
	Password   string `env:"PASSWORD,required=false"`
}

type Crawler struct{
	Conf Config
	Iota i.IoTA
	Fs   i.FiwareService
	Cron *gocron.Scheduler
}

func New()*Crawler {
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

func (c *Crawler) StartJob(jobFunc interface{}){
  c.Cron.Do(jobFunc)
  c.Cron.StartBlocking()
}

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

func (c *Crawler) NewServiceGroup() *i.ServiceGroup {
	sg := &i.ServiceGroup{Apikey: c.Conf.ApiKey, Resource: c.Conf.Resource, EntityType: c.Conf.EntityType}
	return sg
}

func (c *Crawler) NewDevice() *i.Device {
	d := &i.Device{Id: c.Conf.DeviceId, Transport: "MQTT"}
	return d
}

func (c *Crawler) UpsertServiceGroup(sg i.ServiceGroup) {
	ensureServiceGroupExists(c.Iota, c.Fs, sg)
}

func (c *Crawler) UpsertDevice(d i.Device) {
	ensureDeviceExists(c.Iota, c.Fs, d)
}

func ensureServiceGroupExists(ia i.IoTA, fs i.FiwareService, sg i.ServiceGroup) {
	exists := ia.ServiceGroupExists(fs, sg.Resource, sg.Apikey)
	if !exists {
		log.Debug().Msg("Creating service group...")
		err := ia.CreateServiceGroup(fs, sg)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not create service group")
		}
	} else {
		log.Debug().Msg("Update service group...")
		err := ia.UpdateServiceGroup(fs, sg.Resource, sg.Apikey, sg)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not update service group")
		}
	}
}

func ensureDeviceExists(ia i.IoTA, fs i.FiwareService, d i.Device) {

	exists := ia.DeviceExists(fs, d.Id)
	if !exists {
		log.Debug().Msg("Creating device...")
		err := ia.CreateDevice(fs, d)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not create device")
		}
	} else {
		log.Debug().Msg("Update device...")
		dTmp, err := ia.ReadDevice(fs, d.Id)
		if err != nil || dTmp.EntityName == "" {
			log.Fatal().Err(err).Msg("Can not update device, no entity_name")
		}

		d.Transport = ""
		d.EntityName = dTmp.EntityName
		err = ia.UpdateDevice(fs, d)
		if err != nil {
			log.Fatal().Err(err).Msg("Could not update device")
		}
	}

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
