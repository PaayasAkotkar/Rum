package rum

// todo

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"strings"

// 	rumapp "k8s.io/api/apps/v1"
// 	crumapp "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	corev1 "k8s.io/kubernetes/pkg/apis/core"

// 	"k8s.io/client-go/kubernetes"
// 	"k8s.io/client-go/rest"
// )

// const (
// 	group   = "rum.onecolor.io"
// 	version = "v1"
// )

// type RumCluster struct {
// 	client    *kubernetes.Clientset
// 	crdClient *rumapp.ControllerRevisionList
// 	image     string
// 	namespace string
// }

// type IReplicaManager struct {
// 	XXl   bool
// 	Large bool
// 	Mid   bool
// 	XS    bool
// 	Small bool
// }

// func rumReplicaManager(req IReplicaManager) int {
// 	switch true {
// 	case req.XXl:
// 		return 5
// 	case req.Large:
// 		return 4
// 	case req.Mid:
// 		return 3
// 	case req.Small:
// 		return 2
// 	case req.XS:
// 		return 1
// 	default:
// 		return 0
// 	}
// }

// // IRumClusterComponent implements the field struct for controlling the struct
// type IRumClusterComponent struct {
// 	Name       string // name of the service
// 	ObjectName string

// 	// kind -> spec ...
// 	Spec map[string]IRumClusterComponentSpec
// }

// // not per-function replicas
// // but per-component (pod) replicas + its CRD schema
// type IRumClusterComponentSpec struct {
// 	Replicas   int    // how many pods of this component
// 	Served     bool   // is this CRD version active
// 	Storage    bool   // is this the stored version
// 	Deprecated bool   // is this version deprecated
// 	ObjectName string // CRD name e.g. "rum.dispatcher.io"
// }

// type INewRumCluster struct {
// 	Cfg    *rest.Config
// 	Client *http.Client

// 	Namespace string
// }

// func NewRumCluster(cfg INewRumCluster) (*RumCluster, error) {
// 	c, err := kubernetes.NewForConfigAndClient(cfg.Cfg, cfg.Client)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &RumCluster{
// 		client:    c,
// 		namespace: cfg.Namespace,
// 	}, nil
// }

// func newNoReplicaRumClusterComponent(Name string) *IRumClusterComponent {
// 	return &IRumClusterComponent{Name: Name, Spec: nil}
// }

// func newReplicaRumClusterComponent(Name string, Spec map[string]IRumClusterComponentSpec) *IRumClusterComponent {
// 	return &IRumClusterComponent{Name: Name, Spec: Spec}
// }

// func (r *RumCluster) defaultComponentsList() []*IRumClusterComponent {
// 	l := rumReplicaManager(IReplicaManager{Large: true}) // 4
// 	m := rumReplicaManager(IReplicaManager{Mid: true})   // 3
// 	s := rumReplicaManager(IReplicaManager{Small: true}) // 2
// 	xs := rumReplicaManager(IReplicaManager{XS: true})   // 1

// 	return []*IRumClusterComponent{
// 		// no replicas — internal primitives, in-process only
// 		newNoReplicaRumClusterComponent("Light"),
// 		newNoReplicaRumClusterComponent("Stack"),
// 		newNoReplicaRumClusterComponent("Budget"),
// 		newNoReplicaRumClusterComponent("Slate"),

// 		// replicated — these are actual pods K8s manages
// 		newReplicaRumClusterComponent("Rum", map[string]IRumClusterComponentSpec{
// 			"v1": {
// 				Replicas:   l, // 4 Rum pods
// 				ObjectName: "rum.rum.io",
// 				Served:     true,
// 				Storage:    true,
// 			},
// 		}),
// 		newReplicaRumClusterComponent("Store", map[string]IRumClusterComponentSpec{
// 			"v1": {
// 				Replicas:   m, // 3 Store pods
// 				ObjectName: "rum.store.io",
// 				Served:     true,
// 				Storage:    true,
// 			},
// 		}),
// 		newReplicaRumClusterComponent("Embd", map[string]IRumClusterComponentSpec{
// 			"v1": {
// 				Replicas:   s, // 2 Embd pods
// 				ObjectName: "rum.embd.io",
// 				Served:     true,
// 				Storage:    true,
// 			},
// 		}),
// 		newReplicaRumClusterComponent("Dispatcher", map[string]IRumClusterComponentSpec{
// 			"v1": {
// 				Replicas:   xs, // 1 Dispatcher pod
// 				ObjectName: "rum.dispatcher.io",
// 				Served:     true,
// 				Storage:    true,
// 			},
// 		}),
// 	}
// }

