package reconcile_test

import (
	"context"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	core "k8s.io/api/core/v1"
	faketools "k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/google/go-cmp/cmp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("reconciling ", func() {

	var (
		caCrtStr = "initCA"

		// mimic ca.key
		caKeyStr   = "initKey"
		initSecret = runtime.NewSecret(
			"test-secret",
			"test-namespace",
			map[string][]byte{
				"testca":  []byte(caCrtStr),
				"testkey": []byte(caKeyStr),
			},
		)
	)

	var _ = DescribeTable("Secrets", func(initial, desired *core.Secret) {

		eventRecorder := &faketools.FakeRecorder{}
		k8sClient := fake.NewClientBuilder().WithRuntimeObjects(initial).Build()

		Expect(reconcile.Secret(eventRecorder, k8sClient, desired)).To(Succeed(), "Expect no error reconciling secrets")

		key := client.ObjectKeyFromObject(desired)
		act := &core.Secret{}
		Expect(k8sClient.Get(context.TODO(), key, act)).To(Succeed(), "Exp. no error after reconciliation to try and verify")

		Expect(cmp.Diff(act.Data, desired.Data)).To(BeEmpty(), "Exp. the secret data to be the same")
		Expect(cmp.Diff(act.Data, initial.Data)).To(Not(BeEmpty()), "Exp. the secret data have been updated")
	},
		Entry("when it does not exist", &core.Secret{}, runtime.NewSecret(
			"test-secret",
			"test-namespace",
			map[string][]byte{
				"testca":  []byte(caCrtStr),
				"testkey": []byte(caKeyStr),
			},
		)),
		Entry("when values are updated for the existing keys", runtime.NewSecret(
			"test-secret",
			"test-namespace",
			map[string][]byte{
				"testca":  []byte("abc"),
				"testkey": []byte("123"),
			},
		), runtime.NewSecret(
			"test-secret",
			"test-namespace",
			map[string][]byte{
				"testca":  []byte(caCrtStr),
				"testkey": []byte(caKeyStr),
			},
		)),
		Entry("when keys are added to the data",
			initSecret,
			runtime.NewSecret(
				"test-secret",
				"test-namespace",
				map[string][]byte{
					"testca":  []byte(caCrtStr),
					"testkey": []byte(caKeyStr),
					"newkey":  []byte("abc"),
				},
			)),
		Entry("when keys are removed from the data",
			initSecret,
			runtime.NewSecret(
				"test-secret",
				"test-namespace",
				map[string][]byte{
					"testca": []byte(caCrtStr),
				},
			)),
	)
})
