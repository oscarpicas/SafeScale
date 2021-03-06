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
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or provideried.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cloudferro

import (
	"regexp"

	"github.com/asaskevich/govalidator"
	"github.com/sirupsen/logrus"

	"github.com/CS-SI/SafeScale/lib/server/iaas/stacks/api"

	"github.com/CS-SI/SafeScale/lib/server/iaas"
	"github.com/CS-SI/SafeScale/lib/server/iaas/objectstorage"
	"github.com/CS-SI/SafeScale/lib/server/iaas/providers"
	"github.com/CS-SI/SafeScale/lib/server/iaas/stacks"
	"github.com/CS-SI/SafeScale/lib/server/iaas/stacks/openstack"
	"github.com/CS-SI/SafeScale/lib/server/resources/abstract"
	"github.com/CS-SI/SafeScale/lib/server/resources/enums/volumespeed"
	"github.com/CS-SI/SafeScale/lib/utils/fail"
)

var (
	cloudferroIdentityEndpoint = "https://cf2.cloudferro.com:5000/v3"
	cloudferroDefaultImage     = "Ubuntu 18.04"
	cloudferroDNSServers       = []string{"185.48.234.234", "185.48.234.238"}
)

// provider is the implementation of the CloudFerro provider
type provider struct {
	api.Stack /**openstack.Stack*/

	tenantParameters map[string]interface{}
}

// New creates a new instance of cloudferro provider
func New() providers.Provider {
	return &provider{}
}

// Build build a new Client from configuration parameter
// Can be called from nil
func (p *provider) Build(params map[string]interface{}) (providers.Provider, fail.Error) {
	// tenantName, _ := params["name"].(string)

	identity, _ := params["identity"].(map[string]interface{})
	compute, _ := params["compute"].(map[string]interface{})
	network, _ := params["network"].(map[string]interface{})

	username, _ := identity["Username"].(string)
	password, _ := identity["Password"].(string)
	domainName, _ := identity["DomainName"].(string)

	// region, _ := compute["Region"].(string)
	region := "RegionOne"
	// zone, _ := compute["AvailabilityZone"].(string)
	zone := "nova"
	projectName, _ := compute["ProjectName"].(string)
	// projectID, _ := compute["ProjectID"].(string)
	defaultImage, _ := compute["DefaultImage"].(string)
	if defaultImage == "" {
		defaultImage = cloudferroDefaultImage
	}
	operatorUsername := abstract.DefaultUser
	if operatorUsernameIf, ok := compute["OperatorUsername"]; ok {
		operatorUsername = operatorUsernameIf.(string)
		if operatorUsername == "" {
			logrus.Warnf("OperatorUsername is empty ! Check your tenants.toml file ! Using 'safescale' user instead.")
			operatorUsername = abstract.DefaultUser
		}
	}

	providerNetwork, _ := network["ProviderNetwork"].(string)
	if providerNetwork == "" {
		providerNetwork = "external"
	}
	floatingIPPool, _ := network["FloatingIPPool"].(string)
	if floatingIPPool == "" {
		floatingIPPool = providerNetwork
	}

	authOptions := stacks.AuthenticationOptions{
		IdentityEndpoint: cloudferroIdentityEndpoint,
		Username:         username,
		Password:         password,
		DomainName:       domainName,
		TenantName:       projectName,
		Region:           region,
		AvailabilityZone: zone,
		FloatingIPPool:   floatingIPPool, // FIXME: move in ConfigurationOptions
		AllowReauth:      true,
	}

	govalidator.TagMap["alphanumwithdashesandunderscores"] = govalidator.Validator(func(str string) bool {
		rxp := regexp.MustCompile(stacks.AlphanumericWithDashesAndUnderscores)
		return rxp.Match([]byte(str))
	})

	if _, err := govalidator.ValidateStruct(authOptions); err != nil {
		return nil, fail.ToError(err)
	}

	providerName := "openstack"
	metadataBucketName, xerr := objectstorage.BuildMetadataBucketName(providerName, region, domainName, projectName)
	if xerr != nil {
		return nil, xerr
	}

	cfgOptions := stacks.ConfigurationOptions{
		ProviderNetwork:           providerNetwork,
		UseFloatingIP:             true,
		UseLayer3Networking:       true,
		AutoHostNetworkInterfaces: true,
		VolumeSpeeds: map[string]volumespeed.Enum{
			"HDD": volumespeed.HDD,
			"SSD": volumespeed.SSD,
		},
		MetadataBucket:   metadataBucketName,
		DNSList:          cloudferroDNSServers,
		DefaultImage:     defaultImage,
		OperatorUsername: operatorUsername,
		ProviderName:     providerName,
	}

	stack, xerr := openstack.New(authOptions, nil, cfgOptions, nil)
	if xerr != nil {
		return nil, xerr
	}
	//xerr = stack.InitDefaultSecurityGroups()
	//if xerr != nil {
	//	return nil, xerr
	//}

	// VPL: moved to stacks.openstack.New()
	// validRegions, xerr := stack.ListRegions()
	// if xerr != nil {
	// 	switch xerr.(type) {
	// 	case *fail.ErrNotFound:
	// 		// continue
	// 	default:
	// 		return nil, xerr
	// 	}
	// } else {
	// 	if len(validRegions) != 0 {
	// 		regionIsValidInput := false
	// 		for _, vr := range validRegions {
	// 			if region == vr {
	// 				regionIsValidInput = true
	// 			}
	// 		}
	// 		if !regionIsValidInput {
	// 			return nil, fail.InvalidRequestError("invalid Region '%s'", region)
	// 		}
	// 	}
	// }
	//
	// validAvailabilityZones, xerr := stack.ListAvailabilityZones()
	// if xerr != nil {
	// 	switch xerr.(type) {
	// 	case *fail.ErrNotFound:
	// 		// continue
	// 	default:
	// 		return nil, xerr
	// 	}
	// } else {
	// 	if len(validAvailabilityZones) != 0 {
	// 		var validZones []string
	// 		zoneIsValidInput := false
	// 		for az, valid := range validAvailabilityZones {
	// 			if valid {
	// 				if az == zone {
	// 					zoneIsValidInput = true
	// 				}
	// 				validZones = append(validZones, `'`+az+`'`)
	// 			}
	// 		}
	// 		if !zoneIsValidInput {
	// 			return nil, fail.InvalidRequestError("invalid Availability zone '%s', valid zones are %s", zone, strings.Join(validZones, ","))
	// 		}
	// 	}
	// }

	newP := &provider{
		Stack:            stack,
		tenantParameters: params,
	}
	return newP, nil
}

