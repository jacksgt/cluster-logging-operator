package security

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"path/filepath"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
)

type TLS bool

type TLSCertKey struct {
	CertPath string
	KeyPath  string
}

type UserNamePass struct {
	Username string
	Password string
}

type SharedKey struct {
	Key string
}

type CAFile struct {
	CAFilePath string
}

type Passphrase struct {
	KeyPass string
}

type BearerToken struct {
	Token string
}

type TLSConf struct {
	ComponentID        string
	InsecureSkipVerify bool
	TlsMinVersion      string
	CipherSuites       string
}

func NewTLSConf(o logging.OutputSpec, op generator.Options) TLSConf {
	conf := TLSConf{
		ComponentID:        helpers.FormatComponentID(o.Name),
		InsecureSkipVerify: o.TLS != nil && o.TLS.InsecureSkipVerify,
	}
	if version, found := op[generator.MinTLSVersion]; found {
		conf.TlsMinVersion = version.(string)
	}
	if ciphers, found := op[generator.Ciphers]; found {
		conf.CipherSuites = ciphers.(string)
	}
	return conf
}

func (t TLSConf) Name() string {
	return "vectorTLS"
}

func (t TLSConf) Template() string {
	return `
{{define "vectorTLS" -}}
[sinks.{{.ComponentID}}.tls]
enabled = true
{{- if ne .TlsMinVersion "" }}
min_tls_version = "{{ .TlsMinVersion }}"
{{- end }}
{{- if ne .CipherSuites "" }}
ciphersuites = "{{ .CipherSuites }}"
{{- end }}
{{- if .InsecureSkipVerify }}
verify_certificate = false
verify_hostname = false
{{- end }}
{{- end}}`
}

var NoSecrets = map[string]*corev1.Secret{}

func HasUsernamePassword(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.ClientUsername, constants.ClientPassword)
}

func HasTLSCertAndKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.ClientCertKey, constants.ClientPrivateKey)
}

func HasCABundle(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.TrustedCABundleKey)
}

func HasSharedKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.SharedKey)
}

func HasPassphrase(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.Passphrase)
}

func HasBearerTokenFileKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.BearerTokenFileKey)
}

func HasAwsRoleArnKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.AWSWebIdentityRoleKey)
}

func HasAwsCredentialsKey(secret *corev1.Secret) bool {
	return HasKeys(secret, constants.AWSCredentialsKey)
}

// GetKey if found return value and ok=true, else ok=false
func GetKey(secret *corev1.Secret, key string) (data []byte, ok bool) {
	if secret == nil {
		return nil, false
	}
	data, ok = secret.Data[key]
	return data, ok
}

// HasKeys true if all keys are present.
func HasKeys(secret *corev1.Secret, keys ...string) bool {
	for _, k := range keys {
		_, ok := GetKey(secret, k)
		if !ok {
			return false
		}
	}
	return true
}

func SecretPath(name string, file string) string {
	return fmt.Sprintf("%q", filepath.Join("/var/run/ocp-collector/secrets", name, file))
}

// TryKeys try keys in turn return data for fist one present with ok=true.
// If none present return ok=false.
func TryKeys(secret *corev1.Secret, keys ...string) (data []byte, ok bool) {
	for _, k := range keys {
		data, ok := GetKey(secret, k)
		if ok {
			return data, true
		}
	}
	return nil, false
}

func GetFromSecret(secret *corev1.Secret, name string) string {
	if secret != nil {
		return string(secret.Data[name])
	}
	return ""
}
