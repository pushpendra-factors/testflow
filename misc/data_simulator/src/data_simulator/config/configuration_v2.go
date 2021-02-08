package config

/*
Input object corresponding to YAML config
*/

import (
	"time"
)

type AttributeDependency struct {
	Probablity float64
	Attributes map[string][]string
}

type AttributeRule struct {
	Real_time_wait     int
	Attribute_weights  []AttributeDependency
	Overall_probablity float64
}

type CorrelationMatrix struct {
	Events      map[string]map[string]interface{}
	Seed_events map[string]float64
	Exit_events []string
}

type EventProbablity struct {
	Correlation_matrix CorrelationMatrix
	Independent_events map[string]float64
}

type UserSegmentV2 struct {
	Number_of_users            int
	Activity_ticker_in_seconds int
	Activity_probablity_map    map[string]float64
	Event_probablity_map       EventProbablity
	Start_Time                 time.Time
	Event_attributes           EventAttributes
	Event_decorators           map[string]map[string]map[string]float64
	User_decorators            map[string]map[string]map[string]float64
	User_attributes            UserAttributes
	Set_attributes             bool
	Rules                      map[string]AttributeRule
	Smart_events               map[string]string
}

type EventAttributes struct {
	Predefined map[string]map[string]map[string]float64
	Default    []AttributeData
	Custom     []AttributeData
}

type UserAttributes struct {
	Default []AttributeData
	Custom  []AttributeData
}

type AttributeData struct {
	Key         string
	Order_Level int
	Values      map[string]interface{}
	Dependency  string
	Data_type   string
}

type ConfigurationV2 struct {
	User_data_file_name_prefix        string
	Activity_time_in_seconds          int
	Real_Time                         bool
	User_id_prefix                    string
	User_segments                     map[string]UserSegmentV2
	New_user_poll_time                int
	New_user_probablity               float64
	Per_tick_new_user_seed_count      int
	Custom_user_attribute_probablity  float64
	Custom_event_attribute_probablity float64
	Bring_existing_user               float64
}
