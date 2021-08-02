package tests

import (
	"factors/delta"
	P "factors/pattern"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventMatchCriteriaEventPropertiesCategorical(t *testing.T) {
	event := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E1",
		EventTimestamp:    1,
		EventCardinality:  0,
		EventProperties: map[string]interface{}{
			"ep1": "v1",
			"ep2": 1,
			"ep3": 1627290649,
		},
		UserProperties: map[string]interface{}{
			"up1": "v1",
			"up2": 1,
			"up3": 1627290649,
		},
	}
	// No fitlers but just event - positive case
	criteria := delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
	}
	result := delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	// No fitlers but just event - negative case
	criteria = delta.EventCriterion{
		Name:         "E2",
		EqualityFlag: true,
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	// One categorical filter - positive - equals
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "v1",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	// One categorical filter - negative - not equals
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "notEquals",
					Value:     "v1",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	// One categorical filter - negative - equals
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "v2",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	// One categorical filter - positive - not equals
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "notEqual",
					Value:     "v2",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	// One categorical filter - positive - contains
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "contains",
					Value:     "V",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	// One categorical filter - negative - not contains
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "notContains",
					Value:     "v",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	// One categorical filter - negative - equals
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "contains",
					Value:     "w",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	// One categorical filter - positive - not equals
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "notContains",
					Value:     "w",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	// One categorical filter - $none - equals - positive
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep4",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "$none",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	// One categorical filter - $none - notEquals - postive
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "notEqual",
					Value:     "$none",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	// One categorical filter - $none - equals - negative
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "$none",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	// One categorical filter - $none - notEquals - negative
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep4",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "notEqual",
					Value:     "$none",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	// One categorical filter - $none - notEquals - positive - different wrong category
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep2",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "notEqual",
					Value:     "$none",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	// multiple categorical filter - equals - positive
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "v1",
					LogicalOp: "AND",
				}, delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "v2",
					LogicalOp: "OR",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	// multiple categorical filter - notEquals - rejecting these queries - so returning false for these queries
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "notEqual",
					Value:     "v1",
					LogicalOp: "AND",
				}, delta.OperatorValueTuple{
					Operator:  "notEqual",
					Value:     "v2",
					LogicalOp: "OR",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	// multiple categorical filter - equals - negative - illogical filter(testing to make sure that query is handled)
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep1",
				Type: "categorical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "v1",
					LogicalOp: "AND",
				}, delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "v2",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	//single numerical filter - equals - positive
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep2",
				Type: "numerical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "1",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	//single numerical filter - notEquals - positive
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep2",
				Type: "numerical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "notEqual",
					Value:     "1",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	//single numerical filter - greater than - negative
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep2",
				Type: "numerical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "greaterThan",
					Value:     "1",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	//single numerical filter - lesser than - negative
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep2",
				Type: "numerical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "lesserThan",
					Value:     "2",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	//single numerical filter - llesserThanOrEqual - negative
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep2",
				Type: "numerical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "lesserThanOrEqual",
					Value:     "1",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	//single numerical filter - greaterThanOrEqual - negative
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep2",
				Type: "numerical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "greaterThanOrEqual",
					Value:     "1",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	//single datetime filter - equals - positive
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep3",
				Type: "datetime",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "{\"fr\":1627290648,\"to\":1627290650,\"ovp\":false}",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	//single datetime filter - equals - negative
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep3",
				Type: "datetime",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "{\"fr\":1627290650,\"to\":1627290651,\"ovp\":false}",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
	//multiple filter keys - positive
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep3",
				Type: "datetime",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "{\"fr\":1627290648,\"to\":1627290650,\"ovp\":false}",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			}, delta.EventFilterCriterion{
				Key:  "ep2",
				Type: "numerical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "lesserThan",
					Value:     "2",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, true, result)
	//multiple filter keys - positive
	criteria = delta.EventCriterion{
		Name:         "E1",
		EqualityFlag: true,
		FilterCriterionList: []delta.EventFilterCriterion{
			delta.EventFilterCriterion{
				Key:  "ep3",
				Type: "datetime",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "equals",
					Value:     "{\"fr\":1627290648,\"to\":1627290650,\"ovp\":false}",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			}, delta.EventFilterCriterion{
				Key:  "ep2",
				Type: "numerical",
				Values: []delta.OperatorValueTuple{delta.OperatorValueTuple{
					Operator:  "lesserThan",
					Value:     "1",
					LogicalOp: "AND",
				}},
				PropertiesMode: "event",
			},
		},
	}
	result = delta.EventMatchesCriterion(event, criteria)
	assert.Equal(t, false, result)
}

