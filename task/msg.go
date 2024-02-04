package task

import (
	"encoding/json"
	"fmt"
)

// Status
const (
	EraUpdateStarted  = "era_update_started"
	EraUpdateEnded    = "era_update_ended"
	EraStakeStarted   = "era_stake_started"
	EraStakeEnded     = "era_stake_ended"
	WithdrawStarted   = "withdraw_started"
	WithdrawEnded     = "withdraw_ended"
	EraRestakeStarted = "era_restake_started"
	EraRestakeEnded   = "era_restake_ended"
	ActiveEnded       = "active_ended"
)

// QueryKind
const (
	BalancesQueryKind    = "balances"
	DelegationsQueryKind = "delegations"
	ValidatorsQueryKind  = "validators"
)

// ValidatorUpdateStatus
const (
	WaitQueryUpdate = "wait_query_update"
)

type PoolAddr struct {
	Addr string `json:"pool_addr"`
}

type QueryPoolInfoReq struct {
	PoolInfo PoolAddr `json:"pool_info"`
}

type StackInfoReq struct{}

type QueryPoolInfoRes struct {
	IcaId                     string      `json:"ica_id"`
	Era                       uint64      `json:"era"`
	EraSeconds                uint64      `json:"era_seconds"`
	Offset                    uint64      `json:"offset"`
	Bond                      string      `json:"bond"`
	Unbond                    string      `json:"unbond"`
	Active                    string      `json:"active"`
	Rate                      string      `json:"rate"`
	RateChangeLimit           string      `json:"rate_change_limit"`
	Status                    string      `json:"status"`
	ValidatorUpdateStatus     string      `json:"validator_update_status"`
	ShareTokens               []Coin      `json:"share_tokens"`
	RedeemmingShareTokenDenom []string    `json:"redeemming_share_token_denom"`
	EraSnapshot               eraSnapshot `json:"era_snapshot"`
	Paused                    bool        `json:"paused"`
	LsmSupport                bool        `json:"lsm_support"`
}

