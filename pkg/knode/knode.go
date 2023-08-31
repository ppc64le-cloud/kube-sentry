package knode

import (
	"errors"
	"fmt"
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

var (
	eventType string
)

type Notifier struct {
	nodeName string
	config   *rest.Config
	recorder record.EventRecorder
}

func NewNotifier() *Notifier {
	return &Notifier{}
}

// InitializeK8sNotifier sets up the event recorder to send events to the Kube API
// server.
func (k *Notifier) InitializeNotifier() error {
	var err error
	if k.config, err = config.GetConfig(); err != nil {
		return fmt.Errorf("cannot retrieve rest config for kubernetes. %v", err)
	}
	// Proceed to fetch the NODE_NAME set through the downward API.
	k.nodeName = os.Getenv("NODE_NAME")
	// If the node name is unset, obtain the hostname as it may be running in standalone
	// mode.
	if k.nodeName == "" {
		klog.Info("NODE_NAME is unset, assuming standalone mode of execution.")
		k.nodeName, err = os.Hostname()
		if err != nil || k.nodeName == "" {
			return errors.New("hostname is unavailable, node events cannot be reported")
		}
	}

	client := kubernetes.NewForConfigOrDie(k.config).CoreV1()
	k.recorder = getEventRecorder(client, k.nodeName, "RTASNotifier")
	return nil
}

func getEventRecorder(c typedcorev1.CoreV1Interface, nodeName, source string) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	recorder := eventBroadcaster.NewRecorder(legacyscheme.Scheme, v1.EventSource{Component: source, Host: nodeName})
	// The namespace to which the events need to be posted can be configured from c.Events(namespace)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: c.Events("")})
	return recorder
}

// NotifyAPIServer posts events to the Kubernetes API server.
func (k *Notifier) NotifyAPIServer(message string, severity uint8) {
	// RTAS events, whose severity is lesser than 4 (DEBUG,INFO and EVENT) are mapped to
	// Kubernetes' EventTypeNormal, while the rest are mapped to EventTypeWarning
	if severity < 4 {
		eventType = v1.EventTypeNormal
	} else {
		eventType = v1.EventTypeWarning
	}
	k.recorder.Event(&v1.ObjectReference{
		Kind: "Node",
		Name: k.nodeName,
		UID:  types.UID(k.nodeName),
	}, eventType, "PlatformEvent", message)
}
