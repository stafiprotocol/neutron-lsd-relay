package task_test

import (
	"github.com/stafiprotocol/neutron-lsd-relay/task"
	"testing"
)

func TestMsg(t *testing.T) {
	t.Log(string(task.GetEraStakeMsg("cosmos17n0n04nsefgkjer0t2j97cqfpyt0vpnewp8xmk8r70rzryfxxtmq3350af")))
}
