/*
Copyright (c) 2021 TriggerMesh Inc.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/triggermesh/routing/pkg/apis/routing/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSplitters implements SplitterInterface
type FakeSplitters struct {
	Fake *FakeRoutingV1alpha1
	ns   string
}

var splittersResource = schema.GroupVersionResource{Group: "routing.triggermesh.io", Version: "v1alpha1", Resource: "splitters"}

var splittersKind = schema.GroupVersionKind{Group: "routing.triggermesh.io", Version: "v1alpha1", Kind: "Splitter"}

// Get takes name of the splitter, and returns the corresponding splitter object, and an error if there is any.
func (c *FakeSplitters) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Splitter, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(splittersResource, c.ns, name), &v1alpha1.Splitter{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Splitter), err
}

// List takes label and field selectors, and returns the list of Splitters that match those selectors.
func (c *FakeSplitters) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SplitterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(splittersResource, splittersKind, c.ns, opts), &v1alpha1.SplitterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SplitterList{ListMeta: obj.(*v1alpha1.SplitterList).ListMeta}
	for _, item := range obj.(*v1alpha1.SplitterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested splitters.
func (c *FakeSplitters) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(splittersResource, c.ns, opts))

}

// Create takes the representation of a splitter and creates it.  Returns the server's representation of the splitter, and an error, if there is any.
func (c *FakeSplitters) Create(ctx context.Context, splitter *v1alpha1.Splitter, opts v1.CreateOptions) (result *v1alpha1.Splitter, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(splittersResource, c.ns, splitter), &v1alpha1.Splitter{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Splitter), err
}

// Update takes the representation of a splitter and updates it. Returns the server's representation of the splitter, and an error, if there is any.
func (c *FakeSplitters) Update(ctx context.Context, splitter *v1alpha1.Splitter, opts v1.UpdateOptions) (result *v1alpha1.Splitter, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(splittersResource, c.ns, splitter), &v1alpha1.Splitter{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Splitter), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSplitters) UpdateStatus(ctx context.Context, splitter *v1alpha1.Splitter, opts v1.UpdateOptions) (*v1alpha1.Splitter, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(splittersResource, "status", c.ns, splitter), &v1alpha1.Splitter{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Splitter), err
}

// Delete takes name of the splitter and deletes it. Returns an error if one occurs.
func (c *FakeSplitters) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(splittersResource, c.ns, name), &v1alpha1.Splitter{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSplitters) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(splittersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.SplitterList{})
	return err
}

// Patch applies the patch and returns the patched splitter.
func (c *FakeSplitters) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Splitter, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(splittersResource, c.ns, name, pt, data, subresources...), &v1alpha1.Splitter{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Splitter), err
}