package cache

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

type Options struct {
	Indexers cache.Indexers

	Watches map[client.Object]handler.EventHandler

	Scheme *runtime.Scheme
}

type Option interface {
	ApplyToOptions(*Options) *Options
}

func (o *Options) ApplyToOptions(opts *Options) *Options {
	if o.Indexers == nil {
		o.Indexers = cache.Indexers{}
	}
	if o.Watches == nil {
		o.Watches = map[client.Object]handler.EventHandler{}
	}
	if o.Scheme == nil {
		o.Scheme = runtime.NewScheme()
	}
	return o
}
