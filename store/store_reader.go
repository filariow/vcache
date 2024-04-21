package store

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (s *store) Get(ctx context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) error {
	o, exists, err := s.indexer.Get(key)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}

	obj = o.(runtime.Object).DeepCopyObject().(client.Object)
	return nil
}

// List retrieves list of objects for a given namespace and list options. On a
// successful call, Items field in the list will be populated with the
// result returned from the server.
func (s *store) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)

	// list the objects from the store, applying field selectors
	objs, err := s.listWithFieldSelector(listOpts)
	if err != nil {
		return err
	}

	// filter by label and limit
	fobjs, err := filterByLabelSelectorsWithLimit(objs, listOpts.LabelSelector, listOpts.Limit)
	if err != nil {
		return err
	}

	// deepCopy to runtime objects
	robjs, err := deepCopyToRuntimeObject(fobjs)
	if err != nil {
		return err
	}

	// populate result list
	return apimeta.SetList(list, robjs)
}

func (s *store) listWithFieldSelector(listOpts client.ListOptions) ([]interface{}, error) {
	switch {
	case listOpts.FieldSelector != nil:
		return s.listByIndexes(listOpts.FieldSelector)
	case listOpts.Namespace != "":
		return s.indexer.ByIndex(cache.NamespaceIndex, listOpts.Namespace)
	default:
		return s.indexer.List(), nil
	}
}

func (s *store) listByIndexes(fieldSelector fields.Selector) ([]interface{}, error) {
	rr := fieldSelector.Requirements()
	if len(rr) == 0 {
		return s.indexer.List(), nil
	}

	// check only exact matches are used in field selector
	for _, req := range rr {
		if req.Operator != selection.Equals && req.Operator != selection.DoubleEquals {
			return nil, fmt.Errorf("only exact field matches (equal or double equal) are supported")
		}
	}

	// apply first filter
	r := rr[0]
	objs, err := s.indexer.ByIndex(r.Field, r.Value)
	if err != nil {
		return nil, err
	}
	if len(objs) == 0 {
		return nil, nil
	}

	// apply all other indexers
	indexers := s.indexer.GetIndexers()
	for _, r := range rr[1:] {
		if len(objs) == 0 {
			return nil, nil
		}

		// if index does not exist on field, break
		fn, ok := indexers[r.Field]
		if !ok {
			return nil, fmt.Errorf("index not found: %q", r.Field)
		}

		// apply requirement
		fobjs := make([]interface{}, 0, len(objs))
		for _, obj := range objs {
			vals, err := fn(obj)
			if err != nil {
				return nil, err
			}

			for _, val := range vals {
				if val == r.Value {
					fobjs = append(fobjs, obj)
					break
				}
			}
		}
		objs = fobjs
	}

	return objs, nil
}

func filterByLabelSelectorsWithLimit(objs []interface{}, labelSelector labels.Selector, limit int64) ([]interface{}, error) {
	// empty input
	if len(objs) == 0 {
		return nil, nil
	}

	// calculate the maximum number of result we need to provide.
	// Also used to allocate the minimum amount of memory for storing results
	maxResult, unlimited := func() (int64, bool) {
		if lobjs := int64(len(objs)); limit == 0 || limit >= lobjs {
			return lobjs, true
		}
		return limit, false
	}()

	// check edge cases
	if labelSelector == nil {
		if unlimited {
			// no label selection nor limiting, return the input
			return objs, nil
		} else {
			// no label selection to apply, just limiting
			return objs[:maxResult], nil
		}
	}

	// prepare the buffer for the result
	fobjs := make([]interface{}, 0, maxResult)

	// loop on all the input objects applying
	for idx, obj := range objs {
		if int64(idx) > maxResult {
			break
		}

		// extract meta from object
		m, err := meta.Accessor(obj)
		if err != nil {
			return nil, err
		}

		// apply label selection
		lbls := labels.Set(m.GetLabels())
		if labelSelector.Matches(lbls) {
			fobjs = append(fobjs, obj)
		}
	}

	return fobjs, nil
}

func deepCopyToRuntimeObject(objs []interface{}) ([]runtime.Object, error) {
	robjs := make([]runtime.Object, len(objs))
	for idx, obj := range objs {
		robj, ok := obj.(runtime.Object)
		if !ok {
			return nil, fmt.Errorf("error casting %T to runtime.Object", obj)
		}

		robjs[idx] = robj.DeepCopyObject()
	}
	return robjs, nil
}
