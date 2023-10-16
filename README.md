# Kube-sentry: Kubernetes - Servicelog Event Notifier Towards improved Reliability.

kube-sentry reads the servicelog.db to parse the RTAS (Run Time Abstraction Services) events which have been reported by the RTAS daemon. The events are related to the platform, such as Predictive CPU Failure, IO Bus Failure, Fan Failures, etc. and are reported to the Kubernetes API server. 

This gives the cluster administrators visiblity about the errors that may be prevalent in underlying host. The servicelogs are specific to the PowerPC family of servers by IBM Corporation. 

The kube-sentry application can be configured to work in a standalone mode, or be deployed as a DaemonSet in Kubernetes cluster.

---
#### Architecture Diagram:

![kube-sentry](https://github.com/ppc64le-cloud/kube-sentry/assets/110517346/4f010c90-664d-40f0-b694-15c3744dbefc)


---
### Configuration:

```
{
    "PollInterval": 5,
    "ServicelogDBPath": "servicelog.db",
    "Severity": 3
}
```

PollInterval(Hours) - Scan the servicelog.db in specified intervals of time.

ServicelogDBPath - Path to the servicelog.db file on the node.

Severity - Retrieve entries from the servicelog.db whose event severity is greater than or equal to the set value.

---

### Usage Standalone mode
#### Standalone mode

This mode requires the KUBECONFIG environment variable to be set.

```
yum install sqlite-devel

cd cmd
go build -o kube-sentry .
./kube-sentry -c <path to config file (optional) -v <log verbosity>
```

This creates a kube-sentry process and uses the available Kubeconfig to notify the RTAS events to Kubernetes API server.

#### Daemonset deployment
` kubectl apply -f deployment `

This creates the required RBAC, config map and the daemonset to notify RTAS events to the Kubernetes API server.
