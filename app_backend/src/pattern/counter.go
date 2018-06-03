package pattern

import (
	"bufio"
	"strings"
	"time"

	_ "github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

func CountPatterns(scanner *bufio.Scanner, patterns []*Pattern) error {
	var seenUsers map[string]bool = make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()
		splits := strings.Split(line, ",")
		userId, eventName := splits[0], splits[2]
		userCreatedTime, err := time.Parse(time.RFC3339, splits[1])
		if err != nil {
			log.Fatal(err)
		}
		eventCreatedTime, err := time.Parse(time.RFC3339, splits[3])
		if err != nil {
			log.Fatal(err)
		}

		_, isSeenUser := seenUsers[userId]
		for _, p := range patterns {
			if !isSeenUser {
				if err = p.ResetForNewUser(userId, userCreatedTime); err != nil {
					log.Fatal(err)
				}
			}
			if err = p.CountForEvent(eventName, eventCreatedTime, userId, userCreatedTime); err != nil {
				log.Error(err)
			}
		}
		seenUsers[userId] = true
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return nil
}
