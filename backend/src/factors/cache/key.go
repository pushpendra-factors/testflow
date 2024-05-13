package cache

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Key struct {
	// any one must be set.
	ProjectID  int64
	ProjectUID string
	AgentUUID  string
	// Prefix - Helps better grouping and searching
	// i.e table_name + index_name
	Prefix string
	// Suffix - optional
	Suffix string
}

var (
	ErrorInvalidProject  = errors.New("invalid key project")
	ErrorInvalidPrefix   = errors.New("invalid key prefix")
	ErrorInvalidKey      = errors.New("invalid redis cache key")
	ErrorInvalidValues   = errors.New("invalid values to set")
	ErrorPartialFailures = errors.New("Partial failures in Set")
)

func NewKeyWithOnlyPrefix(prefix string) (*Key, error) {

	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}

	return &Key{Prefix: prefix}, nil
}

func NewKey(projectId int64, prefix string, suffix string) (*Key, error) {
	if projectId == 0 {
		return nil, ErrorInvalidProject
	}

	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}
	return &Key{ProjectID: projectId, Prefix: prefix, Suffix: suffix}, nil
}

func NewKeyWithAllProjectsSupport(projectId int64, prefix string, suffix string) (*Key, error) {
	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}
	return &Key{ProjectID: projectId, Prefix: prefix, Suffix: suffix}, nil
}

// NewKeyWithProjectUID - Uses projectUID as project scope on the key.
func NewKeyWithProjectUID(projectUID, prefix, suffix string) (*Key, error) {
	if projectUID == "" {
		return nil, ErrorInvalidProject
	}

	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}

	return &Key{ProjectUID: projectUID, Prefix: prefix, Suffix: suffix}, nil
}

func NewKeyWithAgentUID(agentUUID, prefix, suffix string) (*Key, error) {
	if agentUUID == "" {
		return nil, ErrorInvalidProject
	}

	if prefix == "" {
		return nil, ErrorInvalidPrefix
	}

	return &Key{AgentUUID: agentUUID, Prefix: prefix, Suffix: suffix}, nil
}

func (key *Key) Key() (string, error) {
	if key.ProjectID == 0 && key.ProjectUID == "" && key.AgentUUID == "" {
		return "", ErrorInvalidProject
	}

	if key.Prefix == "" {
		return "", ErrorInvalidPrefix
	}

	var projectScope string
	if key.ProjectID != 0 {
		projectScope = fmt.Sprintf("pid:%d", key.ProjectID)
	} else if key.ProjectUID != "" {
		projectScope = fmt.Sprintf("puid:%s", key.ProjectUID)
	} else {
		projectScope = fmt.Sprintf("auuid:%s", key.AgentUUID)
	}

	// key: i.e, event_names:user_last_event:pid:1:uid:1
	return fmt.Sprintf("%s:%s:%s", key.Prefix, projectScope, key.Suffix), nil
}

func (key *Key) KeyWithAllProjectsSupport() (string, error) {
	if key.Prefix == "" {
		return "", ErrorInvalidPrefix
	}

	var projectScope string
	if key.ProjectID == 0 {
		projectScope = "pid:*"
	} else {
		projectScope = fmt.Sprintf("pid:%d", key.ProjectID)
	}
	// key: i.e, event_names:user_last_event:pid:*:suffix
	return fmt.Sprintf("%s:%s:%s", key.Prefix, projectScope, key.Suffix), nil
}

func (key *Key) KeyWithOnlyPrefix() (string, error) {
	if key.Prefix == "" {
		return "", ErrorInvalidPrefix
	}

	// key: i.e, event_names:user_last_event:pid:*:suffix
	return fmt.Sprintf("%s", key.Prefix), nil
}

// KeyFromStringWithPid - Splits the cache key into prefix/suffix/projectid format
// Only for pid based cache
func KeyFromStringWithPid(key string) (*Key, error) {
	if key == "" {
		return nil, ErrorInvalidValues
	}
	cacheKey := Key{}
	keyPidSplit := strings.Split(key, ":pid:")
	if len(keyPidSplit) == 2 {
		projectIDSuffix := strings.SplitN(keyPidSplit[1], ":", 2)
		if len(projectIDSuffix) == 2 {
			cacheKey.Suffix = projectIDSuffix[1]
		}
		projectId, _ := strconv.Atoi(projectIDSuffix[0])
		cacheKey.ProjectID = int64(projectId)
		cacheKey.Prefix = keyPidSplit[0]
	}
	return &cacheKey, nil
}
