package machineremediation

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	mrv1 "kubevirt.io/machine-remediation-operator/pkg/apis/machineremediation/v1alpha1"
	mrotesting "kubevirt.io/machine-remediation-operator/pkg/utils/testing"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func init() {
	// Add types to scheme
	mrv1.AddToScheme(scheme.Scheme)
}

type FakeRemedatior struct {
}

func (fr *FakeRemedatior) Recreate(context.Context, *mrv1.MachineRemediation) error {
	return nil
}

func (fr *FakeRemedatior) Reboot(context.Context, *mrv1.MachineRemediation) error {
	return nil
}

// newFakeReconciler returns a new reconcile.Reconciler with a fake client
func newFakeReconciler(initObjects ...runtime.Object) *ReconcileMachineRemediation {
	fakeClient := fake.NewFakeClient(initObjects...)
	remediator := &FakeRemedatior{}
	return &ReconcileMachineRemediation{
		client:     fakeClient,
		remediator: remediator,
		namespace:  mrotesting.NamespaceTest,
	}
}

type expectedReconcile struct {
	result reconcile.Result
	error  bool
}

func TestReconcile(t *testing.T) {
	machineRemediationStarted := mrotesting.NewMachineRemediation("machineRemediationStarted", "", mrv1.RemediationTypeReboot, mrv1.RemediationStateStarted)
	machineRemediationPoweroff := mrotesting.NewMachineRemediation("machineRemediationPoweroff", "", mrv1.RemediationTypeRecreate, mrv1.RemediationStatePowerOff)
	machineRemediationPoweron := mrotesting.NewMachineRemediation("machineRemediationPoweron", "", mrv1.RemediationTypeReboot, mrv1.RemediationStatePowerOn)
	machineRemediationSucceeded := mrotesting.NewMachineRemediation("machineRemediationSucceeded", "", mrv1.RemediationTypeRecreate, mrv1.RemediationStateSucceeded)
	machineRemediationFailed := mrotesting.NewMachineRemediation("machineRemediationFailed", "", mrv1.RemediationTypeReboot, mrv1.RemediationStateFailed)

	testsCases := []struct {
		machineRemediation *mrv1.MachineRemediation
		expected           expectedReconcile
	}{
		{
			machineRemediation: machineRemediationStarted,
			expected: expectedReconcile{
				result: reconcile.Result{
					Requeue:      true,
					RequeueAfter: 10 * time.Second,
				},
				error: false,
			},
		},
		{
			machineRemediation: machineRemediationPoweroff,
			expected: expectedReconcile{
				result: reconcile.Result{
					Requeue:      true,
					RequeueAfter: 10 * time.Second,
				},
				error: false,
			},
		},
		{
			machineRemediation: machineRemediationPoweron,
			expected: expectedReconcile{
				result: reconcile.Result{
					Requeue:      true,
					RequeueAfter: 10 * time.Second,
				},
				error: false,
			},
		},
		{
			machineRemediation: machineRemediationSucceeded,
			expected: expectedReconcile{
				result: reconcile.Result{},
				error:  false,
			},
		},
		{
			machineRemediation: machineRemediationFailed,
			expected: expectedReconcile{
				result: reconcile.Result{},
				error:  false,
			},
		},
	}

	r := newFakeReconciler(
		machineRemediationStarted,
		machineRemediationPoweroff,
		machineRemediationPoweron,
		machineRemediationSucceeded,
		machineRemediationFailed,
	)

	for _, tc := range testsCases {
		request := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: mrotesting.NamespaceTest,
				Name:      tc.machineRemediation.Name,
			},
		}
		result, err := r.Reconcile(request)
		if tc.expected.error != (err != nil) {
			var errorExpectation string
			if !tc.expected.error {
				errorExpectation = "no"
			}
			t.Errorf("Test case: %s. Expected: %s error, got: %v", tc.machineRemediation.Name, errorExpectation, err)
		}

		if result != tc.expected.result {
			t.Errorf("Test case: %s. Expected: %v, got: %v", tc.machineRemediation.Name, tc.expected.result, result)
		}
	}
}
