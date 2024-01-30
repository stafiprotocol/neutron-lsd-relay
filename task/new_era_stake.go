package task

import (
	"errors"
	"sync"

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
				_ = t.processPoolNewEraStake(addr)
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

	poolIca, err := t.getPoolIcaInfo(poolInfo.IcaId)
	if err != nil {
		return err
	}
	if len(poolIca) < 2 {
		return errors.New("ica data query failed")
	}

	if !t.checkIcqSubmitHeight(poolAddr, DelegationsQueryKind, poolInfo.EraSnapshot.LastStepHeight) {
		logger.Warnln("delegation interchain query not ready")
		return nil
	}

	txHash, err := t.neutronClient.SendContractExecuteMsg(t.stakeManager, getEraStakeMsg(poolAddr), nil)
	if err != nil {
		logger.Warnf("failed, err: %s \n", err.Error())
		return err
	}

	logger.WithFields(logrus.Fields{
		"txHash": txHash,
	}).Infoln("success")

	return nil
}
