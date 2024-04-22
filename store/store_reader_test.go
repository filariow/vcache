package store_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/filariow/vcache/store"
)

var _ = Describe("StoreReader", func() {
	var ctx context.Context
	var st *store.Store
	var cms []corev1.ConfigMap

	BeforeEach(func() {
		ctx = context.TODO()

		cms = []corev1.ConfigMap{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Data: map[string]string{"test": "test"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "with-labels",
					Namespace: "namespace",
					Labels: map[string]string{
						"my-label": "set",
					},
				},
				Data: map[string]string{"test": "test"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace-2",
				},
				Data: map[string]string{"test": "test"},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "with-labels-2",
					Namespace: "namespace-2",
					Labels: map[string]string{
						"my-label": "set",
					},
				},
				Data: map[string]string{"test": "test"},
			},
		}
	})

	Context("get an object", func() {
		BeforeEach(func() {
			st = store.New(nil)
			for _, cm := range cms {
				err := st.EnsureExists(&cm)
				Expect(err).To(BeNil())
			}
		})

		It("retrieves an existing object", func() {
			for _, e := range cms {
				re := corev1.ConfigMap{}
				err := st.Get(ctx, client.ObjectKeyFromObject(&e), &re)

				Expect(err).To(BeNil())
				Expect(re).To(Equal(e))
			}
		})
	})

	Describe("list objects", func() {
		Context("Without indexes", func() {
			BeforeEach(func() {
				st = store.New(nil)
				for _, cm := range cms {
					err := st.EnsureExists(&cm)
					Expect(err).To(BeNil())
				}
			})

			It("retrieves all objects", func() {
				cc := corev1.ConfigMapList{}
				err := st.List(ctx, &cc)
				Expect(err).To(BeNil())
				Expect(cc.Items).To(ConsistOf(cms))
			})
		})
	})
})

var _ = Describe("list by label", func() {
	var cms = []corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "namespace"},
			Data:       map[string]string{"test": "test"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "with-labels", Namespace: "namespace", Labels: map[string]string{"my-label": "set"}},
			Data:       map[string]string{"test": "test"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "name", Namespace: "namespace-2"},
			Data:       map[string]string{"test": "test"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "with-labels-2", Namespace: "namespace-2", Labels: map[string]string{"my-label": "set"}},
			Data:       map[string]string{"test": "test"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "with-labels-alt-2", Namespace: "namespace-2", Labels: map[string]string{"my-label": "set", "my-label-alt": "set"}},
			Data:       map[string]string{"test": "test"},
		},
	}
	var before = func() (context.Context, *store.Store) {
		st := store.New(nil)
		for _, cm := range cms {
			err := st.EnsureExists(&cm)
			Expect(err).To(BeNil())
		}
		return context.TODO(), st
	}

	DescribeTable("by label selection",
		func(expected []corev1.ConfigMap, labs string) {
			ctx, st := before()

			// given: build label selector
			rr, err := labels.ParseToRequirements(labs)
			Expect(err).To(BeNil())
			ls := labels.NewSelector()
			for _, r := range rr {
				ls = ls.Add(r)
			}

			// when
			cc := corev1.ConfigMapList{}
			err = st.List(ctx, &cc, &client.ListOptions{LabelSelector: ls})

			// then
			Expect(err).To(BeNil())
			Expect(cc.Items).NotTo(BeNil())
			Expect(cc.Items).To(ConsistOf(expected))

		},
		Entry("no labels", cms, ""),
		Entry("my-label=set", []corev1.ConfigMap{cms[1], cms[3], cms[4]}, "my-label=set"),
		Entry("my-label=aset", []corev1.ConfigMap{}, "my-label=aset"),
		Entry("my-label=set and my-label-alt=set", []corev1.ConfigMap{cms[4]}, "my-label=set,my-label-alt=set"),
		Entry("my-label-alt=set", []corev1.ConfigMap{cms[4]}, "my-label-alt=set"),
	)
})
