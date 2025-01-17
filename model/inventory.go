package model

import (
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"
)

type Inventory map[string]InventorySpec

func (inventory Inventory) Validate(ctx AppContext, spec *Spec) bool {
	for groupName, nodes := range inventory {
		if groupName == ctx.NodeGroup() {
			// check only groups that are used
			if !nodes.Validate(ctx, groupName) {
				return false
			}
		}
	}
	return true
}

func (inventory Inventory) GetClient(groupName string) (*rpc.Client, bool) {
	log.Printf("inventory: %+v", inventory)
	group, ok := inventory[groupName]
	if !ok {
		return nil, false
	}
	client, err := rpc.Dial(group[0])
	if err != nil {
		return nil, false
	}
	return client, true
}

type InventorySpec []string

func (spec *InventorySpec) Validate(ctx AppContext, groupName string) bool {
	validateLog := log.WithFields(log.Fields{
		"section": "Inventory",
		"group":   groupName,
	})
	for _, node := range *spec {
		client, err := rpc.Dial(node)
		if err != nil {
			validateLog.WithError(err).Warningln("failed to connect a Geth node")
			continue
		} else if err := client.Call(nil, "net_version"); err != nil {
			validateLog.WithError(err).Warningf("Geth node is limited")
			continue
		}
		client.Close()
		*spec = InventorySpec{
			node,
		}
		return true
	}
	validateLog.Errorln("live Geth nodes not found")
	return false
}
