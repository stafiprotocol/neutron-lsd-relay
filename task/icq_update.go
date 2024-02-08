package task

func (t *Task) checkIcqSubmitHeight(icaAddr, queryKind string, lastStepHeight uint64) (uint64, bool) {
	query, err := t.getRegisteredIcqQuery(icaAddr, queryKind)
	if err != nil {
		return 0, false
	}
	if query.RegisteredQuery.LastSubmittedResultLocalHeight <= lastStepHeight {
		return query.RegisteredQuery.LastSubmittedResultLocalHeight, false
	}

	return query.RegisteredQuery.LastSubmittedResultLocalHeight, true
}
