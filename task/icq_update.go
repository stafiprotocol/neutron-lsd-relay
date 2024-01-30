package task

import "github.com/sirupsen/logrus"

var icqUpdateFuncName = "ExecuteValidatorsIcqUpdate"

func (t *Task) handleIcqUpdate() error {
	if t.runForEntrustedPool {
		stackInfo, err := t.getStackInfoRes()
		if err != nil {
			return err
		}
		for _, pool := range stackInfo.Pools {
			if err := t.processPoolIcqUpdate(pool); err != nil {
				return err
			}
		}
		return nil
	}

	return t.processPoolIcqUpdate(t.poolAddr)
}

func (t *Task) processPoolIcqUpdate(poolAddr string) error {
	poolInfo, err := t.getQueryPoolInfoRes(poolAddr)
	if err != nil {
		return err
	}
	if poolInfo.Status != WaitQueryUpdate {
		return nil
	}
	logger := logrus.WithFields(logrus.Fields{
		"pool":   poolAddr,
		"action": icqUpdateFuncName,
	})

	msg := getPoolUpdateQueryExecuteMsg(poolAddr)
	txHash, err := t.neutronClient.SendContractExecuteMsg(t.stakeManager, msg, nil)
	if err != nil {
		logger.Warnf("failed, err: %s \n", err.Error())
		return err
	}

	logger.WithFields(logrus.Fields{
		"txHash": txHash,
	}).Infoln("success")

	return nil
}

func (t *Task) checkIcqSubmitHeight(icaAddr, queryKind string, lastStepHeight uint64) bool {
	query, err := t.getRegisteredIcqQuery(icaAddr, queryKind)
	if err != nil {
		return false
	}
	if query.RegisteredQuery.LastSubmittedResultLocalHeight <= lastStepHeight {
		return false
	}

	return true
}
