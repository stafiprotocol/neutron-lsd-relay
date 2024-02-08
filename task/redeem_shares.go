package task

import (
	"github.com/sirupsen/logrus"
	"github.com/stafiprotocol/neutron-lsd-relay/pkg/utils"
)

var redeemSharesFuncName = "ExecuteRedeemShares"

func (t *Task) handleRedeemShares() error {
	if t.runForEntrustedPool {
		stackInfo, err := t.getStackInfoRes()
		if err != nil {
			return err
		}
		for _, pool := range stackInfo.Pools {
			if err := t.processPoolRedeemShares(pool); err != nil {
				return err
			}
		}
		return nil
	}

	return t.processPoolRedeemShares(t.poolAddr)
}

func (t *Task) processPoolRedeemShares(poolAddr string) error {
	poolInfo, err := t.getQueryPoolInfoRes(poolAddr)
	if err != nil {
		return err
	}
	if !poolInfo.LsmSupport {
		return nil
	}
	if len(poolInfo.ShareTokens) > 0 {
		var coins []Coin
		for _, k := range poolInfo.ShareTokens {
			shareToken := k
			if utils.ContainsString(poolInfo.RedeemmingShareTokenDenom, shareToken.Denom) {
				continue
			}
			coins = append(coins, shareToken)
		}
		logger := logrus.WithFields(logrus.Fields{
			"pool":   poolAddr,
			"action": redeemSharesFuncName,
		})
		msg := getRedeemTokenForShareMsg(t.poolAddr, coins)

		t.txMutex.Lock()
		txHash, err := t.neutronClient.SendContractExecuteMsg(t.stakeManager, msg, nil)
		if err != nil {
			t.txMutex.Unlock()

			logger.Warnf("failed, err: %s \n", err.Error())
			return err
		}
		t.txMutex.Unlock()

		logger.WithFields(logrus.Fields{
			"txHash": txHash,
		}).Infoln("success")
	}
	return nil
}
