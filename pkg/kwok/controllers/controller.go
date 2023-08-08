/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	clientcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/clock"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/kwok/pkg/apis/internalversion"
	"sigs.k8s.io/kwok/pkg/apis/v1alpha1"
	"sigs.k8s.io/kwok/pkg/client/clientset/versioned"
	"sigs.k8s.io/kwok/pkg/config/resources"
	"sigs.k8s.io/kwok/pkg/consts"
	"sigs.k8s.io/kwok/pkg/log"
	"sigs.k8s.io/kwok/pkg/utils/gotpl"
	"sigs.k8s.io/kwok/pkg/utils/slices"
)

var (
	startTime = time.Now().Format(time.RFC3339Nano)

	defaultFuncMap = gotpl.FuncMap{
		"Quote": func(s any) string {
			data, err := json.Marshal(s)
			if err != nil {
				return strconv.Quote(fmt.Sprint(s))
			}
			if len(data) == 0 {
				return `""`
			}
			if data[0] == '"' {
				return string(data)
			}
			return strconv.Quote(string(data))
		},
		"Now": func() string {
			return time.Now().Format(time.RFC3339Nano)
		},
		"StartTime": func() string {
			return startTime
		},
		"YAML": func(s interface{}, indent ...int) (string, error) {
			d, err := yaml.Marshal(s)
			if err != nil {
				return "", err
			}

			data := string(d)
			if len(indent) == 1 && indent[0] > 0 {
				pad := strings.Repeat(" ", indent[0]*2)
				data = strings.ReplaceAll("\n"+data, "\n", "\n"+pad)
			}
			return data, nil
		},
		"Version": func() string {
			return consts.Version
		},
	}
)

// Controller is a fake kubelet implementation that can be used to test
type Controller struct {
	conf        Config
	nodes       *NodeController
	pods        *PodController
	nodeLeases  *NodeLeaseController
	broadcaster record.EventBroadcaster
	typedClient kubernetes.Interface
}

// Config is the configuration for the controller
type Config struct {
	Clock                                 clock.Clock
	EnableCNI                             bool
	TypedClient                           kubernetes.Interface
	TypedKwokClient                       versioned.Interface
	ManageAllNodes                        bool
	ManageNodesWithAnnotationSelector     string
	ManageNodesWithLabelSelector          string
	DisregardStatusWithAnnotationSelector string
	DisregardStatusWithLabelSelector      string
	CIDR                                  string
	NodeIP                                string
	NodeName                              string
	NodePort                              int
	PodStages                             []*internalversion.Stage
	NodeStages                            []*internalversion.Stage
	PodPlayStageParallelism               uint
	NodePlayStageParallelism              uint
	NodeLeaseDurationSeconds              uint
	NodeLeaseParallelism                  uint
	ID                                    string
	EnableMetrics                         bool
	EnableNodeLeaseShareable              bool
}

// NewController creates a new fake kubelet controller
func NewController(conf Config) (*Controller, error) {
	switch {
	case conf.ManageAllNodes:
		conf.ManageNodesWithAnnotationSelector = ""
		conf.ManageNodesWithLabelSelector = ""
	case conf.ManageNodesWithAnnotationSelector != "":
	case conf.ManageNodesWithLabelSelector != "":
	default:
		return nil, fmt.Errorf("no nodes are managed")
	}

	n := &Controller{
		conf:        conf,
		broadcaster: record.NewBroadcaster(),
		typedClient: conf.TypedClient,
	}

	return n, nil
}

