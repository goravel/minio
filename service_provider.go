package minio

import (
	"context"

	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "goravel.minio"

var App foundation.Application

type ServiceProvider struct {
}

func (r *ServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings: []string{
			Binding,
		},
		Dependencies: []string{
			binding.Config,
		},
		ProvideFor: []string{
			binding.Storage,
		},
	}
}

func (r *ServiceProvider) Register(app foundation.Application) {
	App = app

	app.BindWith(Binding, func(app foundation.Application, parameters map[string]any) (any, error) {
		return NewMinio(context.Background(), app.MakeConfig(), parameters["disk"].(string))
	})
}

func (r *ServiceProvider) Boot(app foundation.Application) {
	app.Publishes("github.com/goravel/minio", map[string]string{
		"config/minio.go": app.ConfigPath(""),
	})
}
