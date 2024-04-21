package store

import (
	"fmt"

	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrNotFound error = fmt.Errorf("obj not found")

type Store interface {
	client.Reader

	EnsureExists(client.Object) error
	EnsureNotExists(client.Object) error
}

func New(indexers cache.Indexers) *store {
	return &store{
		indexer: cache.NewIndexer(cache.MetaNamespaceKeyFunc, indexers),
	}
}

type store struct {
	indexer cache.Indexer
}

func (s *store) EnsureExists(obj client.Object) error {
	return s.indexer.Add(obj)
}

func (s *store) EnsureNotExists(obj client.Object) error {
	return s.indexer.Delete(obj)
}