func TestQueryUser(t *testing.T) {
	event1 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E1",
		EventTimestamp:    1,
		EventCardinality:  0,
	}
	event2 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E2",
		EventTimestamp:    2,
		EventCardinality:  0,
	}
	event3 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E1",
		EventTimestamp:    2,
		EventCardinality:  0,
	}
	event4 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E2",
		EventTimestamp:    3,
		EventCardinality:  0,
	}
	event5 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E3",
		EventTimestamp:    4,
		EventCardinality:  0,
	}
	event6 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E4",
		EventTimestamp:    4,
		EventCardinality:  0,
	}
	event7 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E5",
		EventTimestamp:    5,
		EventCardinality:  0,
	}
	event8 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E6",
		EventTimestamp:    6,
		EventCardinality:  0,
	}
	session := delta.Session{
		Events: []P.CounterEventFormat{
			event1,
			event2,
			event3,
			event4,
			event5,
			event6,
			event7,
			event8,
		},
	}
	// E1 -> E2
	query :=
		delta.Query{
			Base: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E1",
						EqualityFlag: true,
					},
				},
			},
			Target: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E2",
						EqualityFlag: true,
					},
				},
			},
		}
	queryCriteria := delta.MakePerUserQueryResult(query)
	baseIndex, targetIndex := int64(-1), int64(-1)
	baseIt, targetIt := int(-1), int(-1)
	delta.QuerySession(session, query, &queryCriteria, &baseIndex, &targetIndex, &baseIt, &targetIt, 0)
	assert.Equal(t, baseIndex, int64(1))
	assert.Equal(t, targetIndex, int64(2))
	// E3 -> E4 out of order but same timestamp
	query =
		delta.Query{
			Base: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E4",
						EqualityFlag: true,
					},
				},
			},
			Target: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E3",
						EqualityFlag: true,
					},
				},
			},
		}
	queryCriteria = delta.MakePerUserQueryResult(query)
	baseIndex, targetIndex = int64(-1), int64(-1)
	baseIt, targetIt = int(-1), int(-1)
	delta.QuerySession(session, query, &queryCriteria, &baseIndex, &targetIndex, &baseIt, &targetIt, 0)
	assert.Equal(t, baseIndex, int64(4))
	assert.Equal(t, targetIndex, int64(4))
	// E6 -> E5 out of order
	query =
		delta.Query{
			Base: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E6",
						EqualityFlag: true,
					},
				},
			},
			Target: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E5",
						EqualityFlag: true,
					},
				},
			},
		}
	queryCriteria = delta.MakePerUserQueryResult(query)
	baseIndex, targetIndex = int64(-1), int64(-1)
	baseIt, targetIt = int(-1), int(-1)
	delta.QuerySession(session, query, &queryCriteria, &baseIndex, &targetIndex, &baseIt, &targetIt, 0)
	assert.Equal(t, baseIndex, int64(6))
	assert.Equal(t, targetIndex, int64(5))
	// E6 -> E5 out of order
	query =
		delta.Query{
			Base: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E6",
						EqualityFlag: true,
					},
				},
			},
			Target: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E7",
						EqualityFlag: true,
					},
				},
			},
		}
	queryCriteria = delta.MakePerUserQueryResult(query)
	baseIndex, targetIndex = int64(-1), int64(-1)
	baseIt, targetIt = int(-1), int(-1)
	delta.QuerySession(session, query, &queryCriteria, &baseIndex, &targetIndex, &baseIt, &targetIt, 0)
	assert.Equal(t, baseIndex, int64(6))
	assert.Equal(t, targetIndex, int64(-1))
	// E2 -> E1 target appears first but there is also a base -> target case
	query =
		delta.Query{
			Base: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E2",
						EqualityFlag: true,
					},
				},
			},
			Target: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E1",
						EqualityFlag: true,
					},
				},
			},
		}
	queryCriteria = delta.MakePerUserQueryResult(query)
	baseIndex, targetIndex = int64(-1), int64(-1)
	baseIt, targetIt = int(-1), int(-1)
	delta.QuerySession(session, query, &queryCriteria, &baseIndex, &targetIndex, &baseIt, &targetIt, 0)
	assert.Equal(t, baseIndex, int64(2))
	assert.Equal(t, targetIndex, int64(2))
}

func TestQueryUserMultiStepFunnel(t *testing.T) {
	event1 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E1",
		EventTimestamp:    1,
		EventCardinality:  0,
	}
	event2 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E1",
		EventTimestamp:    2,
		EventCardinality:  0,
	}
	event3 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E1",
		EventTimestamp:    3,
		EventCardinality:  0,
	}
	event4 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E1",
		EventTimestamp:    4,
		EventCardinality:  0,
	}
	event5 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E3",
		EventTimestamp:    4,
		EventCardinality:  0,
	}
	event6 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E4",
		EventTimestamp:    4,
		EventCardinality:  0,
	}
	event7 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E5",
		EventTimestamp:    5,
		EventCardinality:  0,
	}
	event8 := P.CounterEventFormat{
		UserId:            "1",
		UserJoinTimestamp: 1,
		EventName:         "E6",
		EventTimestamp:    6,
		EventCardinality:  0,
	}
	session1 := delta.Session{
		Events: []P.CounterEventFormat{
			event1,
			event5,
		},
	}
	session2 := delta.Session{
		Events: []P.CounterEventFormat{
			event2,
			event6,
		},
	}
	session3 := delta.Session{
		Events: []P.CounterEventFormat{
			event3,
			event7,
		},
	}
	session4 := delta.Session{
		Events: []P.CounterEventFormat{
			event4,
			event8,
		},
	}
	sessionList := []delta.Session{
		session1,
		session2,
		session3,
		session4,
	}
	// E1 -> E2
	query :=
		delta.MultiFunnelQuery{
			Base: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E1",
						EqualityFlag: true,
					},
				},
			},
			Intermediate: []delta.EventsCriteria{
				delta.EventsCriteria{
					Operator: "And",
					EventCriterionList: []delta.EventCriterion{
						delta.EventCriterion{
							Name:         "E1",
							EqualityFlag: true,
						},
					},
				},
				delta.EventsCriteria{
					Operator: "And",
					EventCriterionList: []delta.EventCriterion{
						delta.EventCriterion{
							Name:         "E1",
							EqualityFlag: true,
						},
					},
				},
			},
			Target: delta.EventsCriteria{
				Operator: "And",
				EventCriterionList: []delta.EventCriterion{
					delta.EventCriterion{
						Name:         "E3",
						EqualityFlag: true,
					},
				},
			},
		}
	result, _ := delta.QueryUserMultiStepFunnel(make([]P.CounterEventFormat, 0), sessionList, query)
	assert.Equal(t, true, result.BaseAndTargetFlag)
	assert.Equal(t, true, result.BaseFlag)
	assert.Equal(t, true, result.TargetFlag)
}
