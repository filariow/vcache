package store_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/filariow/vcache/store"
)

var _ = Describe("StoreReader", func() {
	var ctx context.Context
	var st *store.Store
	var cms []corev1.ConfigMap

	BeforeEach(func() {
		ctx = context.TODO()
		st = store.New(nil)

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

		for _, cm := range cms {
			err := st.EnsureExists(&cm)
			Expect(err).To(BeNil())
		}
	})

	Context("get an object", func() {
		It("retrieves an existing object", func() {
			for _, e := range cms {
				re := corev1.ConfigMap{}
				err := st.Get(ctx, client.ObjectKeyFromObject(&e), &re)

				Expect(err).To(BeNil())
				Expect(re).To(Equal(e))
			}
		})
	})

	Context("list objects", func() {
		It("retrieves all objects", func() {
			cc := corev1.ConfigMapList{}
			err := st.List(ctx, &cc)
			Expect(err).To(BeNil())
			Expect(cc.Items).To(BeEquivalentTo(cms))
		})
	})
})
