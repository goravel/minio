package minio

import (
	"context"

	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.minio"

var App foundation.Application

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.BindWith(Binding, func(app foundation.Application, parameters map[string]any) (any, error) {
		return NewMinio(context.Background(), app.MakeConfig(), parameters["disk"].(string))
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	app.Publishes("github.com/goravel/minio", map[string]string{
		"config/minio.go": app.ConfigPath(""),
	})
}
