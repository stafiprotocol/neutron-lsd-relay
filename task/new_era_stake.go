package task

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
)

var newEraStakeFuncName = "NewEraStake"

func (t *Task) handleNewEraStake() error {
	if t.runForEntrustedPool {
		stackInfo, err := t.getStackInfoRes()
		if err != nil {
			return err
		}
		wg := sync.WaitGroup{}
		for _, poolAddr := range stackInfo.Pools {
			wg.Add(1)
			poolAddr := poolAddr
			go func(addr string) {
				defer wg.Done()
				err = t.processPoolNewEraStake(addr)
				if err != nil {
					logrus.Error(err)
				}
			}(poolAddr)
		}
		wg.Wait()
		return nil
	}

	return t.processPoolNewEraStake(t.poolAddr)
}

func (t *Task) processPoolNewEraStake(poolAddr string) error {
	poolInfo, err := t.getQueryPoolInfoRes(poolAddr)
	if err != nil {
		return err
	}

	if poolInfo.Status != EraUpdateEnded {
		return nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"pool":           poolAddr,
		"snapshotBond":   poolInfo.EraSnapshot.Bond,
		"snapshotUnbond": poolInfo.EraSnapshot.Unbond,
		"snapshotActive": poolInfo.EraSnapshot.Active,
		"action":         newEraStakeFuncName,
	})

	if submitHeight, ok := t.checkIcqSubmitHeight(poolAddr, DelegationsQueryKind, poolInfo.EraSnapshot.LastStepHeight); !ok {
		logger.Warnln("delegation interchain query not ready", "submitHeight", submitHeight)
		return nil
	}

	ibcFee, err := t.neutronClient.GetTotalIbcFee()
	if err != nil {
		return err
	}
	ibcFeeCoins := types.NewCoins(types.NewCoin(t.neutronClient.GetDenom(), ibcFee))
	t.txMutex.Lock()
	defer t.txMutex.Unlock()

	txHash, err := t.neutronClient.SendContractExecuteMsg(t.stakeManager, GetEraStakeMsg(poolAddr), ibcFeeCoins)
	if err != nil {
		logger.Warnf("failed, err: %s \n", err.Error())
		return err
	}

	logger.WithFields(logrus.Fields{
		"txHash": txHash,
	}).Infoln("success")

	return nil
}
