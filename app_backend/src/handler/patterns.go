package handler

import (
	C "config"
	"net/http"
	"strconv"

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
			EventNames         []string
			Timings            []float64
			EventCardinalities []float64
			Repeats            []float64
			Count              uint
			OncePerUserCount   uint
			UserCount          uint
		}
		results := []result{}
		for _, p := range patterns {
			r := result{EventNames: p.EventNames,
				Timings:            []float64{},
				EventCardinalities: []float64{},
				Repeats:            []float64{},
				Count:              p.Count,
				OncePerUserCount:   p.OncePerUserCount,
				UserCount:          p.UserCount}
			for i := 0; i < len(p.EventNames); i++ {
				r.Timings = append(r.Timings, p.Timings[i].Mean())
				r.Repeats = append(r.Repeats, p.Repeats[i].Mean())
				r.EventCardinalities = append(r.EventCardinalities, p.EventCardinalities[i].Mean())
			}
			results = append(results, r)
		}
		c.JSON(http.StatusOK, results)
	}
}
