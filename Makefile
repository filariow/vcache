GO ?= go
GINKGO ?= ginkgo
KUBECTL ?= kubectl
E2E_KUBECONFIG ?= /tmp/test-e2e-kubeconfig
E2E_CLUSTER_NAME ?= vcache-e2e

ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
LOCALBIN ?= $(ROOT_DIR)/bin
$(LOCALBIN):
	mkdir $(LOCALBIN)

KWOKCTL ?= $(LOCALBIN)/kwokctl

.PHONY: test
test:
	$(GINKGO) run --label-filter '!e2e' ./...

.PHONY: test-e2e
test-e2e: e2e-env-delete e2e-env-prepare e2e-run-tests e2e-env-delete
	@:

.PHONY: e2e-env-prepare
e2e-env-prepare: kwokctl
	KUBECONFIG=$(E2E_KUBECONFIG) $(KWOKCTL) create cluster --name '$(E2E_CLUSTER_NAME)'

.PHONY: e2e-install-crds
e2e-install-crds:
	KUBECONFIG=$(E2E_KUBECONFIG) $(KUBECTL) apply -k ./test/testdata/vconfigmap/config/crd/

.PHONY: e2e-run-tests
e2e-run-tests: e2e-install-crds
	KUBECONFIG=$(E2E_KUBECONFIG) $(GINKGO) run --label-filter 'e2e' ./...
	
.PHONY: e2e-env-delete
e2e-env-delete: kwokctl
	-KUBECONFIG=$(E2E_KUBECONFIG) $(KWOKCTL) delete cluster --name '$(E2E_CLUSTER_NAME)'

.PHONY: kwokctl
kwokctl: $(KWOKCTL)
$(KWOKCTL): $(LOCALBIN)
	$(GO) build \
		-modfile $(ROOT_DIR)/hack/tools/kwokctl/go.mod \
		-o $(LOCALBIN)/kwokctl \
		sigs.k8s.io/kwok/cmd/kwokctl
