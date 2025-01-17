/*
Copyright 2016 The Kubernetes Authors.

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

package storage

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// APIResourceConfigSource is the interface to determine which groups and versions are enabled
type APIResourceConfigSource interface {
	ResourceEnabled(resource schema.GroupVersionResource) bool
	AnyResourceForGroupEnabled(group string) bool
}

var _ APIResourceConfigSource = &ResourceConfig{}

type ResourceConfig struct {
	GroupVersionConfigs map[schema.GroupVersion]bool
	ResourceConfigs     map[schema.GroupVersionResource]bool
}

func NewResourceConfig() *ResourceConfig {
	return &ResourceConfig{GroupVersionConfigs: map[schema.GroupVersion]bool{}, ResourceConfigs: map[schema.GroupVersionResource]bool{}}
}

// DisableMatchingVersions disables all group/versions for which the matcher function returns true. It does not modify individual resource enablement/disablement.
func (o *ResourceConfig) DisableMatchingVersions(matcher func(gv schema.GroupVersion) bool) {
	for k := range o.GroupVersionConfigs {
		if matcher(k) {
			o.GroupVersionConfigs[k] = false
		}
	}
}

// EnableMatchingVersions enables all group/versions for which the matcher function returns true. It does not modify individual resource enablement/disablement.
func (o *ResourceConfig) EnableMatchingVersions(matcher func(gv schema.GroupVersion) bool) {
	for k := range o.GroupVersionConfigs {
		if matcher(k) {
			o.GroupVersionConfigs[k] = true
		}
	}
}

// RemoveMatchingResourcePreferences removes individual resource preferences that match.  This is useful when an override of a version or level enablement should
// override the previously individual preferences.
func (o *ResourceConfig) RemoveMatchingResourcePreferences(matcher func(gvr schema.GroupVersionResource) bool) {
	keysToRemove := []schema.GroupVersionResource{}
	for k := range o.ResourceConfigs {
		if matcher(k) {
			keysToRemove = append(keysToRemove, k)
		}
	}
	for _, k := range keysToRemove {
		delete(o.ResourceConfigs, k)
	}
}

// DisableVersions disables the versions entirely.
func (o *ResourceConfig) DisableVersions(versions ...schema.GroupVersion) {
	for _, version := range versions {
		o.GroupVersionConfigs[version] = false
	}
}

func (o *ResourceConfig) EnableVersions(versions ...schema.GroupVersion) {
	for _, version := range versions {
		o.GroupVersionConfigs[version] = true
	}
}

// TODO this must be removed and we enable/disable individual resources.
func (o *ResourceConfig) VersionEnabled(version schema.GroupVersion) bool {
	enabled, _ := o.GroupVersionConfigs[version]
	return enabled
}

func (o *ResourceConfig) DisableResources(resources ...schema.GroupVersionResource) {
	for _, resource := range resources {
		o.ResourceConfigs[resource] = false
	}
}

func (o *ResourceConfig) EnableResources(resources ...schema.GroupVersionResource) {
	for _, resource := range resources {
		o.ResourceConfigs[resource] = true
	}
}

func (o *ResourceConfig) ResourceEnabled(resource schema.GroupVersionResource) bool {
	// if a resource is explicitly set, that takes priority over the preference of the version.
	resourceEnabled, explicitlySet := o.ResourceConfigs[resource]
	if explicitlySet {
		return resourceEnabled
	}

	if !o.VersionEnabled(resource.GroupVersion()) {
		return false
	}
	// they are enabled by default.
	return true
}

func (o *ResourceConfig) AnyResourceForGroupEnabled(group string) bool {
	for version := range o.GroupVersionConfigs {
		if version.Group == group {
			if o.VersionEnabled(version) {
				return true
			}
		}
	}
	for resource := range o.ResourceConfigs {
		if resource.Group == group && o.ResourceEnabled(resource) {
			return true
		}
	}

	return false
}
