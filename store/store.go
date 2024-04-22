package store

import (
	"fmt"

	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var ErrNotFound error = fmt.Errorf("obj not found")

func New(indexers cache.Indexers) *Store {
	return &Store{
		indexer: cache.NewIndexer(cache.MetaNamespaceKeyFunc, indexers),
	}
}

type Store struct {
	indexer cache.Indexer
}

func (s *Store) EnsureExists(obj client.Object) error {
  return s.indexer.Add(obj)
}

func (s *Store) EnsureNotExists(obj client.Object) error {
	return s.indexer.Delete(obj)
}

