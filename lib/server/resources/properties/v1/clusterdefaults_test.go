package propertiesv1

import (
    "reflect"
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/CS-SI/SafeScale/lib/server/resources/abstract"
)

func TestDefaults_Clone(t *testing.T) {
    ct := newClusterDefaults()
    ct.Image = "something"
    ct.GatewaySizing = abstract.HostEffectiveSizing{
        RAMSize: 3,
        GPUType: "NVidia",
    }

    clonedCt, ok := ct.Clone().(*ClusterDefaults)
    if !ok {
        t.Fail()
    }

    assert.Equal(t, ct, clonedCt)
    clonedCt.GatewaySizing.GPUNumber = 7
    clonedCt.GatewaySizing.GPUType = "Culture"

    areEqual := reflect.DeepEqual(ct, clonedCt)
    if areEqual {
        t.Error("It's a shallow clone !")
        t.Fail()
    }
}
