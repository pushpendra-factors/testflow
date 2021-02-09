package querycounter

import (
	"factors/model/model"
	P "factors/pattern"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Query Tracker tracks a sequence of a group of events and properties (EventsPropertiesQueryTracker).
type QueryTracker struct {
	eventsSequenceTracker []EventsPropertiesQueryTracker
	// QueryType is one of QueryTypeEventsOccurrence or QueryTypeUniqueUsers.
	queryType         string
	groupByProperties []model.QueryGroupByProperty

	// private state variables.
	count         float64
	waitIndex     int
	prevUserId    string
	currUserId    string
	prevTimestamp int64
	currTimestamp int64
	numUsersSeen  int
}

func NewQueryTracker(query *model.Query) (*QueryTracker, error) {
	if query.Class == model.QueryClassInsights {
		var qTracker QueryTracker
		qTracker.queryType = query.Type
		qTracker.groupByProperties = query.GroupByProperties
		qTracker.count = 0.0
		qTracker.waitIndex = 0
		qTracker.prevUserId = ""
		qTracker.currUserId = ""
		qTracker.prevTimestamp = 0
		qTracker.currTimestamp = 0
		// No sequence. Just one query with the given events, properties and condition.
		epqTracker, err := NewEventPropertiesQueryTracker(
			query.EventsWithProperties,
			query.EventsCondition,
		)
		if err != nil {
			log.WithError(err).WithFields(
				log.Fields{"query": *query}).Info("Error Building query")
			return nil, err

		}
		qTracker.eventsSequenceTracker = []EventsPropertiesQueryTracker{*epqTracker}
		return &qTracker, nil
	} else if query.Class == model.QueryClassEvents {
		// TODO(aravind): This is in new dashboard. Check how this and query groups can be supported.
		var qTracker QueryTracker
		return &qTracker, nil
	} else if query.Class == model.QueryClassFunnel {
		var qTracker QueryTracker
		qTracker.queryType = query.Type
		qTracker.groupByProperties = query.GroupByProperties
		qTracker.eventsSequenceTracker = []EventsPropertiesQueryTracker{}
		// Each event with property is expected to occur in sequence.
		// Create a new EventsPropertiesQueryTracker for each event.
		for _, ewp := range query.EventsWithProperties {
			// For funnel queries since one event is tracked at a given point of time
			// either any or all will work. Without loss of generality we will set it as all.
			epqTracker, err := NewEventPropertiesQueryTracker(
				[]model.QueryEventWithProperties{ewp}, model.EventCondAllGivenEvent)
			if err != nil {
				log.WithError(err).WithFields(
					log.Fields{"query": *query}).Info("Error Building query")
				return nil, err

			}
			qTracker.eventsSequenceTracker = append(qTracker.eventsSequenceTracker, *epqTracker)
		}
		return &qTracker, nil
	}
	return nil, fmt.Errorf(fmt.Sprintf("Unsupported query type %s for tracking", query.Class))
}

func (qT *QueryTracker) ResetForNextTrack() {
	qT.waitIndex = 0
	for _, epqTracker := range qT.eventsSequenceTracker {
		epqTracker.Reset()
	}
}

// Updates the count based on incoming data point. Caller should call this method sequentially.
// eventDetails are expected to arrive in timestamp order, grouped by userId.
func (qT *QueryTracker) CountForEvent(eventDetails *P.CounterEventFormat) error {
	qLen := len(qT.eventsSequenceTracker)
	if qLen == 0 {
		log.Info("Nothing to count. Sequence of length 0")
		return nil
	}

	if eventDetails.UserId == "" || eventDetails.EventName == "" || eventDetails.EventTimestamp < 0 {
		err := fmt.Errorf(fmt.Sprintf("Invalid data: %s", eventDetails))
		log.WithError(err)
		return err
	}

	qT.prevUserId = qT.currUserId
	qT.currUserId = eventDetails.UserId

	qT.prevTimestamp = qT.currTimestamp
	qT.currTimestamp = eventDetails.EventTimestamp

	if qT.prevUserId != qT.currUserId && qT.currUserId != "" && qT.prevUserId != "" {
		qT.numUsersSeen++
		qT.prevTimestamp = 0
		if qT.queryType == model.QueryTypeUniqueUsers {
			qT.ResetForNextTrack()
		}
	}

	if qT.prevTimestamp > qT.currTimestamp {
		err := fmt.Errorf(fmt.Sprintf("Timestamp not in order for user. %d+", eventDetails))
		log.WithError(err)
		return err
	}

	if qT.waitIndex >= qLen && qT.queryType == model.QueryTypeUniqueUsers {
		// Has already completed an iteration for the current user. Return.
		// waitIndex will get reset on ResetOnNewUser.
		return nil
	}

	if qT.eventsSequenceTracker[qT.waitIndex].IsTrackCompleteOnData(eventDetails) {
		qT.waitIndex += 1
		qT.count += 1

		if qT.waitIndex == qLen && qT.queryType == model.QueryTypeEventsOccurrence {
			// Allow recounting for the same user, if of type events occurrence.
			qT.ResetForNextTrack()
		}
	}
	return nil
}
