package main

import (
	"github.com/moira-alert/moira/cmd"
	"github.com/moira-alert/moira/notifier"
	"github.com/moira-alert/moira/notifier/selfstate"
	"menteslibres.net/gosexy/to"
)

type config struct {
	Redis    cmd.RedisConfig    `yaml:"redis"`
	Graphite cmd.GraphiteConfig `yaml:"graphite"`
	Logger   cmd.LoggerConfig   `yaml:"log"`
	Notifier notifierConfig     `yaml:"notifier"`
}

type notifierConfig struct {
	SenderTimeout    string              `yaml:"sender_timeout"`
	ResendingTimeout string              `yaml:"resending_timeout"`
	Senders          []map[string]string `yaml:"senders"`
	SelfState        selfStateConfig     `yaml:"moira_selfstate"`
	FrontURI         string              `yaml:"front_uri"`
}

type selfStateConfig struct {
	Enabled                 string              `yaml:"enabled"`
	RedisDisconnectDelay    int64               `yaml:"redis_disconect_delay"`
	LastMetricReceivedDelay int64               `yaml:"last_metric_received_delay"`
	LastCheckDelay          int64               `yaml:"last_check_delay"`
	Contacts                []map[string]string `yaml:"contacts"`
	NoticeInterval          int64               `yaml:"notice_interval"`
}

func getDefault() config {
	return config{
		Redis: cmd.RedisConfig{
			Host: "localhost",
			Port: "6379",
			DBID: 0,
		},
		Graphite: cmd.GraphiteConfig{
			URI:      "localhost:2003",
			Prefix:   "DevOps.Moira",
			Interval: "60s0ms",
		},
		Logger: cmd.LoggerConfig{
			LogFile:  "stdout",
			LogLevel: "debug",
		},
		Notifier: notifierConfig{
			SenderTimeout:    "10s0ms",
			ResendingTimeout: "24:00",
			SelfState: selfStateConfig{
				Enabled:                 "false",
				RedisDisconnectDelay:    30,
				LastMetricReceivedDelay: 60,
				LastCheckDelay:          60,
				NoticeInterval:          300,
			},
			FrontURI: "http:// localhost",
		},
	}
}

func (config *notifierConfig) getSettings() notifier.Config {
	return notifier.Config{
		SendingTimeout:   to.Duration(config.SenderTimeout),
		ResendingTimeout: to.Duration(config.ResendingTimeout),
		Senders:          config.Senders,
		FrontURL:         config.FrontURI,
	}
}

func (config *selfStateConfig) getSettings() selfstate.Config {
	return selfstate.Config{
		Enabled:                 cmd.ToBool(config.Enabled),
		RedisDisconnectDelay:    config.RedisDisconnectDelay,
		LastMetricReceivedDelay: config.LastMetricReceivedDelay,
		LastCheckDelay:          config.LastCheckDelay,
		Contacts:                config.Contacts,
		NoticeInterval:          config.NoticeInterval,
	}
}
