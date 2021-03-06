/*
 * Copyright 2018-2021, CS Systemes d'Information, http://csgroup.eu
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package listeners

import (
	"context"

	"github.com/asaskevich/govalidator"
	googleprotobuf "github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"

	"github.com/CS-SI/SafeScale/lib/protocol"
	"github.com/CS-SI/SafeScale/lib/server/handlers"
	"github.com/CS-SI/SafeScale/lib/server/iaas"
	"github.com/CS-SI/SafeScale/lib/utils/debug"
	"github.com/CS-SI/SafeScale/lib/utils/debug/tracing"
	"github.com/CS-SI/SafeScale/lib/utils/fail"
)

// Tenant structure to handle name and clientAPI for a tenant
type Tenant struct {
	name    string
	Service iaas.Service
}

var (
	currentTenant *Tenant
)

// GetCurrentTenant contains the current tenant
var GetCurrentTenant = getCurrentTenant

// getCurrentTenant returns the tenant used for commands or, if not set, set the tenant to use if it is the only one registered
func getCurrentTenant() *Tenant {
	if currentTenant == nil {
		tenants, err := iaas.GetTenantNames()
		if err != nil || len(tenants) != 1 {
			return nil
		}
		// Set unique tenant as selected
		logrus.Println("Unique tenant set")
		for name := range tenants {
			service, err := iaas.UseService(name)
			if err != nil {
				return nil
			}
			currentTenant = &Tenant{name: name, Service: service}
		}
	}
	return currentTenant
}

// TenantListener server is used to implement SafeScale.safescale.
type TenantListener struct{}

// List lists registered tenants
func (s *TenantListener) List(ctx context.Context, in *googleprotobuf.Empty) (_ *protocol.TenantList, err error) {
	defer fail.OnExitConvertToGRPCStatus(&err)
	defer fail.OnExitWrapError(&err, "cannot list tenants")

	if s == nil {
		return nil, fail.InvalidInstanceError()
	}
	if ctx == nil {
		return nil, fail.InvalidParameterError("ctx", "cannot be nil")
	}

	ok, err := govalidator.ValidateStruct(in)
	if err != nil || !ok {
		logrus.Warnf("Structure validation failure: %v", in) // FIXME Generate json tags in protobuf
	}

	defer fail.OnExitLogError(&err)

	tenants, xerr := iaas.GetTenantNames()
	if xerr != nil {
		return nil, xerr
	}

	var list []*protocol.Tenant
	for tenantName, providerName := range tenants {
		list = append(list, &protocol.Tenant{
			Name:     tenantName,
			Provider: providerName,
		})
	}

	return &protocol.TenantList{Tenants: list}, nil
}

// Get returns the name of the current tenant used
func (s *TenantListener) Get(ctx context.Context, in *googleprotobuf.Empty) (_ *protocol.TenantName, err error) {
	defer fail.OnExitConvertToGRPCStatus(&err)

	if s == nil {
		return nil, fail.InvalidInstanceError()
	}
	if ctx == nil {
		return nil, fail.InvalidParameterError("ctx", "cannot be nil")
	}

	if ok, err := govalidator.ValidateStruct(in); err != nil && !ok {
		logrus.Warnf("Structure validation failure: %v", in) // FIXME: Generate json tags in protobuf
	}

	defer fail.OnExitLogError(&err)

	getCurrentTenant()
	if currentTenant == nil {
		return nil, fail.NotFoundError("no tenant set")
	}
	return &protocol.TenantName{Name: currentTenant.name}, nil
}

// Set the the tenant to use for each command
func (s *TenantListener) Set(ctx context.Context, in *protocol.TenantName) (empty *googleprotobuf.Empty, err error) {
	defer fail.OnExitConvertToGRPCStatus(&err)
	defer fail.OnExitWrapError(&err, "cannot set tenant")

	empty = &googleprotobuf.Empty{}
	if s == nil {
		return empty, fail.InvalidInstanceError()
	}
	if ctx == nil {
		return empty, fail.InvalidParameterError("ctx", "cannot be nil")
	}
	if in == nil {
		return empty, fail.InvalidParameterError("in", "cannot be nil")
	}

	ok, err := govalidator.ValidateStruct(in)
	if err != nil || !ok {
		logrus.Warnf("Structure validation failure: %v", in) // FIXME: Generate json tags in protobuf
	}

	defer fail.OnExitLogError(&err)

	if currentTenant != nil && currentTenant.name == in.GetName() {
		return empty, nil
	}

	service, xerr := iaas.UseService(in.GetName())
	if xerr != nil {
		return empty, xerr
	}
	currentTenant = &Tenant{name: in.GetName(), Service: service}
	return empty, nil
}

// Cleanup removes everything corresponding to SafeScale from tenant (metadata in particular)
func (s *TenantListener) Cleanup(ctx context.Context, in *protocol.TenantCleanupRequest) (empty *googleprotobuf.Empty, err error) {
	defer fail.OnExitConvertToGRPCStatus(&err)
	defer fail.OnExitWrapError(&err, "cannot cleanup tenant")

	empty = &googleprotobuf.Empty{}
	if s == nil {
		return empty, fail.InvalidInstanceError()
	}
	if ctx == nil {
		return empty, fail.InvalidParameterError("ctx", "cannot be nil")
	}
	if in == nil {
		return empty, fail.InvalidParameterError("in", "cannot be nil")
	}

	ok, err := govalidator.ValidateStruct(in)
	if err == nil {
		if !ok {
			logrus.Warnf("Structure validation failure: %v", in) // FIXME Generate json tags in protobuf
		}
	}

	job, xerr := PrepareJob(ctx, "", "tenant cleanup")
	if xerr != nil {
		return nil, xerr
	}
	defer job.Close()

	name := in.GetName()
	tracer := debug.NewTracer(job.GetTask(), tracing.ShouldTrace("listeners.tenant"), "('%s')", name).WithStopwatch().Entering()
	defer tracer.Exiting()
	defer fail.OnExitLogError(&err, tracer.TraceMessage())

	if currentTenant != nil && currentTenant.name == in.GetName() {
		return empty, nil
	}

	service, xerr := iaas.UseService(in.GetName())
	if xerr != nil {
		return empty, xerr
	}

	xerr = service.TenantCleanup(in.Force)
	return empty, xerr
}

// Scan proceeds a scan of host corresponding to each template to gather real data(metadata in particular)
func (s *TenantListener) Scan(ctx context.Context, in *googleprotobuf.Empty) (empty *googleprotobuf.Empty, err error) {
	defer fail.OnExitConvertToGRPCStatus(&err)
	defer fail.OnExitWrapError(&err, "cannot scan tenant")

	empty = &googleprotobuf.Empty{}
	if s == nil {
		return empty, fail.InvalidInstanceError()
	}
	if ctx == nil {
		return empty, fail.InvalidParameterError("ctx", "cannot be nil")
	}

	job, xerr := PrepareJob(ctx, "", "tenant scan")
	if xerr != nil {
		return nil, xerr
	}
	defer job.Close()

	getCurrentTenant()
	if currentTenant == nil {
		return nil, fail.NotFoundError("no tenant set")
	}

	name := currentTenant.name
	tracer := debug.NewTracer(job.GetTask(), tracing.ShouldTrace("listeners.tenant"), "('%s')", name).WithStopwatch().Entering()
	defer tracer.Exiting()
	defer fail.OnExitLogError(&err, tracer.TraceMessage())

	handler := handlers.NewScannerHandler(job)

	xerr = handler.Scan()
	return empty, xerr
}

// Inspect returns information about a tenant
func (s *TenantListener) Inspect(ctx context.Context, in *protocol.TenantName) (_ *protocol.TenantInspectResponse, xerr error) {
	defer fail.OnExitConvertToGRPCStatus(&xerr)
	defer fail.OnExitWrapError(&xerr, "cannot inspect tenant")

	if s == nil {
		return nil, fail.InvalidInstanceError()
	}
	if ctx == nil {
		return nil, fail.InvalidParameterError("ctx", "cannot be nil")
	}
	if in == nil {
		return nil, fail.InvalidParameterError("in", "cannot be nil")
	}

	ok, err := govalidator.ValidateStruct(in)
	if err != nil || !ok {
		logrus.Warnf("Structure validation failure: %v", in) // FIXME: Generate json tags in protobuf
	}

	job, xerr := PrepareJob(ctx, "", "tenant inspect")
	if xerr != nil {
		return nil, xerr
	}
	defer job.Close()

	name := in.GetName()
	tracer := debug.NewTracer(job.GetTask(), tracing.ShouldTrace("listeners.tenant"), "('%s')", name).WithStopwatch().Entering()
	defer tracer.Exiting()
	defer fail.OnExitLogError(&err, tracer.TraceMessage())

	return nil, fail.NotImplementedError("tenant inspect not yet implemented")
}
