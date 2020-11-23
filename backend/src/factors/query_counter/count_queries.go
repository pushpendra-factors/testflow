package querycounter

import (
	"bufio"
	"encoding/json"
	M "factors/model"
	P "factors/pattern"

	log "github.com/sirupsen/logrus"
)

type StreamingQueryResult struct {
	Metric float64
}

func CountQueries(scanner *bufio.Scanner, queries map[string]M.Query) (map[string]StreamingQueryResult, error) {
	// A tracker for each query.
	queryTrackers := map[string]QueryTracker{}
	queryResults := map[string]StreamingQueryResult{}

	for qid, query := range queries {
		queryTracker, err := NewQueryTracker(&query)
		if err != nil || queryTracker == nil {
			log.WithError(err).WithFields(
				log.Fields{"query_id": qid, "query": query}).Error(
				"Failed to initialize query tracker.")
		}
		queryTrackers[qid] = *queryTracker
	}

	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		var eventDetails P.CounterEventFormat

		if err := json.Unmarshal([]byte(line), &eventDetails); err != nil {
			log.WithFields(log.Fields{"line": line, "err": err}).Fatal("Read failed.")
			return queryResults, err
		}

		for qid, qt := range queryTrackers {
			if err := (&qt).CountForEvent(&eventDetails); err != nil {
				log.WithError(err).WithFields(log.Fields{
					"Query": queries[qid],
					"Event": eventDetails,
				}).Error("Error when counting event")
			}
		}
	}
	for qid, qt := range queryTrackers {
		queryResults[qid] = StreamingQueryResult{
			Metric: qt.count,
		}
	}
	return queryResults, nil
}
