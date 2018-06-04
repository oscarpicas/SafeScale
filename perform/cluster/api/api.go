/*
 * Copyright 2018, CS Systemes d'Information, http://www.c-s.fr
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

package api

import (
	"encoding/gob"

	providerapi "github.com/CS-SI/SafeScale/providers/api"

	"github.com/CS-SI/SafeScale/perform/cluster/api/ClusterState"
	"github.com/CS-SI/SafeScale/perform/cluster/api/Complexity"
	"github.com/CS-SI/SafeScale/perform/cluster/api/Flavor"

	pb "github.com/CS-SI/SafeScale/broker"
)

//Request defines what kind of Cluster is wanted
type Request struct {
	//Name is the name of the cluster wanted
	Name string
	//CIDR defines the network to create
	CIDR string
	//Mode is the implementation wanted, can be Simple, HighAvailability or HighVolume
	Complexity Complexity.Enum
	//Flavor tells what kind of cluster to create
	Flavor Flavor.Enum
	//NetworkID is the ID of the network to use
	NetworkID string
	//Tenant contains the name of the tenant
	Tenant string
}

//ClusterAPI is an interface of methods associated to Cluster-like structs
type ClusterAPI interface {
	//Start starts the cluster
	Start() error
	//Stop stops the cluster
	Stop() error
	//GetState returns the current state of the cluster
	GetState() (ClusterState.Enum, error)
	//GetNetworkID returns the ID of the network used by the cluster
	GetNetworkID() string

	//AddNode adds a node
	AddNode(bool, *pb.VMDefinition) (*pb.VM, error)
	//DeleteLastNode deletes a node
	DeleteLastNode(bool) error
	//DeleteSpecificNode deletes a node identified by its ID
	DeleteSpecificNode(string) error
	//ListNodes lists the nodes in the cluster
	ListNodes(bool) []string
	//FindNode tells if the ID of the VM passed as parameter is a node
	SearchNode(string, bool) bool
	//GetNode returns a node based on its ID
	GetNode(string) (*pb.VM, error)
	//CountNodes counts the nodes of the cluster
	CountNodes(bool) uint

	//Delete allows to destroy infrastructure of cluster
	Delete() error

	//GetDefinition
	GetDefinition() Cluster
	//UpdateMetadata
	//UpdateMetadata() error
	//RemoveMetadata
	//RemoveMetadata() error
}

//Cluster contains the bare minimum information about a cluster
type Cluster struct {
	//Name is the name of the cluster
	Name string
	//CIDR is the network CIDR wanted for the Network
	CIDR string
	//Flavor tells what kind of cluster it is
	Flavor Flavor.Enum
	//Mode is the mode of cluster; can be Simple, HighAvailability, HighVolume
	Complexity Complexity.Enum
	//Keypair contains the key-pair used inside the Cluster
	Keypair *providerapi.KeyPair
	//State
	State ClusterState.Enum
	//Tenant is the name of the tenant
	Tenant string
	//NetworkID is the ID of the network to use
	NetworkID string
	//PublicNodedIDs is a slice of VMIDs of the public cluster nodes
	PublicNodeIDs []string
	//PrivateNodedIDs is a slice of VMIDs of the private cluster nodes
	PrivateNodeIDs []string
}

//GetNetworkID returns the ID of the Network used by the cluster
func (c *Cluster) GetNetworkID() string {
	return c.NetworkID
}

//CountNodes returns the number of public or private nodes in the cluster
func (c *Cluster) CountNodes(public bool) uint {
	if public {
		return uint(len(c.PublicNodeIDs))
	}
	return uint(len(c.PrivateNodeIDs))
}

func init() {
	gob.Register(Cluster{})
}