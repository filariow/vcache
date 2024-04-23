package cache_test

import (
	"context"
	"log"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/filariow/vcache/cache"
	vconfigmapv1alpha1 "github.com/filariow/vcache/test/testdata/vconfigmap/api/v1alpha1"
)

var _ = Describe("Cache", Label("e2e"), func() {
	var c *cache.Cache
	var cfg *rest.Config
	var scheme *runtime.Scheme
	var ctx context.Context
	var cli client.Client
	var namespaceName string

	BeforeEach(func() {
		ctx = context.TODO()

		apiConfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
		Expect(err).To(BeNil())

		cfg, err = clientcmd.NewDefaultClientConfig(*apiConfig, nil).ClientConfig()
		Expect(err).To(BeNil())

		cli, err = client.New(cfg, client.Options{Scheme: scheme})
		Expect(err).To(BeNil())

		namespaceName = "test"
		err = cli.Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}})
		Expect(err).To(BeNil())

		scheme = runtime.NewScheme()
		err = corev1.AddToScheme(scheme)
		Expect(err).To(BeNil())
		err = vconfigmapv1alpha1.AddToScheme(scheme)
		Expect(err).To(BeNil())

		opts := ctrl.Options{Scheme: scheme}
		mgr, err := ctrl.NewManager(cfg, opts)
		Expect(err).To(BeNil())

		r := func(ctx context.Context, c *cache.Cache, r ctrl.Request) (ctrl.Result, error) {
			log.Printf("reconciling %v", r)
			cli := c.Client()
			vcm := vconfigmapv1alpha1.VirtualConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      r.Name,
					Namespace: r.Namespace,
				},
			}

			cm := corev1.ConfigMap{}
			if err := cli.Get(ctx, r.NamespacedName, &cm); err != nil {
				if errors.IsNotFound(err) {
					log.Printf("configmap not found, deleting %T %s/%s from cache", vcm, vcm.Namespace, vcm.Name)
					return ctrl.Result{}, c.Store().EnsureNotExists(&vcm)
				}
				return ctrl.Result{}, err
			}

			vcm.Spec.Data = cm.Data
			if err := c.Store().EnsureExists(&vcm); err != nil {
				return ctrl.Result{}, err
			}
			log.Printf("%T %s/%s persisted in store", vcm, vcm.Namespace, vcm.Name)
			return ctrl.Result{}, nil
		}

		ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{Development: true})))

		c, err = cache.New(mgr, &vconfigmapv1alpha1.VirtualConfigMap{}, r, &cache.Options{
			Scheme: scheme,
			Watches: map[client.Object]handler.EventHandler{
				&corev1.ConfigMap{}: handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, o client.Object) []reconcile.Request {
					log.Printf("converting ConfigMap %s/%s to VConfigMap event", o.GetNamespace(), o.GetName())
					return []reconcile.Request{
						{
							NamespacedName: types.NamespacedName{
								Namespace: o.GetNamespace(),
								Name:      o.GetName(),
							},
						},
					}
				}),
			},
		})
		Expect(err).To(BeNil())

		go func() {
			c.Start(ctx)
		}()
	})

	AfterEach(func() {
		ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}
		err := cli.Get(ctx, client.ObjectKeyFromObject(&ns), &ns)
		Expect(err).To(BeNil())

		err = cli.Delete(ctx, &ns)
		Expect(err).To(BeNil())
	})

	When("a secret is created", func() {
		It("creates a virtual configmap", func() {
			cm := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-configmap",
					Namespace: "test",
				},
				Data: map[string]string{"key": "test"},
			}
			err := cli.Create(ctx, &cm)
			Expect(err).To(BeNil())

			vcm := vconfigmapv1alpha1.VirtualConfigMap{}
			err = wait.PollUntilContextTimeout(ctx, 1*time.Second, 1*time.Minute, true, func(ctx context.Context) (done bool, err error) {
				if err := c.Get(ctx, client.ObjectKeyFromObject(&cm), &vcm); err != nil {
					return false, nil
				}
				return true, nil
			})
			Expect(err).To(BeNil())
			Expect(vcm).NotTo(BeZero())
			Expect(vcm.ObjectMeta.Name).To(Equal(cm.ObjectMeta.Name))
			Expect(vcm.ObjectMeta.Namespace).To(Equal(cm.ObjectMeta.Namespace))
			Expect(vcm.Spec.Data).To(Equal(cm.Data))
		})
	})
})
