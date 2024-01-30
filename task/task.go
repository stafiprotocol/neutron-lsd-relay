package task

import (
	"errors"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/stafiprotocol/neutron-lsd-relay/pkg/config"
	"github.com/stafiprotocol/neutron-lsd-relay/pkg/utils"

	"github.com/stafihub/neutron-relay-sdk/client"
	"github.com/stafihub/neutron-relay-sdk/common/log"

	"github.com/sirupsen/logrus"
)

type Task struct {
	taskTicker          uint32
	stop                chan struct{}
	neutronClient       *client.Client
	runForEntrustedPool bool
	poolAddr            string
	stakeManager        string
	handlers            []Handler
}

type Handler struct {
	method func() error
	name   string
}

func NewTask(cfg *config.Config) (*Task, error) {
	if cfg.StakeManager == "" {
		return nil, errors.New("stake manager is empty")
	}
	if cfg.PoolAddr == "" && !cfg.RunForEntrustedPool {
		return nil, errors.New("pool addr is empty")
	}
	t := &Task{
		taskTicker:          cfg.TaskTicker,
		stop:                make(chan struct{}),
		poolAddr:            cfg.PoolAddr,
		stakeManager:        cfg.StakeManager,
		runForEntrustedPool: cfg.RunForEntrustedPool,
	}

	kr, err := keyring.New("neutron", cfg.BackendOptions, cfg.KeystorePath, os.Stdin, client.MakeEncodingConfig().Marshaler)
	if err != nil {
		return nil, err
	}

	c, err := client.NewClient(kr, cfg.KeyName, cfg.GasPrice, "neutron", cfg.EndpointList, log.NewLog("client", "lsd-relay"))
	if err != nil {
		return nil, err
	}
	t.neutronClient = c

	t.handlers = append(t.handlers, Handler{
		method: t.handleNewEraUpdate,
		name:   newEraUpdateFuncName,
	})
	t.handlers = append(t.handlers, Handler{
		method: t.handleNewEraStake,
		name:   newEraStakeFuncName,
	})
	t.handlers = append(t.handlers, Handler{
		method: t.handleNewEraWithdrawCollect,
		name:   newEraWithdrawCollectFuncName,
	})
	t.handlers = append(t.handlers, Handler{
		method: t.handleNewEraRestake,
		name:   newEraRestakeFuncName,
	})
	t.handlers = append(t.handlers, Handler{
		method: t.handleNewEraActive,
		name:   eraActiveFuncName,
	})
	t.handlers = append(t.handlers, Handler{
		method: t.handleIcqUpdate,
		name:   icqUpdateFuncName,
	})
	t.handlers = append(t.handlers, Handler{
		method: t.handleRedeemShares,
		name:   redeemSharesFuncName,
	})

	return t, nil
}

func (t *Task) Start() error {
	utils.SafeGoWithRestart(t.handler)
	return nil
}

func (t *Task) Stop() {
	close(t.stop)
}

func (t *Task) handler() {
	logrus.Info("start handlers")

Out:
	for {
		select {
		case <-t.stop:
			logrus.Info("task has stopped")
			return
		default:

			for _, handler := range t.handlers {
				funcName := handler.name
				logrus.Debugf("handler %s start...", funcName)

				err := handler.method()
				if err != nil {
					logrus.Warnf("handler %s failed: %s, will retry.", funcName, err)
					time.Sleep(time.Second * 6)
					continue Out
				}
				logrus.Debugf("handler %s end", funcName)
			}
		}

		time.Sleep(time.Duration(t.taskTicker) * time.Second)
	}
}
