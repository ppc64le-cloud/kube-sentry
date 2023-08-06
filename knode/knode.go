package knode

import (
	"errors"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/api/legacyscheme"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var Kubeconfig *string

type K8sNotifier struct {
	nodeName string
	config   *rest.Config
	recorder record.EventRecorder
}

func NewK8sNotifier() *K8sNotifier {
	return &K8sNotifier{}
}

// InitializeK8sNotifier sets up the event recorder to send events to the Kube API
// server.
func (k *K8sNotifier) InitializeK8sNotifier() error {
	var err error
	if k.config, err = config.GetConfig(); err != nil {
		klog.Fatalf("cannot retrieve rest config for kubernetes. %v ", err)
	}
	// Proceed to fetch the NODE_NAME set through the downward API.
	k.nodeName = os.Getenv("NODE_NAME")
	// If the node name is still unset, obtain the hostname as it may be running in standalone
	// mode.
	if k.nodeName == "" {
		k.nodeName, err = os.Hostname()
		if err != nil || k.nodeName == "" {
			return errors.New("hostname is unavailable, node events cannot be reported")
		}
	}

	client := kubernetes.NewForConfigOrDie(k.config).CoreV1()
	k.recorder = getEventRecorder(client, "", k.nodeName, "RTASNotifier")
	return nil
}

func getEventRecorder(c typedcorev1.CoreV1Interface, namespace, nodeName, source string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(legacyscheme.Scheme, v1.EventSource{Component: source, Host: nodeName})
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: c.Events(namespace)})
	return recorder
}

// NotifyAPIServer posts events to the Kubernetes API server.
func (k *K8sNotifier) NotifyAPIServer(message string) {
	k.recorder.Event(&v1.ObjectReference{
		Kind: "Node",
		Name: k.nodeName,
		UID:  types.UID(k.nodeName),
	}, v1.EventTypeWarning, "PlatformEvent", message)
}