type RegisterQueryInfoRes struct {
	RegisteredQuery struct {
		Id    int    `json:"id"`
		Owner string `json:"owner"`
		Keys  []struct {
			Path string `json:"path"`
			Key  string `json:"key"`
		} `json:"keys"`
		QueryType                       string `json:"query_type"`
		TransactionsFilter              string `json:"transactions_filter"`
		ConnectionId                    string `json:"connection_id"`
		UpdatePeriod                    uint64 `json:"update_period"`
		LastSubmittedResultLocalHeight  uint64 `json:"last_submitted_result_local_height"`
		LastSubmittedResultRemoteHeight struct {
			RevisionNumber int `json:"revision_number"`
			RevisionHeight int `json:"revision_height"`
		} `json:"last_submitted_result_remote_height"`
		Deposit []struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"deposit"`
		SubmitTimeout      int `json:"submit_timeout"`
		RegisteredAtHeight int `json:"registered_at_height"`
	} `json:"registered_query"`
}

type ICAData struct {
	IcaAddr string `json:"ica_addr"`
}

type StackInfoRes struct {
	// Pools []string `json:"entrusted_pools"`
	Pools []string `json:"pools"`
}

type eraSnapshot struct {
	Era            uint64 `json:"era"`
	Bond           string `json:"bond"`
	Unbond         string `json:"unbond"`
	Active         string `json:"active"`
	RestakeAmount  string `json:"restake_amount"`
	LastStepHeight uint64 `json:"last_step_height"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type RedeemTokenForShareMsg struct {
	PoolAddr string `json:"pool_addr"`
	Tokens   []Coin `json:"tokens"`
}

type UpdateIcqUpdatePeriodMsg struct {
	Addr            string `json:"pool_addr"`
	NewUpdatePeriod uint64 `json:"new_update_period"`
}

func getQueryPoolInfoReq(poolAddr string) []byte {
	poolReq := QueryPoolInfoReq{
		PoolInfo: PoolAddr{
			Addr: poolAddr,
		},
	}
	marshal, _ := json.Marshal(poolReq)
	return marshal
}

func getEraUpdateMsg(poolAddr string) []byte {
	eraUpdateMsg := struct {
		PoolAddr `json:"era_update"`
	}{
		PoolAddr: PoolAddr{Addr: poolAddr},
	}
	marshal, _ := json.Marshal(eraUpdateMsg)
	return marshal
}

func GetEraStakeMsg(poolAddr string) []byte {
	eraBondMsg := struct {
		PoolAddr `json:"era_stake"`
	}{
		PoolAddr: PoolAddr{Addr: poolAddr},
	}
	marshal, _ := json.Marshal(eraBondMsg)
	return marshal
}

func getEraCollectWithdrawMsg(poolAddr string) []byte {
	eraCollectWithdrawMsg := struct {
		PoolAddr `json:"era_collect_withdraw"`
	}{
		PoolAddr: PoolAddr{Addr: poolAddr},
	}
	marshal, _ := json.Marshal(eraCollectWithdrawMsg)
	return marshal
}

func getEraRestakeMsg(poolAddr string) []byte {
	msg := struct {
		PoolAddr `json:"era_restake"`
	}{
		PoolAddr: PoolAddr{Addr: poolAddr},
	}
	marshal, _ := json.Marshal(msg)
	return marshal
}

func getEraActiveMsg(poolAddr string) []byte {
	eraActiveMsg := struct {
		PoolAddr `json:"era_active"`
	}{
		PoolAddr: PoolAddr{Addr: poolAddr},
	}
	marshal, _ := json.Marshal(eraActiveMsg)
	return marshal
}

func getPoolUpdateQueryExecuteMsg(poolAddr string) []byte {
	msg := struct {
		PoolAddr `json:"pool_update_validators_icq"`
	}{
		PoolAddr: PoolAddr{Addr: poolAddr},
	}
	marshal, _ := json.Marshal(msg)
	return marshal
}

func getRedeemTokenForShareMsg(poolAddr string, tokens []Coin) []byte {
	msg := struct {
		RedeemTokenForShareMsg `json:"redeem_token_for_share"`
	}{
		RedeemTokenForShareMsg: RedeemTokenForShareMsg{
			PoolAddr: poolAddr,
			Tokens:   tokens,
		},
	}
	marshal, _ := json.Marshal(msg)
	return marshal
}

func (t *Task) getQueryPoolInfoRes(poolAddr string) (*QueryPoolInfoRes, error) {
	poolInfoRes, err := t.neutronClient.QuerySmartContractState(t.stakeManager, getQueryPoolInfoReq(poolAddr))
	if err != nil {
		return nil, err
	}
	var res QueryPoolInfoRes
	err = json.Unmarshal(poolInfoRes.Data.Bytes(), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (t *Task) getRegisteredIcqQuery(icaAddr, queryKind string) (*RegisterQueryInfoRes, error) {
	msg := fmt.Sprintf("{\"get_ica_registered_query\":{\"ica_addr\":\"%s\",\"query_kind\":\"%s\"}}", icaAddr, queryKind)
	rawRes, err := t.neutronClient.QuerySmartContractState(t.stakeManager, []byte(msg))
	if err != nil {
		return nil, err
	}
	var res RegisterQueryInfoRes
	err = json.Unmarshal(rawRes.Data.Bytes(), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (t *Task) getPoolIcaInfo(icaId string) ([]ICAData, error) {
	msg := fmt.Sprintf("{\"interchain_account_address_from_contract\":{\"interchain_account_id\":\"%s\"}}", icaId)
	rawRes, err := t.neutronClient.QuerySmartContractState(t.stakeManager, []byte(msg))
	if err != nil {
		return nil, err
	}
	var res []ICAData
	_ = json.Unmarshal(rawRes.Data.Bytes(), &res)
	return res, nil
}

func (t *Task) getStackInfoRes() (*StackInfoRes, error) {
	msg := struct {
		StackInfoReq `json:"stack_info"`
	}{}
	marshal, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	stackInfoRes, err := t.neutronClient.QuerySmartContractState(t.stakeManager, marshal)
	if err != nil {
		return nil, err
	}
	var res StackInfoRes
	err = json.Unmarshal(stackInfoRes.Data.Bytes(), &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
