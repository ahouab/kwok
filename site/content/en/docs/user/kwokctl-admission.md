# `kwokctl` Admission

{{< hint "info" >}}

This document walks you known about kubernetes admission in kwokctl.

{{< /hint >}}

## What is Admission

Kubernetes provides a mechanism called admission controller to intercept requests to the Kubernetes API server prior to persistence of the object, but after the request is authenticated and authorized.
There are two special controllers: `MutatingAdmissionWebhook` and `ValidatingAdmissionWebhook`.
These two controllers call out to a webhook service to do some processing.

## Enable Admission

Before the release of v0.3.0, on Kind runtime is enabled by default, but on other runtimes, you need to enable it manually.
If you want to enable kube authorization, you need with `--kube-admission` flag when creating cluster.

Starting from v0.3.0, Admission is enabled by default.
If you want to disable kube authorization, you need with `--kube-admission=false` flag when creating cluster.

If you are creating a cluster with kube version < `1.21`, then [authorization] also needs to be enabled.

## Webhook Configuration

### In Binary

All components run using local binary, you need to use `127.0.0.1` to access the webhook service on the host.

### In Docker/Podman/Nerdctl

All components run in containers, you need to use the service name of the network to access the webhook service on the container.

### In Kind

All components run in kind cluster, just like general kubernetes cluster.

[authorization]: {{< relref "/docs/user/kwokctl-authorization" >}}
