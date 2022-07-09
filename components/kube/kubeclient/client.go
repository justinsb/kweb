package kubeclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/justinsb/kweb/components/kube"
	"github.com/justinsb/kweb/components/kube/kubejson"
	"google.golang.org/protobuf/proto"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type Client struct {
	dynamic    dynamic.Interface
	restConfig *rest.Config
}

func New(restConfig *rest.Config) (*Client, error) {
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		dynamic:    dynamicClient,
		restConfig: restConfig,
	}, nil
}

func (c *Client) Dynamic() dynamic.Interface {
	return c.dynamic
}

func (c *Client) Get(ctx context.Context, id types.NamespacedName, obj proto.Message) error {
	kindInfo, err := kube.GetKindInfo(obj)
	if err != nil {
		return err
	}

	httpClient, err := rest.HTTPClientFor(c.restConfig)
	if err != nil {
		return err
	}

	if id.Name == "" {
		return fmt.Errorf("name is required")
	}

	var path []string
	if c.restConfig.APIPath != "" {
		path = append(path, c.restConfig.APIPath)
	}
	if kindInfo.Group == "" {
		path = append(path, "api")
	} else {
		path = append(path, "apis", kindInfo.Group)
	}
	path = append(path, kindInfo.Version)

	if id.Namespace != "" {
		path = append(path, "namespaces", id.Namespace)
	}
	path = append(path, kindInfo.Resource)

	path = append(path, id.Name)

	url := c.restConfig.Host + "/" + strings.Join(path, "/")
	klog.Infof("GET url is %v", url)

	httpRequest, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error building request: %w", err)
	}
	response, err := httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("error from request: %w", err)
	}
	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	klog.Infof("response is %v", string(b))

	if response.StatusCode != 200 {
		switch response.StatusCode {
		case 404:
			return apierrors.NewNotFound(kindInfo.GroupResource(), id.Name)
		}
		return fmt.Errorf("unexpected response %v", response.Status)
	}

	parser := kubejson.UnmarshalOptions{}
	if err := parser.Unmarshal(b, obj); err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	return nil
}

func (c *Client) Create(ctx context.Context, obj kube.Object) error {
	metadata := obj.GetMetadata()

	kindInfo, err := kube.GetKindInfo(obj)
	if err != nil {
		return err
	}

	httpClient, err := rest.HTTPClientFor(c.restConfig)
	if err != nil {
		return err
	}

	if metadata.Name == "" {
		return fmt.Errorf("name is required")
	}

	var path []string
	if c.restConfig.APIPath != "" {
		path = append(path, c.restConfig.APIPath)
	}
	if kindInfo.Group == "" {
		path = append(path, "api")
	} else {
		path = append(path, "apis", kindInfo.Group)
	}
	path = append(path, kindInfo.Version)

	if metadata.Namespace != "" {
		path = append(path, "namespaces", metadata.Namespace)
	}
	path = append(path, kindInfo.Resource)

	url := c.restConfig.Host + "/" + strings.Join(path, "/")
	klog.Infof("create url is %v", url)

	body, err := kubejson.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal to JSON: %w", err)
	}

	klog.Infof("body is %v", string(body))

	httpRequest, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("error building request: %w", err)
	}
	httpRequest.Header.Add("Content-Type", runtime.ContentTypeJSON)

	response, err := httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("error from request: %w", err)
	}
	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	klog.Infof("response is %v", string(b))

	if response.StatusCode != 201 {
		switch response.StatusCode {
		case 404:
			return apierrors.NewNotFound(kindInfo.GroupResource(), metadata.Name)
		}
		return fmt.Errorf("unexpected response %v", response.Status)
	}

	parser := kubejson.UnmarshalOptions{}
	if err := parser.Unmarshal(b, obj); err != nil {
		return fmt.Errorf("error parsing response: %w", err)
	}

	return nil
}
