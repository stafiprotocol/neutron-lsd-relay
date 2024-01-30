package task

import (
	"sync"

	"github.com/sirupsen/logrus"
)

var newEraUpdateFuncName = "NewEraUpdate"

func (t *Task) handleNewEraUpdate() error {
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
				_ = t.processPoolNewEraUpdate(addr)
			}(poolAddr)
		}
		wg.Wait()
		return nil
	}

	return t.processPoolNewEraUpdate(t.poolAddr)
}

func (t *Task) processPoolNewEraUpdate(poolAddr string) error {
	var err error
	poolInfo, err := t.getQueryPoolInfoRes(poolAddr)
	if err != nil {
		return err
	}

	if poolInfo.Status != ActiveEnded {
		return nil
	}

	_, timestamp, err := t.neutronClient.GetCurrentBLockAndTimestamp()
	if err != nil {
		return err
	}
	targetEra := uint64(timestamp)/poolInfo.EraSeconds + poolInfo.Offset

	// check targetEra to skip
	if targetEra <= poolInfo.Era {
		return nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"pool":      poolAddr,
		"targetEra": targetEra,
		"newEra":    poolInfo.Era + 1,
		"action":    newEraUpdateFuncName,
	})

	txHash, err := t.neutronClient.SendContractExecuteMsg(t.stakeManager, getEraUpdateMsg(poolAddr), nil)
	if err != nil {
		logger.Warnf("failed, err: %s \n", err.Error())
		return nil
	}

	logger.WithFields(logrus.Fields{
		"txHash": txHash,
	}).Infoln("success")

	return nil
}