// Start starts the controller
func (c *Controller) Start(ctx context.Context) error {
	if c.pods != nil || c.nodes != nil || c.nodeLeases != nil {
		return fmt.Errorf("controller already started")
	}

	conf := c.conf

	recorder := c.broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: "kwok_controller"})

	var (
		nodeLeases            *NodeLeaseController
		getNodeOwnerFunc      func(nodeName string) []metav1.OwnerReference
		onLeaseNodeManageFunc func(nodeName string)
		onNodeManagedFunc     func(nodeName string)
		readOnlyFunc          func(nodeName string) bool
		nodeSelectorFunc      func(node *corev1.Node) bool
	)
	switch {
	case conf.ManageAllNodes:
		nodeSelectorFunc = func(node *corev1.Node) bool {
			return true
		}
	case conf.ManageNodesWithAnnotationSelector != "":
		selector, err := labels.Parse(conf.ManageNodesWithAnnotationSelector)
		if err != nil {
			return err
		}
		nodeSelectorFunc = func(node *corev1.Node) bool {
			return selector.Matches(labels.Set(node.Annotations))
		}
	case conf.ManageNodesWithLabelSelector != "":
		// client-go supports label filtering, so return true is ok.
		nodeSelectorFunc = func(node *corev1.Node) bool {
			return true
		}
	}

	if conf.NodeLeaseDurationSeconds != 0 {
		leaseDuration := time.Duration(conf.NodeLeaseDurationSeconds) * time.Second
		// https://github.com/kubernetes/kubernetes/blob/02f4d643eae2e225591702e1bbf432efea453a26/pkg/kubelet/kubelet.go#L199-L200
		renewInterval := leaseDuration / 4
		// https://github.com/kubernetes/component-helpers/blob/d17b6f1e84500ee7062a26f5327dc73cb3e9374a/apimachinery/lease/controller.go#L100
		renewIntervalJitter := 0.04
		l, err := NewNodeLeaseController(NodeLeaseControllerConfig{
			Clock:                conf.Clock,
			TypedClient:          conf.TypedClient,
			LeaseDurationSeconds: conf.NodeLeaseDurationSeconds,
			LeaseParallelism:     conf.NodeLeaseParallelism,
			RenewInterval:        renewInterval,
			RenewIntervalJitter:  renewIntervalJitter,
			LeaseNamespace:       corev1.NamespaceNodeLease,
			MutateLeaseFunc: setNodeOwnerFunc(func(nodeName string) []metav1.OwnerReference {
				return getNodeOwnerFunc(nodeName)
			}),
			HolderIdentity: conf.ID,
			OnNodeManagedFunc: func(nodeName string) {
				onLeaseNodeManageFunc(nodeName)
			},
			EnableShareable: conf.EnableNodeLeaseShareable,
		})
		if err != nil {
			return fmt.Errorf("failed to create node leases controller: %w", err)
		}
		nodeLeases = l

		// Not holding the lease means the node is not managed
		readOnlyFunc = func(nodeName string) bool {
			return !nodeLeases.Held(nodeName)
		}
	}

	logger := log.FromContext(ctx)

	var nodeLifecycleGetter resources.Getter[Lifecycle]
	var podLifecycleGetter resources.Getter[Lifecycle]

	if len(conf.PodStages) == 0 && len(conf.NodeStages) == 0 {
		getter := resources.NewDynamicGetter[
			[]*internalversion.Stage,
			*v1alpha1.Stage,
			*v1alpha1.StageList,
		](
			conf.TypedKwokClient.KwokV1alpha1().Stages(),
			func(objs []*v1alpha1.Stage) []*internalversion.Stage {
				return slices.FilterAndMap(objs, func(obj *v1alpha1.Stage) (*internalversion.Stage, bool) {
					r, err := internalversion.ConvertToInternalStage(obj)
					if err != nil {
						logger.Error("failed to convert to internal stage", err, "obj", obj)
						return nil, false
					}
					return r, true
				})
			},
		)

		nodeLifecycleGetter = resources.NewFilter[Lifecycle, []*internalversion.Stage](getter, func(stages []*internalversion.Stage) Lifecycle {
			lifecycle := slices.FilterAndMap(stages, func(stage *internalversion.Stage) (*LifecycleStage, bool) {
				if stage.Spec.ResourceRef.Kind != "Node" {
					return nil, false
				}

				lifecycleStage, err := NewLifecycleStage(stage)
				if err != nil {
					logger.Error("failed to create node lifecycle stage", err, "stage", stage)
					return nil, false
				}
				return lifecycleStage, true
			})
			return lifecycle
		})

		podLifecycleGetter = resources.NewFilter[Lifecycle, []*internalversion.Stage](getter, func(stages []*internalversion.Stage) Lifecycle {
			lifecycle := slices.FilterAndMap(stages, func(stage *internalversion.Stage) (*LifecycleStage, bool) {
				if stage.Spec.ResourceRef.Kind != "Pod" {
					return nil, false
				}

				lifecycleStage, err := NewLifecycleStage(stage)
				if err != nil {
					logger.Error("failed to create node lifecycle stage", err, "stage", stage)
					return nil, false
				}
				return lifecycleStage, true
			})
			return lifecycle
		})

		err := getter.Start(ctx)
		if err != nil {
			return err
		}
	} else {
		lifecycle, err := NewLifecycle(conf.PodStages)
		if err != nil {
			return fmt.Errorf("failed to create pod lifecycle: %w", err)
		}
		podLifecycleGetter = resources.NewStaticGetter(lifecycle)

		lifecycle, err = NewLifecycle(conf.NodeStages)
		if err != nil {
			return fmt.Errorf("failed to create node lifecycle: %w", err)
		}
		nodeLifecycleGetter = resources.NewStaticGetter(lifecycle)
	}

	nodes, err := NewNodeController(NodeControllerConfig{
		Clock:                                 conf.Clock,
		TypedClient:                           conf.TypedClient,
		NodeIP:                                conf.NodeIP,
		NodeName:                              conf.NodeName,
		NodePort:                              conf.NodePort,
		DisregardStatusWithAnnotationSelector: conf.DisregardStatusWithAnnotationSelector,
		DisregardStatusWithLabelSelector:      conf.DisregardStatusWithLabelSelector,
		ManageNodesWithLabelSelector:          conf.ManageNodesWithLabelSelector,
		NodeSelectorFunc:                      nodeSelectorFunc,
		OnNodeManagedFunc: func(nodeName string) {
			onNodeManagedFunc(nodeName)
		},
		Lifecycle:            nodeLifecycleGetter,
		PlayStageParallelism: conf.NodePlayStageParallelism,
		FuncMap:              defaultFuncMap,
		Recorder:             recorder,
		ReadOnlyFunc:         readOnlyFunc,
		EnableMetrics:        conf.EnableMetrics,
	})
	if err != nil {
		return fmt.Errorf("failed to create nodes controller: %w", err)
	}

	pods, err := NewPodController(PodControllerConfig{
		Clock:                                 conf.Clock,
		EnableCNI:                             conf.EnableCNI,
		TypedClient:                           conf.TypedClient,
		NodeIP:                                conf.NodeIP,
		CIDR:                                  conf.CIDR,
		DisregardStatusWithAnnotationSelector: conf.DisregardStatusWithAnnotationSelector,
		DisregardStatusWithLabelSelector:      conf.DisregardStatusWithLabelSelector,
		Lifecycle:                             podLifecycleGetter,
		PlayStageParallelism:                  conf.PodPlayStageParallelism,
		Namespace:                             corev1.NamespaceAll,
		NodeGetFunc:                           nodes.Get,
		FuncMap:                               defaultFuncMap,
		Recorder:                              recorder,
		ReadOnlyFunc:                          readOnlyFunc,
		EnableMetrics:                         conf.EnableMetrics,
	})
	if err != nil {
		return fmt.Errorf("failed to create pods controller: %w", err)
	}

	if nodeLeases != nil {
		getNodeOwnerFunc = func(nodeName string) []metav1.OwnerReference {
			nodeInfo, ok := nodes.Get(nodeName)
			if !ok || nodeInfo == nil {
				return nil
			}
			return nodeInfo.OwnerReferences
		}
		onLeaseNodeManageFunc = func(nodeName string) {
			// Manage the node and play stage all pods on the node
			nodes.Manage(nodeName)
			pods.PlayStagePodsOnNode(nodeName)
		}

		onNodeManagedFunc = func(nodeName string) {
			// Try to hold the lease
			nodeLeases.TryHold(nodeName)
		}
	} else {
		onNodeManagedFunc = func(nodeName string) {
			// Play stage all pods on the node
			pods.PlayStagePodsOnNode(nodeName)
		}
	}

	c.broadcaster.StartRecordingToSink(&clientcorev1.EventSinkImpl{Interface: c.typedClient.CoreV1().Events("")})
	if nodeLeases != nil {
		err := nodeLeases.Start(ctx)
		if err != nil {
			return fmt.Errorf("failed to start node leases controller: %w", err)
		}
	}
	err = pods.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start pods controller: %w", err)
	}
	err = nodes.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start nodes controller: %w", err)
	}

	c.pods = pods
	c.nodes = nodes
	c.nodeLeases = nodeLeases
	return nil
}

// GetNode returns the node with the given name
func (c *Controller) GetNode(nodeName string) (*NodeInfo, bool) {
	return c.nodes.Get(nodeName)
}

// ListNodes returns all nodes
func (c *Controller) ListNodes() []*NodeInfo {
	return c.nodes.List()
}

// ListPods returns all pods on the given node
func (c *Controller) ListPods(nodeName string) ([]*PodInfo, bool) {
	return c.pods.List(nodeName)
}

// Identity returns a unique identifier for this controller
func Identity() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("unable to get hostname: %w", err)
	}
	// add a uniquifier so that two processes on the same host don't accidentally both become active
	return hostname + "_" + string(uuid.NewUUID()), nil
}
