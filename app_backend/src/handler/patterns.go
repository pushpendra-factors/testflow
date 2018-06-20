package handler

import (
	C "config"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Test command.
// curl -i -X GET http://localhost:8080/projects/1/patterns?start_event=login&end_event=payment
func QueryPatternsHandler(c *gin.Context) {
	_, err := strconv.ParseUint(c.Params.ByName("project_id"), 10, 64)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	qParams := c.Request.URL.Query()

	var startEvent string = ""
	startEvents := qParams["start_event"]
	if startEvents != nil {
		startEvent = startEvents[0]
	}
	var endEvent string = ""
	endEvents := qParams["end_event"]
	if endEvents != nil {
		endEvent = endEvents[0]
	}

	ps := C.GetServices().PatternService
	if patterns, err := ps.Query(startEvent, endEvent); err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Patterns query failed.")
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		type result struct {
			EventNames     []string  `json:"event_names"`
			Timings        []float64 `json:"timings"`
			Cardinalities  []float64 `json:"cardinalities"`
			Repeats        []float64 `json:"repeats"`
			Counts         []uint    `json:"counts"`
			PerUserCounts  []uint    `json:"per_user_counts"`
			TotalUserCount uint      `json:"total_user_count"`
		}
		results := []result{}
		for _, p := range patterns {
			r := result{
				EventNames:     p.EventNames,
				Timings:        []float64{},
				Cardinalities:  []float64{},
				Repeats:        []float64{},
				Counts:         []uint{},
				PerUserCounts:  []uint{},
				TotalUserCount: p.UserCount}
			for i := 0; i < len(p.EventNames); i++ {
				r.Timings = append(r.Timings, p.Timings[i].Quantile(0.5))
				r.Repeats = append(r.Repeats, p.Repeats[i].Quantile(0.5))
				r.Cardinalities = append(r.Cardinalities, p.EventCardinalities[i].Quantile(0.5))
				subsequenceCount, ok := ps.GetCount(p.EventNames[:i+1])
				if !ok {
					log.Errorf(fmt.Sprintf(
						"Subsequence %s not as frequent as sequence %s",
						strings.Join(p.EventNames[:i+1], ","), p.String()))
					r.Counts = append(r.Counts, p.Count)
				} else {
					r.Counts = append(r.Counts, subsequenceCount)
				}

				subsequencePerUserCount, ok := ps.GetPerUserCount(p.EventNames[:i+1])
				if !ok {
					log.Errorf(fmt.Sprintf(
						"Subsequence %s not as frequent as sequence %s",
						strings.Join(p.EventNames[:i+1], ","), p.String()))
					r.PerUserCounts = append(r.Counts, p.OncePerUserCount)
				} else {
					r.PerUserCounts = append(r.Counts, subsequencePerUserCount)
				}
			}
			results = append(results, r)
		}
		c.JSON(http.StatusOK, results)
	}
}
