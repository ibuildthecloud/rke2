/*
Copyright The Kubernetes Authors.

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

// Code generated by main. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	v1 "github.com/rancher/k3s/pkg/apis/k3s.cattle.io/v1"
	clientset "github.com/rancher/k3s/pkg/generated/clientset/versioned/typed/k3s.cattle.io/v1"
	informers "github.com/rancher/k3s/pkg/generated/informers/externalversions/k3s.cattle.io/v1"
	listers "github.com/rancher/k3s/pkg/generated/listers/k3s.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type AddonHandler func(string, *v1.Addon) (*v1.Addon, error)

type AddonController interface {
	generic.ControllerMeta
	AddonClient

	OnChange(ctx context.Context, name string, sync AddonHandler)
	OnRemove(ctx context.Context, name string, sync AddonHandler)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, duration time.Duration)

	Cache() AddonCache
}

type AddonClient interface {
	Create(*v1.Addon) (*v1.Addon, error)
	Update(*v1.Addon) (*v1.Addon, error)
	UpdateStatus(*v1.Addon) (*v1.Addon, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.Addon, error)
	List(namespace string, opts metav1.ListOptions) (*v1.AddonList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Addon, err error)
}

type AddonCache interface {
	Get(namespace, name string) (*v1.Addon, error)
	List(namespace string, selector labels.Selector) ([]*v1.Addon, error)

	AddIndexer(indexName string, indexer AddonIndexer)
	GetByIndex(indexName, key string) ([]*v1.Addon, error)
}

type AddonIndexer func(obj *v1.Addon) ([]string, error)

type addonController struct {
	controllerManager *generic.ControllerManager
	clientGetter      clientset.AddonsGetter
	informer          informers.AddonInformer
	gvk               schema.GroupVersionKind
}

func NewAddonController(gvk schema.GroupVersionKind, controllerManager *generic.ControllerManager, clientGetter clientset.AddonsGetter, informer informers.AddonInformer) AddonController {
	return &addonController{
		controllerManager: controllerManager,
		clientGetter:      clientGetter,
		informer:          informer,
		gvk:               gvk,
	}
}

func FromAddonHandlerToHandler(sync AddonHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.Addon
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.Addon))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *addonController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.Addon))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateAddonDeepCopyOnChange(client AddonClient, obj *v1.Addon, handler func(obj *v1.Addon) (*v1.Addon, error)) (*v1.Addon, error) {
	if obj == nil {
		return obj, nil
	}

	copyObj := obj.DeepCopy()
	newObj, err := handler(copyObj)
	if newObj != nil {
		copyObj = newObj
	}
	if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
		return client.Update(copyObj)
	}

	return copyObj, err
}

func (c *addonController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, handler)
}

func (c *addonController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), handler)
	c.controllerManager.AddHandler(ctx, c.gvk, c.informer.Informer(), name, removeHandler)
}

func (c *addonController) OnChange(ctx context.Context, name string, sync AddonHandler) {
	c.AddGenericHandler(ctx, name, FromAddonHandlerToHandler(sync))
}

func (c *addonController) OnRemove(ctx context.Context, name string, sync AddonHandler) {
	removeHandler := generic.NewRemoveHandler(name, c.Updater(), FromAddonHandlerToHandler(sync))
	c.AddGenericHandler(ctx, name, removeHandler)
}

func (c *addonController) Enqueue(namespace, name string) {
	c.controllerManager.Enqueue(c.gvk, c.informer.Informer(), namespace, name)
}

func (c *addonController) EnqueueAfter(namespace, name string, duration time.Duration) {
	c.controllerManager.EnqueueAfter(c.gvk, c.informer.Informer(), namespace, name, duration)
}

func (c *addonController) Informer() cache.SharedIndexInformer {
	return c.informer.Informer()
}

func (c *addonController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *addonController) Cache() AddonCache {
	return &addonCache{
		lister:  c.informer.Lister(),
		indexer: c.informer.Informer().GetIndexer(),
	}
}

func (c *addonController) Create(obj *v1.Addon) (*v1.Addon, error) {
	return c.clientGetter.Addons(obj.Namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
}

func (c *addonController) Update(obj *v1.Addon) (*v1.Addon, error) {
	return c.clientGetter.Addons(obj.Namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
}

func (c *addonController) UpdateStatus(obj *v1.Addon) (*v1.Addon, error) {
	return c.clientGetter.Addons(obj.Namespace).UpdateStatus(context.TODO(), obj, metav1.UpdateOptions{})
}

func (c *addonController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.clientGetter.Addons(namespace).Delete(context.TODO(), name, *options)
}

func (c *addonController) Get(namespace, name string, options metav1.GetOptions) (*v1.Addon, error) {
	return c.clientGetter.Addons(namespace).Get(context.TODO(), name, options)
}

func (c *addonController) List(namespace string, opts metav1.ListOptions) (*v1.AddonList, error) {
	return c.clientGetter.Addons(namespace).List(context.TODO(), opts)
}

func (c *addonController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.clientGetter.Addons(namespace).Watch(context.TODO(), opts)
}

func (c *addonController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Addon, err error) {
	return c.clientGetter.Addons(namespace).Patch(context.TODO(), name, pt, data, metav1.PatchOptions{}, subresources...)
}

type addonCache struct {
	lister  listers.AddonLister
	indexer cache.Indexer
}

func (c *addonCache) Get(namespace, name string) (*v1.Addon, error) {
	return c.lister.Addons(namespace).Get(name)
}

func (c *addonCache) List(namespace string, selector labels.Selector) ([]*v1.Addon, error) {
	return c.lister.Addons(namespace).List(selector)
}

func (c *addonCache) AddIndexer(indexName string, indexer AddonIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.Addon))
		},
	}))
}

func (c *addonCache) GetByIndex(indexName, key string) (result []*v1.Addon, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v1.Addon, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v1.Addon))
	}
	return result, nil
}

type AddonStatusHandler func(obj *v1.Addon, status v1.AddonStatus) (v1.AddonStatus, error)

type AddonGeneratingHandler func(obj *v1.Addon, status v1.AddonStatus) ([]runtime.Object, v1.AddonStatus, error)

func RegisterAddonStatusHandler(ctx context.Context, controller AddonController, condition condition.Cond, name string, handler AddonStatusHandler) {
	statusHandler := &addonStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, FromAddonHandlerToHandler(statusHandler.sync))
}

func RegisterAddonGeneratingHandler(ctx context.Context, controller AddonController, apply apply.Apply,
	condition condition.Cond, name string, handler AddonGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &addonGeneratingHandler{
		AddonGeneratingHandler: handler,
		apply:                  apply,
		name:                   name,
		gvk:                    controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterAddonStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type addonStatusHandler struct {
	client    AddonClient
	condition condition.Cond
	handler   AddonStatusHandler
}

func (a *addonStatusHandler) sync(key string, obj *v1.Addon) (*v1.Addon, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		var newErr error
		obj.Status = newStatus
		obj, newErr = a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
	}
	return obj, err
}

type addonGeneratingHandler struct {
	AddonGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *addonGeneratingHandler) Remove(key string, obj *v1.Addon) (*v1.Addon, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v1.Addon{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

func (a *addonGeneratingHandler) Handle(obj *v1.Addon, status v1.AddonStatus) (v1.AddonStatus, error) {
	objs, newStatus, err := a.AddonGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	return newStatus, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}
