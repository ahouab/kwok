## kwok

kwok is a tool for simulating the lifecycle of fake nodes, pods, and other Kubernetes API resources.

```
kwok [flags]
```

### Options

```
      --cidr string                                        CIDR of the pod ip (default "10.0.0.1/24")
  -c, --config strings                                     config path (default [~/.kwok/kwok.yaml])
      --disregard-status-with-annotation-selector string   All node/pod status excluding the ones that match the annotation selector will be watched and managed.
      --disregard-status-with-label-selector string        All node/pod status excluding the ones that match the label selector will be watched and managed.
      --enable-crd strings                                 List of CRDs to enable
      --enable-node-lease-shareable                        Enable node lease shareable, means that the controller will share the node lease with others
      --experimental-enable-cni                            Experimental support for getting pod ip from CNI, for CNI-related components, Only works with Linux
  -h, --help                                               help for kwok
      --kubeconfig string                                  Path to the kubeconfig file to use (default "~/.kube/config")
      --manage-all-nodes                                   All nodes will be watched and managed. It's conflicted with manage-nodes-with-annotation-selector and manage-nodes-with-label-selector.
      --manage-nodes-with-annotation-selector string       Nodes that match the annotation selector will be watched and managed. It's conflicted with manage-all-nodes.
      --manage-nodes-with-label-selector string            Nodes that match the label selector will be watched and managed. It's conflicted with manage-all-nodes.
      --master string                                      The address of the Kubernetes API server (overrides any value in kubeconfig).
      --node-ip string                                     IP of the node
      --node-lease-duration-seconds uint                   Duration of node lease seconds
      --node-name string                                   Name of the node
      --node-port int                                      Port of the node
      --server-address string                              Address to expose the server on
      --tls-cert-file string                               File containing the default x509 Certificate for HTTPS
      --tls-private-key-file string                        File containing the default x509 private key matching --tls-cert-file
  -v, --v log-level                                        number for the log level verbosity (DEBUG, INFO, WARN, ERROR) or (-4, 0, 4, 8) (default INFO)
```

