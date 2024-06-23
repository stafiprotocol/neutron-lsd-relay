package task

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/types"
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
				err = t.processPoolNewEraWithdrawCollect(addr)
				if err != nil {
					logrus.Error(err)
				}
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

	logger := logrus.WithFields(logrus.Fields{
		"pool":   poolAddr,
		"action": newEraWithdrawCollectFuncName,
	})

	if subHeight, ok := t.checkIcqSubmitHeight(poolIca.WithdrawAddressIcaInfo.IcaAddr, BalancesQueryKind, poolInfo.EraSnapshot.LastStepHeight); !ok {
		logger.Warnln("withdraw address balance interchain query not ready", "submitHeight", subHeight)
		return nil
	}

	ibcFee, err := t.neutronClient.GetTotalIbcFee()
	if err != nil {
		return err
	}
	ibcFeeCoins := types.NewCoins(types.NewCoin(t.neutronClient.GetDenom(), ibcFee))

	t.txMutex.Lock()
	defer t.txMutex.Unlock()

	txHash, err := t.neutronClient.SendContractExecuteMsg(t.stakeManager, getEraCollectWithdrawMsg(poolAddr), ibcFeeCoins)
	if err != nil {
		logger.Warnf("failed, err: %s \n", err.Error())
		return err
	}
	logger.WithFields(logrus.Fields{
		"txHash": txHash,
	}).Infoln("success")

	return nil
}
