apiVersion: apiregistration.k8s.io/v1
kind: APIService
metadata:
  name: v1beta1.external.metrics.k8s.io
spec:
  group: external.metrics.k8s.io
  groupPriorityMinimum: 100
  insecureSkipTLSVerify: true
  service:
    name: kwok-controller
    namespace: kube-system
    port: {{ .KwokControllerPort }}
  version: v1beta1
  versionPriority: 100
