package task

import (
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var eraActiveFuncName = "NewEraActive"

func (t *Task) handleNewEraActive() error {
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
				err = t.processPoolNewEraActive(addr)
				if err != nil {
					logrus.Error(err)
				}
			}(poolAddr)
		}
		wg.Wait()
		return nil
	}

	return t.processPoolNewEraActive(t.poolAddr)
}

func (t *Task) processPoolNewEraActive(poolAddr string) error {
	var err error

	poolInfo, err := t.getQueryPoolInfoRes(poolAddr)
	if err != nil {
		return err
	}

	if poolInfo.Status != EraRestakeEnded {
		return nil
	}

	logger := logrus.WithFields(logrus.Fields{
		"pool":    poolAddr,
		"oldRate": poolInfo.Rate,
		"action":  eraActiveFuncName,
	})

	if submitHeight, ok := t.checkIcqSubmitHeight(poolAddr, DelegationsQueryKind, poolInfo.EraSnapshot.LastStepHeight); !ok {
		logger.Warnln("delegation interchain query not ready", "submitHeight", submitHeight)
		return nil
	}

	t.txMutex.Lock()
	defer t.txMutex.Unlock()

	txHash, err := t.neutronClient.SendContractExecuteMsg(t.stakeManager, getEraActiveMsg(poolAddr), nil)
	if err != nil {
		logger.Warnf("failed, err: %s \n", err.Error())
		return err
	}

	retry := 0
	for {
		retry++
		if retry > 30 {
			logger.WithFields(logrus.Fields{
				"txHash": txHash,
			}).Warnln("tx success but result check been timeout")
			break
		}
		poolNewInfo, _ := t.getQueryPoolInfoRes(poolAddr)
		if poolNewInfo.Status == ActiveEnded {
			logger.WithFields(logrus.Fields{
				"active":  poolNewInfo.Active,
				"newRate": poolNewInfo.Rate,
				"txHash":  txHash,
			}).
				Infof("success(the new era task has been completed)")
			break
		}
		time.Sleep(3 * time.Second)
	}

	return nil
}
