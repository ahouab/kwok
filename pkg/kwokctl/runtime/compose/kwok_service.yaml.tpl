apiVersion: v1
kind: Service
metadata:
  labels:
    k8s-app: kwok-controller
  name: kwok-controller
  namespace: kube-system
spec:
  externalName: {{ .Name }}-kwok-controller
  type: ExternalName
