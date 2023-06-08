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

	app.Bind(Binding, func(app foundation.Application) (any, error) {
		return NewMinio(context.Background(), app.MakeConfig())
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	app.Publishes("github.com/goravel/minio", map[string]string{
		"config/minio.go": app.ConfigPath(""),
	})
}