// GetAuthenticationOptions returns the auth options
func (p provider) GetAuthenticationOptions() (providers.Config, fail.Error) {
	cfg := providers.ConfigMap{}
	if p.IsNull() {
		return cfg, fail.InvalidInstanceError()
	}

	opts := p.Stack.(api.ReservedForProviderUse).GetAuthenticationOptions()
	cfg.Set("TenantName", opts.TenantName)
	cfg.Set("Login", opts.Username)
	cfg.Set("Password", opts.Password)
	cfg.Set("AuthUrl", opts.IdentityEndpoint)
	cfg.Set("Region", opts.Region)
	return cfg, nil
}

// GetConfigurationOptions return configuration parameters
func (p provider) GetConfigurationOptions() (providers.Config, fail.Error) {
	cfg := providers.ConfigMap{}
	if p.IsNull() {
		return cfg, fail.InvalidInstanceError()
	}

	opts := p.Stack.(api.ReservedForProviderUse).GetConfigurationOptions()
	cfg.Set("DNSList", opts.DNSList)
	cfg.Set("AutoHostNetworkInterfaces", opts.AutoHostNetworkInterfaces)
	cfg.Set("UseLayer3Networking", opts.UseLayer3Networking)
	cfg.Set("DefaultImage", opts.DefaultImage)
	cfg.Set("MetadataBucketName", opts.MetadataBucket)
	cfg.Set("OperatorUsername", opts.OperatorUsername)
	cfg.Set("ProviderName", p.GetName())
	cfg.Set("UseNATService", opts.UseNATService)

	return cfg, nil
}

// ListTemplates ...
// Value of all has no impact on the result
func (p provider) ListTemplates(all bool) ([]abstract.HostTemplate, fail.Error) {
	if p.IsNull() {
		return []abstract.HostTemplate{}, fail.InvalidInstanceError()
	}

	allTemplates, xerr := p.Stack.(api.ReservedForProviderUse).ListTemplates()
	if xerr != nil {
		return []abstract.HostTemplate{}, xerr
	}
	return allTemplates, nil
}

// ListImages ...
// Value of all has no impact on the result
func (p provider) ListImages(all bool) ([]abstract.Image, fail.Error) {
	if p.IsNull() {
		return []abstract.Image{}, fail.InvalidInstanceError()
	}

	allImages, xerr := p.Stack.(api.ReservedForProviderUse).ListImages()
	if xerr != nil {
		return nil, xerr
	}
	return allImages, nil
}

// GetName returns the providerName
func (p provider) GetName() string {
	return "cloudferro"
}

// GetTenantParameters returns the tenant parameters as-is
func (p provider) GetTenantParameters() map[string]interface{} {
	if p.IsNull() {
		return map[string]interface{}{}
	}

	return p.tenantParameters
}

// GetCapabilities returns the capabilities of the provider
func (p *provider) GetCapabilities() providers.Capabilities {
	if p.IsNull() {
		return providers.Capabilities{}
	}

	return providers.Capabilities{
		PrivateVirtualIP: true,
	}
}

func init() {
	iaas.Register("cloudferro", &provider{})
}
