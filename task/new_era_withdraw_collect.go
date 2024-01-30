package task

import (
	"errors"
	"sync"

	"github.com/sirupsen/logrus"
)

var newEraWithdrawCollectFuncName = "NewEraWithdrawCollect"

func (t *Task) handleNewEraWithdrawCollect() error {
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
				_ = t.processPoolNewEraWithdrawCollect(addr)
			}(poolAddr)
		}
		wg.Wait()
		return nil
	}

	return t.processPoolNewEraWithdrawCollect(t.poolAddr)
}

func (t *Task) processPoolNewEraWithdrawCollect(poolAddr string) error {
	var err error

	poolInfo, err := t.getQueryPoolInfoRes(poolAddr)
	if err != nil {
		return err
	}

	if poolInfo.Status != EraStakeEnded {
		return nil
	}

	poolIca, err := t.getPoolIcaInfo(poolInfo.IcaId)
	if err != nil {
		return err
	}
	if len(poolIca) < 2 {
		return errors.New("ica data query failed")
	}

	logger := logrus.WithFields(logrus.Fields{
		"pool":   poolAddr,
		"action": newEraWithdrawCollectFuncName,
	})

	if !t.checkIcqSubmitHeight(poolIca[1].IcaAddr, BalancesQueryKind, poolInfo.EraSnapshot.LastStepHeight) {
		logger.Warnln("withdraw address balance interchain query not ready")
		return nil
	}

	txHash, err := t.neutronClient.SendContractExecuteMsg(t.stakeManager, getEraCollectWithdrawMsg(poolAddr), nil)
	if err != nil {
		logger.Warnf("failed, err: %s \n", err.Error())
		return err
	}
	logger.WithFields(logrus.Fields{
		"txHash": txHash,
	}).Infoln("success")

	return nil
}
