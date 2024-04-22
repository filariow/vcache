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

	Describe("ensure an object exists", func() {
		Context("adding an object", func() {
			BeforeEach(func() {
				// given
				st = store.New(nil)
			})

			It("returns an error if the object is nil", func() {
				// when
				err := st.EnsureExists(nil)

				// then: an error is returned
				Expect(err).NotTo(BeNil())

				// then: cached entries didn't change
				cmm := corev1.ConfigMapList{}
				err = st.List(ctx, &cmm)
				Expect(err).To(BeNil())
				Expect(cmm.Items).To(BeEmpty())
			})

			It("doesn't return an error if storing an empty valid object", func() {
				// when
				err := st.EnsureExists(
					&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{}})

				// then: no error is returned
				Expect(err).To(BeNil())

				// then: cache contains the added object
				rcm := corev1.ConfigMap{}
				err = st.Get(ctx, client.ObjectKey{}, &rcm)
				Expect(err).To(BeNil())
				Expect(rcm).To(BeZero())

				// then: cached entries changed
				cmm := corev1.ConfigMapList{}
				err = st.List(ctx, &cmm)
				Expect(err).To(BeNil())
				Expect(cmm.Items).To(HaveLen(1))
				Expect(cmm.Items[0]).To(BeZero())
			})

			It("doesn't return an error if the object is valid", func() {
				// when
				cm = corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "name",
						Namespace: "namespace",
					},
					Data: map[string]string{"test": "test"},
				}
				err := st.EnsureExists(&cm)

				// then: no error is returned
				Expect(err).To(BeNil())

				// then: cache contains the added object
				rcm := corev1.ConfigMap{}
				err = st.Get(ctx, client.ObjectKeyFromObject(&cm), &rcm)
				Expect(err).To(BeNil())
				Expect(rcm).To(Equal(cm))

				// then: cached entries changed
				cmm := corev1.ConfigMapList{}
				err = st.List(ctx, &cmm)
				Expect(err).To(BeNil())
				Expect(cmm.Items).To(HaveLen(1))
				Expect(cmm.Items[0]).To(Equal(cm))
			})
		})

		Context("updating an object", func() {
			BeforeEach(func() {
				// given
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
				// when
				ucm := corev1.ConfigMap{
					ObjectMeta: cm.ObjectMeta,
					Data:       map[string]string{"test": "test-update"},
				}
				err := st.EnsureExists(&ucm)

				// then: no error is returned
				Expect(err).To(BeNil())

				// then: cache contains the added object
				rcm := corev1.ConfigMap{}
				err = st.Get(ctx, client.ObjectKeyFromObject(&cm), &rcm)
				Expect(err).To(BeNil())
				Expect(rcm).To(Equal(ucm))

				// then: cached entries changed
				cmm := corev1.ConfigMapList{}
				err = st.List(ctx, &cmm)
				Expect(err).To(BeNil())
				Expect(cmm.Items).To(HaveLen(1))
				Expect(cmm.Items[0]).To(Equal(ucm))
			})
		})
	})

	Context("ensuring an object not exists", func() {
		BeforeEach(func() {
			// given
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

		It("doesn't return an error on an existing object", func() {
			// when
			err := st.EnsureNotExists(&cm)

			// then: no error is returned
			Expect(err).To(BeNil())

			// then: cache does not contain the deleted object
			rcm := corev1.ConfigMap{}
			err = st.Get(ctx, client.ObjectKeyFromObject(&cm), &rcm)
			Expect(errors.Is(err, store.ErrNotFound)).To(BeTrue())
			Expect(rcm).To(BeZero())

			// then: cached entries changed
			cmm := corev1.ConfigMapList{}
			err = st.List(ctx, &cmm)
			Expect(err).To(BeNil())
			Expect(cmm.Items).To(BeEmpty())
		})

		It("doesn't return an error on a non-existing valid object", func() {
			// when
			err := st.EnsureNotExists(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{}})

			// then: no error is returned
			Expect(err).To(BeNil())

			// then: cache does not contain the deleted object
			rcm := corev1.ConfigMap{}
			err = st.Get(ctx, client.ObjectKey{}, &rcm)
			Expect(errors.Is(err, store.ErrNotFound)).To(BeTrue())
			Expect(rcm).To(BeZero())

			// then: cached entries didn't change
			cmm := corev1.ConfigMapList{}
			err = st.List(ctx, &cmm)
			Expect(err).To(BeNil())
			Expect(cmm.Items).To(HaveLen(1))
			Expect(cmm.Items[0]).To(Equal(cm))
		})

		It("returns an error on an invalid object", func() {
			// when
			err := st.EnsureNotExists(nil)

			// then: an error is returned
			Expect(err).NotTo(BeNil())

			// then: cached entries didn't change
			cmm := corev1.ConfigMapList{}
			err = st.List(ctx, &cmm)
			Expect(err).To(BeNil())
			Expect(cmm.Items).To(HaveLen(1))
			Expect(cmm.Items[0]).To(Equal(cm))
		})
	})
})
