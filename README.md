# KubeRTAS - RTAS event notifier for Kubernetes on ppc64le.


KubeRTAS reads the servicelog.db to read the RTAS (Run Time Abstraction Services) events which have been reported by the RTAS daemon. The events are related to the platform, such as Power Suppy or Fan Failures, etc. and are reported to the Kubernetes API server. 

The KubeRTAS application can be configured to work in a standalone mode, or be deployed as a DaemonSet in Kubernetes cluster. This gives additional visibility to the cluster administrators regarding the issues that may be prevelent in the underlying infrastructure.

---
#### Architecture Diagram:

![RTAS](https://github.com/kishen-v/kube-rtas/assets/110517346/bbbd9436-c87f-4757-a4f3-5c9c9c8a8835)

---
### Configuration:

```
{
    "PollInterval": 5,
    "ServicelogDBPath": "servicelog.db",
    "Severity": 3,
    "InCluster": false,
    "Verbosity": 1
}
```

PollInterval(Hours) - Scan the servicelog.db in specified intervals of time.

ServicelogDBPath - Path to the servicelog.db file on the node.

Severity - Retrieve entries from the servicelog.db whose event severity is greater than or equal to the set value.

InCluster - Run the KubeRTAS application either as an Deployment or a standalone application.

Verbosity - 1 to enable debug level logs.

---

### Usage Standalone mode
#### Standalone mode

This mode requires the KUBECONFIG environment variable to be set.

```
yum install sqlite-devel

cd cmd
go build -o kubertas .
./kubertas -c <path to config file (optional)>
```

This creates a kubeRTAS process and uses the available Kubeconfig to notify the RTAS events to Kubernetes API server.

#### Daemonset deployment
` kubectl apply -f deployment `

This creates the required RBAC, config map and the daemonset to notify RTAS events to the Kubernetes API server.
