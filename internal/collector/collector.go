package collector

import (
	"path"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	vector "github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	coreFactory "github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	clusterLoggingPriorityClassName = "system-node-critical"
	ExporterPort                    = int32(2112)
	ExporterPortName                = "logfile-metrics"
	MetricsPort                     = int32(24231)
	MetricsPortName                 = "metrics"
	metricsVolumeName               = "collector-metrics"
	logContainers                   = "varlogcontainers"
	logContainersValue              = "/var/log/containers"
	logPods                         = "varlogpods"
	logPodsValue                    = "/var/log/pods"
	logJournal                      = "varlogjournal"
	logJournalValue                 = "/var/log/journal"
	logAudit                        = "varlogaudit"
	logAuditValue                   = "/var/log/audit"
	logOvn                          = "varlogovn"
	logOvnValue                     = "/var/log/ovn"
	logOauthapiserver               = "varlogoauthapiserver"
	logOauthapiserverValue          = "/var/log/oauth-apiserver"
	logOauthserver                  = "varlogoauthserver"
	logOauthserverValue             = "/var/log/oauth-server"
	logOpenshiftapiserver           = "varlogopenshiftapiserver"
	logOpenshiftapiserverValue      = "/var/log/openshift-apiserver"
	logKubeapiserver                = "varlogkubeapiserver"
	logKubeapiserverValue           = "/var/log/kube-apiserver"
	metricsVolumePath               = "/etc/collector/metrics"
	tmpVolumeName                   = "tmp"
	tmpPath                         = "/tmp"
)

var (
	defaultTolerations = []v1.Toleration{
		{
			Key:      "node-role.kubernetes.io/master",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
		{
			Key:      "node.kubernetes.io/disk-pressure",
			Operator: v1.TolerationOpExists,
			Effect:   v1.TaintEffectNoSchedule,
		},
	}
)

type Visitor func(collector *v1.Container, podSpec *v1.PodSpec)

type Factory struct {
	ConfigHash    string
	CollectorSpec logging.CollectionSpec
	CollectorType logging.LogCollectionType
	ClusterID     string
	ImageName     string
	TrustedCAHash string
	Visit         Visitor
	Secrets       map[string]*v1.Secret
	ForwarderSpec logging.ClusterLogForwarderSpec
}

// CollectorResourceRequirements returns the resource requirements for a given collector implementation
// or it's default if none are specified
func (f *Factory) CollectorResourceRequirements() v1.ResourceRequirements {
	if f.CollectorSpec.Resources == nil {
		if f.CollectorType == logging.LogCollectionTypeVector {
			return v1.ResourceRequirements{}
		}
		return v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: fluentd.DefaultMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: fluentd.DefaultMemory,
				v1.ResourceCPU:    fluentd.DefaultCpuRequest,
			},
		}
	}
	return *f.CollectorSpec.Resources
}

func (f *Factory) NodeSelector() map[string]string {
	return f.CollectorSpec.CollectorSpec.NodeSelector
}
func (f *Factory) Tolerations() []v1.Toleration {
	return f.CollectorSpec.CollectorSpec.Tolerations
}

func New(confHash, clusterID string, collectorSpec logging.CollectionSpec, secrets map[string]*v1.Secret, forwarderSpec logging.ClusterLogForwarderSpec) *Factory {
	factory := &Factory{
		ClusterID:     clusterID,
		ConfigHash:    confHash,
		CollectorSpec: collectorSpec,
		CollectorType: collectorSpec.Type,
		ImageName:     constants.FluentdName,
		Visit:         fluentd.CollectorVisitor,
		Secrets:       secrets,
		ForwarderSpec: forwarderSpec,
	}
	if collectorSpec.Type == logging.LogCollectionTypeVector {
		factory.ImageName = constants.VectorName
		factory.Visit = vector.CollectorVisitor
	}
	return factory
}

func (f *Factory) NewDaemonSet(namespace, name string, trustedCABundle *v1.ConfigMap, tlsProfileSpec configv1.TLSProfileSpec) *apps.DaemonSet {
	podSpec := f.NewPodSpec(trustedCABundle, f.ForwarderSpec, f.ClusterID, f.TrustedCAHash, tlsProfileSpec)
	ds := coreFactory.NewDaemonSet(name, namespace, constants.CollectorName, constants.CollectorName, *podSpec)
	return ds
}

