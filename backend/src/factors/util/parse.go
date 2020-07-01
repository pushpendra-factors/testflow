package util

import (
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

/**
 * This util method parses string int list like "1,2,3,4" and returns a map of [int][bool] which is set
 * true for all the int values in the list.
 * Input : *string -> "1,2,3,4"
 * Output: map[uint64]bool -> map[1]=true, map[2]=true...
 */
func GetIntBoolMapFromStringList(intListSepByComma *string) map[uint64]bool {

	stringList := strings.Split(*intListSepByComma, ",")
	idToBoolMap := make(map[uint64]bool)
	for _, pid := range stringList {
		if pid == "" {
			continue
		}
		if pidUint, err := strconv.ParseUint(pid, 10, 64); err == nil {
			idToBoolMap[pidUint] = true
		} else {
			log.WithError(err).Errorln("Failed to parse provided string list to skip", intListSepByComma)
			panic(err)
		}
	}
	return idToBoolMap
}
