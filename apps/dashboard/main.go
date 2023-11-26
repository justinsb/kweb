package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/justinsb/kweb"
	"github.com/justinsb/kweb/components"
	"github.com/justinsb/kweb/templates/scopes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

//go:embed all:pages
var pages embed.FS

func main() {
	ctx := context.Background()

	log := klog.FromContext(ctx)

	app := &App{}

	opt := kweb.NewOptions("dashboard")
	opt.Server.Pages.Base = pages

	opt.Server.Pages.ScopeValues = append(opt.Server.Pages.ScopeValues, app.GlobalValues)

	var errors []error
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		name := f.Name
		envVar := name
		v := os.Getenv(envVar)
		if v != "" {
			if err := f.Value.Set(v); err != nil {
				errors = append(errors, fmt.Errorf("error setting flag %q to env var %q value %q: %w", name, envVar, v, err))
			}
		}
		log.Info("flag/env", "flag", f.Name, "env", envVar, "value", v)
	})
	if len(errors) != 0 {
		klog.Fatalf("error mapping env vars to flags: %v", errors[0])
	}

	restConfig, err := ctrl.GetConfig()
	if err != nil {
		klog.Fatalf("error getting kubernetes config: %v", err)
	}
	restConfig.QPS = 1000
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		klog.Fatalf("error building kubernetes client: %v", err)
	}
	app.kubeClient = kubeClient
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		klog.Fatalf("error building dynamic client: %v", err)
	}
	app.dynamicClient = dynamicClient

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		klog.Fatalf("error building discovery client: %v", err)
	}
	cachedDiscoveryClient := memory.NewMemCacheClient(discoveryClient)
	app.discoveryClient = cachedDiscoveryClient

	app.App, err = kweb.NewApp(opt)
	if err != nil {
		klog.Fatalf("error building app: %v", err)
	}

	app.Users()

	app.RunFromMain()
}

type App struct {
	*kweb.App

	kubeClient      *kubernetes.Clientset
	dynamicClient   dynamic.Interface
	discoveryClient discovery.CachedDiscoveryInterface
}

func (a *App) GlobalValues(ctx context.Context, scope *scopes.Scope) {
	scope.Values["nodes"] = scopes.Value{
		Function: func() interface{} {
			return a.Nodes(ctx)
		},
	}
	scope.Values["pods"] = scopes.Value{
		Function: func() interface{} {
			return a.Pods(ctx)
		},
	}
	scope.Values["namespaces"] = scopes.Value{
		Function: func() interface{} {
			return a.Namespaces(ctx)
		},
	}
	scope.Values["namespace"] = scopes.Value{
		Function: func() interface{} {
			return a.Namespace(ctx)
		},
	}
	scope.Values["objects"] = scopes.Value{
		Function: func() interface{} {
			return a.Objects(ctx)
		},
	}
	scope.Values["object"] = scopes.Value{
		Function: func() interface{} {
			return a.Object(ctx)
		},
	}
	scope.Values["groupresources"] = scopes.Value{
		Function: func() interface{} {
			return a.GroupResources(ctx)
		},
	}
	scope.Values["path"] = scopes.Value{
		Function: func() interface{} {
			return a.Path(ctx)
		},
	}

}

func (a *App) Nodes(ctx context.Context) interface{} {
	var opts metav1.ListOptions
	nodes, err := a.dynamicClient.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nodes"}).List(ctx, opts)
	if err != nil {
		klog.Fatalf("todo: %v", err)
	}
	return nodes.Items
}

func (a *App) Pods(ctx context.Context) interface{} {
	var opts metav1.ListOptions
	pods, err := a.dynamicClient.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}).List(ctx, opts)
	if err != nil {
		klog.Fatalf("todo: %v", err)
	}
	return pods.Items
}

func (a *App) Namespaces(ctx context.Context) interface{} {
	var opts metav1.ListOptions
	namespaces, err := a.dynamicClient.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}).List(ctx, opts)
	if err != nil {
		klog.Fatalf("todo: %v", err)
	}
	return namespaces.Items
}

func (a *App) Namespace(ctx context.Context) interface{} {
	req := components.GetRequest(ctx)
	name := req.PathParameter("name")

	var opts metav1.GetOptions
	namespace, err := a.dynamicClient.Resource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}).Get(ctx, name, opts)
	if err != nil {
		klog.Fatalf("todo: %v", err)
	}
	return namespace
}

func (a *App) preferredVersion(ctx context.Context, groupResource schema.GroupResource) (string, error) {
	response, err := a.discoveryClient.ServerPreferredResources()
	if err != nil {
		return "", fmt.Errorf("getting server preferred resources: %w", err)
	}

	for _, resourceList := range response {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			return "", fmt.Errorf("parsing group version %q: %w", resourceList.GroupVersion, err)
		}
		if gv.Group != groupResource.Group {
			continue
		}
		for _, r := range resourceList.APIResources {
			if r.Name == groupResource.Resource {
				return gv.Version, nil
			}
		}
	}

	return "", fmt.Errorf("cannot find version for %s", groupResource.String())
}

func (a *App) Objects(ctx context.Context) interface{} {
	req := components.GetRequest(ctx)
	group := req.PathParameter("group")
	resource := req.PathParameter("resource")
	version, err := a.preferredVersion(ctx, schema.GroupResource{Group: group, Resource: resource})
	if err != nil {
		klog.Fatalf("todo: %v", err)
	}
	gvr := schema.GroupVersionResource{Group: group, Resource: resource, Version: version}
	var opts metav1.ListOptions
	response, err := a.dynamicClient.Resource(gvr).List(ctx, opts)
	if err != nil {
		klog.Fatalf("todo: %v", err)
	}
	return response.Items
}

func (a *App) Object(ctx context.Context) interface{} {
	req := components.GetRequest(ctx)

	group := req.PathParameter("group")
	resource := req.PathParameter("resource")
	version, err := a.preferredVersion(ctx, schema.GroupResource{Group: group, Resource: resource})
	if err != nil {
		klog.Fatalf("todo: %v", err)
	}

	name := req.PathParameter("name")
	namespace := req.PathParameter("namespace")

	gvr := schema.GroupVersionResource{Group: group, Resource: resource, Version: version}
	var opts metav1.GetOptions
	response, err := a.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, opts)
	if err != nil {
		klog.Fatalf("todo get %v/%s: %v", gvr, name, err)
	}
	return response
}

func (a *App) Path(ctx context.Context) interface{} {
	req := components.GetRequest(ctx)

	return req.PathParameters
}

func (a *App) GroupResources(ctx context.Context) interface{} {
	response, err := a.discoveryClient.ServerPreferredResources()
	if err != nil {
		klog.Fatalf("todo: %v", err)
	}

	var grs []schema.GroupResource
	for _, resourceList := range response {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			klog.Fatalf("todo: %v", err)
		}
		for _, r := range resourceList.APIResources {
			grs = append(grs, schema.GroupResource{
				Group:    gv.Group,
				Resource: r.Name,
			})
		}
	}

	return grs
}