func (f *Factory) NewPodSpec(trustedCABundle *v1.ConfigMap, forwarderSpec logging.ClusterLogForwarderSpec, clusterID, trustedCAHash string, tlsProfileSpec configv1.TLSProfileSpec) *v1.PodSpec {

	podSpec := &v1.PodSpec{
		NodeSelector:                  utils.EnsureLinuxNodeSelector(f.NodeSelector()),
		PriorityClassName:             clusterLoggingPriorityClassName,
		ServiceAccountName:            constants.CollectorServiceAccountName,
		TerminationGracePeriodSeconds: utils.GetInt64(10),
		Tolerations:                   defaultTolerations,
		Volumes: []v1.Volume{
			{Name: logContainers, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logContainersValue}}},
			{Name: logPods, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logPodsValue}}},
			{Name: logJournal, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logJournalValue}}},
			{Name: logAudit, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logAuditValue}}},
			{Name: logOvn, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOvnValue}}},
			{Name: logOauthapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOauthapiserverValue}}},
			{Name: logOauthserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOauthserverValue}}},
			{Name: logOpenshiftapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logOpenshiftapiserverValue}}},
			{Name: logKubeapiserver, VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: logKubeapiserverValue}}},
			{Name: metricsVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: constants.CollectorMetricSecretName}}},
			{Name: tmpVolumeName, VolumeSource: v1.VolumeSource{EmptyDir: &v1.EmptyDirVolumeSource{Medium: v1.StorageMediumMemory}}},
		},
	}
	podSpec.Tolerations = append(podSpec.Tolerations, f.Tolerations()...)

	secretNames := AddSecretVolumes(podSpec, forwarderSpec)

	exporter := newLogMetricsExporterContainer(tlsProfileSpec)
	collector := f.NewCollectorContainer(secretNames, clusterID)

	addTrustedCABundle(collector, podSpec, trustedCABundle)

	f.Visit(collector, podSpec)

	addWebIdentityForCloudwatch(collector, podSpec, forwarderSpec, f.Secrets)

	podSpec.Containers = []v1.Container{
		*collector,
		*exporter,
	}

	return podSpec
}

