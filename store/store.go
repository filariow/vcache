package store

import (
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(indexers cache.Indexers) *Store {
	return &Store{
		indexer: cache.NewIndexer(cache.MetaNamespaceKeyFunc, indexers),
	}
}

type Store struct {
	indexer cache.Indexer
}

func (s *Store) EnsureExists(obj client.Object) error {
	return s.indexer.Add(obj.DeepCopyObject())
}

func (s *Store) EnsureNotExists(obj client.Object) error {
	return s.indexer.Delete(obj)
}
