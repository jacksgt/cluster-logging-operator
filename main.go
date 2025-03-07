package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"
	"k8s.io/apimachinery/pkg/api/errors"
	"os"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/apis"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/version"

	apiruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	configv1 "github.com/openshift/api/config/v1"
	consolev1 "github.com/openshift/api/console/v1"
	consolev1alpha1 "github.com/openshift/api/console/v1alpha1"
	oauth "github.com/openshift/api/oauth/v1"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	"github.com/openshift/cluster-logging-operator/controllers/clusterlogging"
	"github.com/openshift/cluster-logging-operator/controllers/forwarding"
	loggingruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// Change below variables to serve metrics on different host or port.
var (
	scheme   = apiruntime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	UnHealthyStatus = "0"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(apis.AddToScheme(scheme))
	utilruntime.Must(elasticsearch.AddToScheme(scheme))
	utilruntime.Must(routev1.AddToScheme(scheme))
	utilruntime.Must(consolev1.AddToScheme(scheme))
	utilruntime.Must(consolev1alpha1.AddToScheme(scheme))
	utilruntime.Must(oauth.AddToScheme(scheme))
	utilruntime.Must(monitoringv1.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
	utilruntime.Must(securityv1.AddToScheme(scheme))

	utilruntime.Must(loggingv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8686", "The address the metric endpoint binds to.")
	//flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe end point binds to.")

	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	utils.InitLogger("cluster-logging-operator")
	log.Info("starting up...",
		"operator_version", version.Version,
		"go_version", runtime.Version(),
		"go_os", runtime.GOOS,
		"go_arch", runtime.GOARCH,
	)

	// https://issues.redhat.com/browse/LOG-3321
	syncPeriod := time.Minute * 3
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Namespace:              constants.WatchNamespace,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "b430cc2e.openshift.io",
		SyncPeriod:             &syncPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	clusterVersion, err := getClusterVersion(mgr.GetAPIReader())
	if err != nil {
		setupLog.Error(err, "unable to retrieve the clusterID")
		os.Exit(1)
	}
	clusterID := string(clusterVersion.Spec.ClusterID)

	migrateManifestResources(mgr.GetClient())

	log.Info("Registering Components.")

	if err = (&clusterlogging.ReconcileClusterLogging{
		Client:         mgr.GetClient(),
		Reader:         mgr.GetAPIReader(),
		Scheme:         mgr.GetScheme(),
		Recorder:       mgr.GetEventRecorderFor("clusterlogging-controller"),
		ClusterVersion: clusterVersion.Status.Desired.Version,
		ClusterID:      clusterID,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterLogForwarder")
		telemetry.Data.CLInfo.M["healthStatus"] = UnHealthyStatus
		os.Exit(1)
	}
	if err = (&forwarding.ReconcileForwarder{
		Client:         mgr.GetClient(),
		Scheme:         mgr.GetScheme(),
		Recorder:       mgr.GetEventRecorderFor("clusterlogforwarder"),
		ClusterVersion: clusterVersion.Status.Desired.Version,
		ClusterID:      clusterID,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterLogging")
		telemetry.Data.CLFInfo.M["healthStatus"] = UnHealthyStatus
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	// updating clo Telemetry Data - to be published by prometheus
	cloversion, err := getCLOVersion()
	if err != nil {
		cloversion = version.Version
		log.Info("Failed to get clo version from env variable OPERATOR_CONDITION_NAME so falling back to default version")
	}
	telemetry.Data.CLInfo.M["version"] = cloversion

	errr := telemetry.RegisterMetrics()
	if errr != nil {
		log.Error(err, "Error in registering clo metrics for telemetry")
	}

	log.Info("Starting the Cmd.")
	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Manager exited non-zero")
		os.Exit(1)
	}

}

func migrateManifestResources(k8sClient client.Client) {
	log.Info("migrating resources provided by the manifest")
	if err := k8sClient.Delete(context.TODO(), loggingruntime.NewPriorityClass("cluster-logging", 0, false, "")); err != nil && !errors.IsNotFound(err) {
		log.V(1).Error(err, "There was an error trying to remove the old collector PriorityClass named 'cluster-logging'")
	}
}

// TODO - Re-enable when CLO can be installed in alternate namespace
//// getWatchNamespace get the namespace name of the scoped operator
//// - https://sdk.operatorframework.io/docs/building-operators/golang/operator-scope/#configuring-namespace-scoped-operators
//func getWatchNamespace() (string, error) {
//	watchNamespaceEnvVar := "WATCH_NAMESPACE"
//	ns, found := os.LookupEnv(watchNamespaceEnvVar)
//	if !found {
//		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
//	}
//	return ns, nil
//}

// get clo operator version from CLUSTER_OPERATOR_CONDITION ENV variable .. supported OCP 4.8 version onwards
func getCLOVersion() (string, error) {
	CLOVersionEnvVar := "OPERATOR_CONDITION_NAME"
	cloversion, found := os.LookupEnv(CLOVersionEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", CLOVersionEnvVar)
	}
	return cloversion, nil
}

// getClusterVersion retrieves the ID of the cluster
func getClusterVersion(k8client client.Reader) (*configv1.ClusterVersion, error) {
	clusterVersion := &configv1.ClusterVersion{}
	key := client.ObjectKey{Name: "version"}
	if err := k8client.Get(context.TODO(), key, clusterVersion); err != nil {
		return nil, err
	}
	return clusterVersion, nil
}
