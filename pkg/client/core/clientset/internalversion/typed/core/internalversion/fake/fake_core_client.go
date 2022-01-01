/*
Copyright (c) SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

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
	internalversion "github.com/gardener/gardener/pkg/client/core/clientset/internalversion/typed/core/internalversion"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeCore struct {
	*testing.Fake
}

func (c *FakeCore) BackupBuckets() internalversion.BackupBucketInterface {
	return &FakeBackupBuckets{c}
}

func (c *FakeCore) BackupEntries(namespace string) internalversion.BackupEntryInterface {
	return &FakeBackupEntries{c, namespace}
}

func (c *FakeCore) CloudProfiles() internalversion.CloudProfileInterface {
	return &FakeCloudProfiles{c}
}

func (c *FakeCore) ControllerDeployments() internalversion.ControllerDeploymentInterface {
	return &FakeControllerDeployments{c}
}

func (c *FakeCore) ControllerInstallations() internalversion.ControllerInstallationInterface {
	return &FakeControllerInstallations{c}
}

func (c *FakeCore) ControllerRegistrations() internalversion.ControllerRegistrationInterface {
	return &FakeControllerRegistrations{c}
}

func (c *FakeCore) ExposureClasses() internalversion.ExposureClassInterface {
	return &FakeExposureClasses{c}
}

func (c *FakeCore) Plants(namespace string) internalversion.PlantInterface {
	return &FakePlants{c, namespace}
}

func (c *FakeCore) Projects() internalversion.ProjectInterface {
	return &FakeProjects{c}
}

func (c *FakeCore) Quotas(namespace string) internalversion.QuotaInterface {
	return &FakeQuotas{c, namespace}
}

func (c *FakeCore) SecretBindings(namespace string) internalversion.SecretBindingInterface {
	return &FakeSecretBindings{c, namespace}
}

func (c *FakeCore) Seeds() internalversion.SeedInterface {
	return &FakeSeeds{c}
}

func (c *FakeCore) Shoots(namespace string) internalversion.ShootInterface {
	return &FakeShoots{c, namespace}
}

func (c *FakeCore) ShootExtensionStatuses(namespace string) internalversion.ShootExtensionStatusInterface {
	return &FakeShootExtensionStatuses{c, namespace}
}

func (c *FakeCore) ShootStates(namespace string) internalversion.ShootStateInterface {
	return &FakeShootStates{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeCore) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
