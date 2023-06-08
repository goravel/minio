package config

import (
	"github.com/goravel/framework/facades"
)

func init() {
	config := facades.Config()
	config.Add("minio", map[string]any{
		"key":      config.Env("MINIO_ACCESS_KEY_ID"),
		"secret":   config.Env("MINIO_ACCESS_KEY_SECRET"),
		"bucket":   config.Env("MINIO_BUCKET"),
		"region":   config.Env("MINIO_REGION"),
		"url":      config.Env("MINIO_URL"),
		"ssl":      config.Env("MINIO_SSL"),
		"endpoint": config.Env("MINIO_ENDPOINT"),
	})
}
