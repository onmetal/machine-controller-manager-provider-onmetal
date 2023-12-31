// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package onmetal

import (
	"context"
	"fmt"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	apiv1alpha1 "github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/api/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (d *onmetalDriver) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	if req.MachineClass.Provider != apiv1alpha1.ProviderName {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("requested provider '%s' is not suppored by the driver '%s'", req.MachineClass.Provider, apiv1alpha1.ProviderName))
	}

	klog.V(3).Infof("Machine list request has been received for %q", req.MachineClass.Name)
	defer klog.V(3).Infof("Machine list request has been processed for %q", req.MachineClass.Name)

	providerSpec, err := validateProviderSpecAndSecret(req.MachineClass, req.Secret)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("provider spec for requested provider '%s' is invalid: %v", req.MachineClass.Provider, err))
	}

	// Get onmetal machine list
	onmetalMachineList := &computev1alpha1.MachineList{}
	matchingLabels := client.MatchingLabels{}
	for k, v := range providerSpec.Labels {
		matchingLabels[k] = v
	}
	if err := d.OnmetelClient.List(ctx, onmetalMachineList, client.InNamespace(d.OnmetalNamespace), matchingLabels); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	//Creating machineList from onmetalMachineList items
	machineList := make(map[string]string, len(onmetalMachineList.Items))
	for _, machine := range onmetalMachineList.Items {
		machineID := getProviderIDForOnmetalMachine(&machine)
		machineList[machineID] = machine.Name
	}

	return &driver.ListMachinesResponse{MachineList: machineList}, nil
}
