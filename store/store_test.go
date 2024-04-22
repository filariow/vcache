package store_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/filariow/vcache/store"
)

var _ = Describe("Store", func() {
	var ctx context.Context
	var st *store.Store
	var cm corev1.ConfigMap

	BeforeEach(func() {
		ctx = context.TODO()
	})

	Context("adding an object", func() {
		BeforeEach(func() {
			st = store.New(nil)
		})

		It("returns an error if the object is nil", func() {
			err := st.EnsureExists(nil)
			Expect(err).NotTo(BeNil())
		})

		It("doesn't return an error if storing an empty valid object", func() {
			err := st.EnsureExists(
				&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{}})
			Expect(err).To(BeNil())
		})

		It("doesn't return an error if the object is valid", func() {
			cm = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
        Data: map[string]string{"test": "test"},
			}
			err := st.EnsureExists(&cm)
			Expect(err).To(BeNil())
		})
	})

	Context("updating an object", func() {
		BeforeEach(func() {
			st = store.New(nil)
			cm = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
        Data: map[string]string{"test": "test"},
			}
			err := st.EnsureExists(&cm)
			Expect(err).To(BeNil())
		})

		It("doesn't return an error when updating an object", func() {
      ucm := corev1.ConfigMap{
        ObjectMeta: cm.ObjectMeta,
        Data: map[string]string{"test": "test-update"},
      }
      err := st.EnsureExists(&ucm)
			Expect(err).To(BeNil())

			rcm := corev1.ConfigMap{}
			err = st.Get(ctx, client.ObjectKeyFromObject(&cm), &rcm)
			Expect(err).To(BeNil())
			Expect(rcm).To(Equal(ucm))
		})
	})

	Context("deleting object", func() {
		BeforeEach(func() {
			st = store.New(nil)
			cm = corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
        Data: map[string]string{"test": "test"},
			}
			err := st.EnsureExists(&cm)
			Expect(err).To(BeNil())
		})

		It("doesn't return an error when deleting an existing object", func() {
      err := st.EnsureNotExists(&cm)
			Expect(err).To(BeNil())
      
			rcm := corev1.ConfigMap{}
			err = st.Get(ctx, client.ObjectKeyFromObject(&cm), &rcm)
			Expect(errors.Is(err, store.ErrNotFound)).To(BeTrue())
			Expect(rcm).To(BeZero())
		})
		
    It("doesn't return an error when deleting an existing object", func() {
      err := st.EnsureNotExists(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{}})
			Expect(err).To(BeNil())
      
			rcm := corev1.ConfigMap{}
			err = st.Get(ctx, client.ObjectKeyFromObject(&cm), &rcm)
			Expect(errors.Is(err, store.ErrNotFound)).To(BeFalse())
			Expect(rcm).To(Equal(cm))
		})
	})
})
