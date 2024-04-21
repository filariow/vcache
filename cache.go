package main

import (
	"context"
	"errors"

	"github.com/filariow/vcache/store"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Cache struct {
	client.Reader

	store   store.Store
	manager ctrl.Manager

	options *Options
}

type ReconcileFunc func(context.Context, store.Store, ctrl.Request) (ctrl.Result, error)

func New(
	mgr ctrl.Manager,
	forObject client.Object,
	reconcileFunc ReconcileFunc,
	options Option,
) (*Cache, error) {
	opts := options.ApplyToOptions(&Options{})

	b := ctrl.NewControllerManagedBy(mgr).For(forObject)
	for o, h := range opts.Watches {
		b = b.Watches(o, h)
	}

	store := store.New(opts.Indexers)
	if err := b.Complete(reconcile.Func(
		func(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
			return reconcileFunc(ctx, store, req)
		}),
	); err != nil {
		return nil, err
	}

	return &Cache{
		store:   store,
		manager: mgr,
		options: opts,
	}, nil
}

func (c *Cache) Start(ctx context.Context) error {
	return c.manager.Start(ctx)
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (c *Cache) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if err := c.store.Get(ctx, key, obj, opts...); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			gvk, err := apiutil.GVKForObject(obj, c.options.Scheme)
			if err != nil {
				return err
			}

			return kerrors.NewNotFound(schema.GroupResource{
				Group: gvk.Group,
				// Resource gets set as Kind in the error so this is fine
				Resource: gvk.Kind,
			}, key.Name)
		}
		return err
	}
	return nil
}

// List retrieves list of objects for a given namespace and list options. On a
// successful call, Items field in the list will be populated with the
// result returned from the server.
func (c *Cache) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return c.store.List(ctx, list, opts...)
}
