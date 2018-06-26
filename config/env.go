package config

import "os"

type EnvConfig struct {
	DatabasePath string
	ThreadListURL string
	ThreadBaseURL string
	SlackToken string
	SlackChannel string
	SlackBotID string
	ThreadNameContains string
}

var config *EnvConfig

func init() {
	config = &EnvConfig{
		DatabasePath: os.Getenv("DATABASE_PATH"),
		ThreadListURL: os.Getenv("THREAD_LIST_URL"),
		ThreadBaseURL: os.Getenv("THREAD_BASE_URL"),
		SlackToken: os.Getenv("SLACK_TOKEN"),
		SlackChannel: os.Getenv("SLACK_CHANNEL"),
		SlackBotID: os.Getenv("SLACK_BOT_ID"),
		ThreadNameContains: os.Getenv("THREAD_NAME_CONTAINS"),
	}
}

func GetEnvConfig() *EnvConfig {
	return config
}
