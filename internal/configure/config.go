package configure

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func checkErr(err error) {
	if err != nil {
		zap.S().Fatalw("config",
			"error", err,
		)
	}
}

func New() *Config {
	initLogging("info")

	config := viper.New()

	// Default config
	b, _ := json.Marshal(Config{
		ConfigFile: "config.yaml",
	})
	tmp := viper.New()
	defaultConfig := bytes.NewReader(b)

	tmp.SetConfigType("json")
	checkErr(tmp.ReadConfig(defaultConfig))
	checkErr(config.MergeConfigMap(viper.AllSettings()))

	pflag.String("config", "config.yaml", "Config file location")
	pflag.Bool("noheader", false, "Disable the startup header")

	pflag.Parse()
	checkErr(config.BindPFlags(pflag.CommandLine))

	// File
	config.SetConfigFile(config.GetString("config"))
	config.AddConfigPath(".")

	if err := config.ReadInConfig(); err == nil {
		checkErr(config.MergeInConfig())
	}

	bindEnvs(config, Config{})

	// Environment
	config.AutomaticEnv()
	config.SetEnvPrefix("CD")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AllowEmptyEnv(true)

	// Print final config
	c := &Config{}
	checkErr(config.Unmarshal(&c))

	initLogging(c.Level)

	return c
}

func bindEnvs(config *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)

	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)

		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}

		switch v.Kind() {
		case reflect.Struct:
			bindEnvs(config, v.Interface(), append(parts, tv)...)
		default:
			_ = config.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}

func BindEnvs(config *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)

	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)

		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}

		switch v.Kind() {
		case reflect.Struct:
			BindEnvs(config, v.Interface(), append(parts, tv)...)
		default:
			_ = config.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}

type MessageQueueMode string

const (
	MessageQueueModeRMQ = "RMQ"
	MessageQueueModeSQS = "SQS"
)

type Config struct {
	Level         string `mapstructure:"level" json:"level"`
	ConfigFile    string `mapstructure:"config" json:"config"`
	NoHeader      bool   `mapstructure:"noheader" json:"noheader"`
	WebsiteURL    string `mapstructure:"website_url" json:"website_url"`
	OldWebsiteURL string `mapstructure:"website_old_url" json:"website_old_url"`
	CdnURL        string `mapstructure:"cdn_url" json:"cdn_url"`

	K8S struct {
		NodeName string `mapstructure:"node_name" json:"node_name"`
		PodName  string `mapstructure:"pod_name" json:"pod_name"`
	} `mapstructure:"k8s" json:"k8s"`

	Discord struct {
		GuildID string `mapstructure:"guild_id" json:"guild_id"`
		Token   string `mapstructure:"token" json:"token"`
	} `mapstructure:"discord" json:"discord"`

	Redis struct {
		Username   string   `mapstructure:"username" json:"username"`
		Password   string   `mapstructure:"password" json:"password"`
		Database   int      `mapstructure:"db" json:"db"`
		Sentinel   bool     `mapstructure:"sentinel" json:"sentinel"`
		Addresses  []string `mapstructure:"addresses" json:"addresses"`
		MasterName string   `mapstructure:"master_name" json:"master_name"`
	} `mapstructure:"redis" json:"redis"`

	Mongo struct {
		URI    string `mapstructure:"uri" json:"uri"`
		DB     string `mapstructure:"db" json:"db"`
		Direct bool   `mapstructure:"direct" json:"direct"`
	} `mapstructure:"mongo" json:"mongo"`

	Health struct {
		Enabled bool   `mapstructure:"enabled" json:"enabled"`
		Bind    string `mapstructure:"bind" json:"bind"`
	} `mapstructure:"health" json:"health"`

	Monitoring struct {
		Enabled bool   `mapstructure:"enabled" json:"enabled"`
		Bind    string `mapstructure:"bind" json:"bind"`
		Labels  Labels `mapstructure:"labels" json:"labels"`
	} `mapstructure:"monitoring" json:"monitoring"`

	Http struct {
		Addr string `mapstructure:"addr" json:"addr"`
		Port int    `mapstructure:"port" json:"port"`
	} `mapstructure:"http" json:"http"`
}

type Labels []struct {
	Key   string `mapstructure:"key" json:"key"`
	Value string `mapstructure:"value" json:"value"`
}

func (l Labels) ToPrometheus() prometheus.Labels {
	mp := prometheus.Labels{}

	for _, v := range l {
		mp[v.Key] = v.Value
	}

	return mp
}