// func (r *RumCluster) app(ctx context.Context) error {
// 	comp := r.defaultComponentsList()

// 	for _, c := range comp {
// 		// no spec = no replicas = skip K8s entirely
// 		// these live in-process (Light, Stack, Budget)
// 		if c.Spec == nil {
// 			log.Printf("skipping %s — in-process only", c.Name)
// 			continue
// 		}

// 		for version, spec := range c.Spec {
// 			// 1. create the CRD — tells K8s this resource type exists
// 			if err := r.applyCRD(ctx, c, version, spec); err != nil {
// 				return fmt.Errorf("crd %s: %w", c.Name, err)
// 			}

// 			// 2. create the StatefulSet — tells K8s how many pods to run
// 			if err := r.applyStatefulSet(ctx, c, spec); err != nil {
// 				return fmt.Errorf("statefulset %s: %w", c.Name, err)
// 			}

// 			log.Printf("applied %s version=%s replicas=%d", c.Name, version, spec.Replicas)
// 		}
// 	}
// 	return nil
// }

// func (r *RumCluster) applyCRD(ctx context.Context, c *IRumClusterComponent, version string, spec IRumClusterComponentSpec) error {
// 	crd := &crumapp.CustomResourceDefinition{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name: spec.ObjectName, // e.g. "rum.rum.io"
// 		},
// 		Spec: crumapp.CustomResourceDefinitionSpec{
// 			Group: group, // "rum.onecolor.io"
// 			Names: crumapp.CustomResourceDefinitionNames{
// 				Kind:   c.Name,                        // "Rum"
// 				Plural: strings.ToLower(c.Name) + "s", // "rums"
// 			},
// 			Scope: crumapp.NamespaceScoped,
// 			Versions: []crumapp.CustomResourceDefinitionVersion{
// 				{
// 					Name:       version,
// 					Served:     spec.Served,
// 					Storage:    spec.Storage,
// 					Deprecated: spec.Deprecated,
// 				},
// 			},
// 		},
// 	}

// 	_, err := r.crdClient.ApiextensionsV1().
// 		CustomResourceDefinitions().
// 		Create(ctx, crd, metav1.CreateOptions{})
// 	if err != nil {
// 		log.Printf("crd %s: %v (may already exist)", spec.ObjectName, err)
// 	}
// 	return nil
// }

// func (r *RumCluster) applyStatefulSet(ctx context.Context, c *IRumClusterComponent, spec IRumClusterComponentSpec) error {
// 	replicas := int32(spec.Replicas)

// 	ss := &rumapp.StatefulSet{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      strings.ToLower(c.Name),
// 			Namespace: r.namespace,
// 			Labels:    map[string]string{"component": c.Name},
// 		},
// 		Spec: rumapp.StatefulSetSpec{
// 			Replicas:    &replicas,
// 			ServiceName: strings.ToLower(c.Name),
// 			Selector: &metav1.LabelSelector{
// 				MatchLabels: map[string]string{"component": c.Name},
// 			},
// 			Template: corev1.PodTemplateSpec{
// 				ObjectMeta: metav1.ObjectMeta{
// 					Labels: map[string]string{"component": c.Name},
// 				},
// 				Spec: corev1.PodSpec{
// 					Containers: []corev1.Container{
// 						{
// 							Name:  strings.ToLower(c.Name),
// 							Image: r.image, // set on RumCluster
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	_, err := r.client.AppsV1().
// 		StatefulSets(r.namespace).
// 		Create(ctx, ss, metav1.CreateOptions{})
// 	if err != nil {
// 		log.Printf("statefulset %s: %v (may already exist)", c.Name, err)
// 	}
// 	return nil
// }