// NewCollectorContainer is a constructor for creating the collector container spec.  Note the secretNames are assumed
// to be a unique list
func (f *Factory) NewCollectorContainer(secretNames []string, clusterID string) *v1.Container {

	collector := factory.NewContainer(constants.CollectorName, f.ImageName, v1.PullIfNotPresent, f.CollectorResourceRequirements())
	collector.Ports = []v1.ContainerPort{
		{
			Name:          MetricsPortName,
			ContainerPort: MetricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}
	collector.Env = []v1.EnvVar{
		{Name: "COLLECTOR_CONF_HASH", Value: f.ConfigHash},
		{Name: common.TrustedCABundleHashName, Value: f.TrustedCAHash},
		{Name: "K8S_NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "spec.nodeName"}}},
		{Name: "NODE_IPV4", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"}}},
		{Name: "OPENSHIFT_CLUSTER_ID", Value: clusterID},
		{Name: "POD_IP", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
		{Name: "POD_IPS", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIPs"}}},
	}
	collector.Env = append(collector.Env, utils.GetProxyEnvVars()...)

	collector.VolumeMounts = []v1.VolumeMount{
		{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
		{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
		{Name: logJournal, ReadOnly: true, MountPath: logJournalValue},
		{Name: logAudit, ReadOnly: true, MountPath: logAuditValue},
		{Name: logOvn, ReadOnly: true, MountPath: logOvnValue},
		{Name: logOauthapiserver, ReadOnly: true, MountPath: logOauthapiserverValue},
		{Name: logOauthserver, ReadOnly: true, MountPath: logOauthserverValue},
		{Name: logOpenshiftapiserver, ReadOnly: true, MountPath: logOpenshiftapiserverValue},
		{Name: logKubeapiserver, ReadOnly: true, MountPath: logKubeapiserverValue},
		{Name: metricsVolumeName, ReadOnly: true, MountPath: metricsVolumePath},
		{Name: tmpVolumeName, MountPath: tmpPath},
	}
	// List of _unique_ output secret names, several outputs may use the same secret.
	AddSecretVolumeMounts(&collector, secretNames)

	addSecurityContextTo(&collector)
	return &collector
}

func newLogMetricsExporterContainer(tlsProfileSpec configv1.TLSProfileSpec) *v1.Container {
	// deliberately not passing any resources for running the below container process, let it have cpu and memory as the process requires
	exporterResources := &v1.ResourceRequirements{}
	exporter := factory.NewContainer(constants.LogfilesmetricexporterName, constants.LogfilesmetricexporterName, v1.PullIfNotPresent, *exporterResources)
	exporter.Ports = []v1.ContainerPort{
		{
			Name:          ExporterPortName,
			ContainerPort: ExporterPort,
			Protocol:      v1.ProtocolTCP,
		},
	}
	exporter.Command = []string{"/bin/bash"}
	exporter.Args = []string{"-c",
		"/usr/local/bin/log-file-metric-exporter -verbosity=2 -dir=/var/log/pods -http=:2112 -keyFile=/etc/collector/metrics/tls.key -crtFile=/etc/collector/metrics/tls.crt -tlsMinVersion=" +
			tls.MinTLSVersion(tlsProfileSpec) + " -cipherSuites=" + strings.Join(tls.TLSCiphers(tlsProfileSpec), ",")}

	exporter.VolumeMounts = []v1.VolumeMount{
		{Name: logContainers, ReadOnly: true, MountPath: logContainersValue},
		{Name: logPods, ReadOnly: true, MountPath: logPodsValue},
		{Name: metricsVolumeName, ReadOnly: true, MountPath: metricsVolumePath},
	}

	addSecurityContextTo(&exporter)
	return &exporter
}

// AddSecretVolumeMounts to the collector container
func AddSecretVolumeMounts(collector *v1.Container, secretNames []string) {
	// List of _unique_ output secret names, several outputs may use the same secret.
	for _, name := range secretNames {
		path := path.Join(constants.CollectorSecretsDir, name)
		collector.VolumeMounts = append(collector.VolumeMounts, v1.VolumeMount{Name: name, ReadOnly: true, MountPath: path})
	}
}

// AddSecretVolumes adds secret volumes to the pod spec for the unique set of pipeline secrets and returns the list of
// the secret names
func AddSecretVolumes(podSpec *v1.PodSpec, pipelineSpec logging.ClusterLogForwarderSpec) []string {
	// List of _unique_ output secret names, several outputs may use the same secret.
	unique := sets.NewString()
	for _, o := range pipelineSpec.Outputs {
		if o.Secret != nil && o.Secret.Name != "" {
			unique.Insert(o.Secret.Name)
		}
	}
	secretNames := unique.List()
	for _, name := range secretNames {
		podSpec.Volumes = append(podSpec.Volumes, v1.Volume{Name: name, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: name}}})
	}
	return secretNames
}

func addSecurityContextTo(container *v1.Container) *v1.Container {
	container.SecurityContext = &v1.SecurityContext{
		Capabilities: &v1.Capabilities{
			Drop: RequiredDropCapabilities,
		},
		SELinuxOptions: &v1.SELinuxOptions{
			Type: "spc_t",
		},
		ReadOnlyRootFilesystem:   utils.GetBool(true),
		AllowPrivilegeEscalation: utils.GetBool(false),
		SeccompProfile: &v1.SeccompProfile{
			Type: v1.SeccompProfileTypeRuntimeDefault,
		},
	}
	return container
}

func addTrustedCABundle(collector *v1.Container, podSpec *v1.PodSpec, trustedCABundleCM *v1.ConfigMap) {
	if trustedCABundleCM != nil && hasTrustedCABundle(trustedCABundleCM) {
		collector.VolumeMounts = append(collector.VolumeMounts,
			v1.VolumeMount{
				Name:      constants.CollectorTrustedCAName,
				ReadOnly:  true,
				MountPath: constants.TrustedCABundleMountDir,
			})

		podSpec.Volumes = append(podSpec.Volumes,
			v1.Volume{
				Name: constants.CollectorTrustedCAName,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: constants.CollectorTrustedCAName,
						},
						Items: []v1.KeyToPath{
							{
								Key:  constants.TrustedCABundleKey,
								Path: constants.TrustedCABundleMountFile,
							},
						},
					},
				},
			})
		//TODO add trusted ca hash to env vars
	}
}

func hasTrustedCABundle(configMap *v1.ConfigMap) bool {
	if configMap == nil {
		return false
	}
	caBundle, ok := configMap.Data[constants.TrustedCABundleKey]
	return ok && caBundle != ""
}
