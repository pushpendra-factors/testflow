import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Card,
  Col,
  Row,
} from 'reactstrap';

import { fetchFactors } from "../../actions/factorsActions";
import { fetchCurrentProjectEvents } from "../../actions/projectsActions";
import BarChartCard from './BarChartCard.js';
import LineChartCard from './LineChartCard.js';
import FunnelChartCard from './FunnelChartCard.js';
import QueryBuilderCard from './QueryBuilderCard';
import {
  ALLOW_NUMBER_CREATE, ALLOW_STRING_CREATE,
  DYNAMIC_FETCH_EVENT_PROPERTIES, DYNAMIC_FETCH_EVENT_PROPERTY_VALUES,
  NUMERICAL_VALUE_TYPE, STRING_VALUE_TYPE,
  STATE_EVENTS, STATE_PROPERTY_TYPE,
  STATE_EVENT_PROPERTY_NAME, STATE_USER_PROPERTY_NAME,
  STATE_EVENT_NUMERIC_PROPERTY_VALUE, STATE_EVENT_STRING_PROPERTY_VALUE,
  STATE_USER_NUMERIC_PROPERTY_VALUE, STATE_USER_STRING_PROPERTY_VALUE
} from './QueryBuilderCard';

const EVENT_NAME_TYPE = "eventName";
const EVENT_PROPERTY_START_TYPE = "eventPropertyStart";
const USER_PROPERTY_STARTY_TYPE = "userPropertyStart";
const TO_TYPE = "to";
const EVENT_PROPERTY_NAME_TYPE = "eventPropertyName";
const USER_PROPERTY_NAME_TYPE = "userPropertyName";

const OPERATOR_EQUALS = "equals";
const OPERATOR_GREATER_THAN = "greaterThan";
const OPERATOR_LESSER_THAN = "lesserThan";

const cardColumnSetting = {
  size: '10',
  offset: '1'
};

const chartCardRowStyle = {
  marginTop: '50px',
  marginBottom: '2px',
  marginRight: '2px',
  marginLeft: '2px',
};

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    factors: store.factors.factors,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ fetchFactors, fetchCurrentProjectEvents }, dispatch);
}

class Factor extends Component {
  constructor(props) {
    super(props);
    this.toggle = this.toggle.bind(this);
    this.toggleFade = this.toggleFade.bind(this);
    this.factor = this.factor.bind(this);

    this.state = {
      collapse: true,
      fadeIn: true,
      timeout: 300,

      eventNames: {
        loaded: false,
        error: null
      }
    }
  }

  componentWillMount() {
    this.props.fetchCurrentProjectEvents(this.props.currentProjectId)
      .then((response) => {
        this.setState({ eventNames: { loaded: true } });
      })
      .catch((response) => {
        this.setState({ eventNames: { loaded: true, error: response.payload } });
      });
  }

  getEventPropertiesOptions(eventProperties) {
    var lp = [];
    var categoricalProperties = eventProperties["categorical"];
    categoricalProperties.forEach(function (categoricalProperty) {
      lp.push(
        {
          label: categoricalProperty + " equals", value: lp.length + 1, currentState: STATE_EVENT_PROPERTY_NAME,
          nextState: STATE_EVENT_STRING_PROPERTY_VALUE,
          type: EVENT_PROPERTY_NAME_TYPE, property: categoricalProperty, operator: OPERATOR_EQUALS
        });
    });
    var numericalProperties = eventProperties["numerical"];
    numericalProperties.forEach(function (numericalProperty) {
      lp.push(
        {
          label: numericalProperty + " equals", value: lp.length + 1, currentState: STATE_EVENT_PROPERTY_NAME,
          nextState: STATE_EVENT_NUMERIC_PROPERTY_VALUE,
          type: EVENT_PROPERTY_NAME_TYPE, property: numericalProperty, operator: OPERATOR_EQUALS
        });
      lp.push(
        {
          label: numericalProperty + " greater than", value: lp.length + 1, currentState: STATE_EVENT_PROPERTY_NAME,
          nextState: STATE_EVENT_NUMERIC_PROPERTY_VALUE,
          type: EVENT_PROPERTY_NAME_TYPE, property: numericalProperty, operator: OPERATOR_GREATER_THAN
        });
      lp.push(
        {
          label: numericalProperty + " lesser than", value: lp.length + 1, currentState: STATE_EVENT_PROPERTY_NAME,
          nextState: STATE_EVENT_NUMERIC_PROPERTY_VALUE,
          type: EVENT_PROPERTY_NAME_TYPE, property: numericalProperty, operator: OPERATOR_LESSER_THAN
        });
    });
    return lp;
  }

  getEventPropertyValueOptions(eventPropertyValues) {
    var lp = [];
    eventPropertyValues.forEach(function (propertyValue) {
      lp.push(
        { label: propertyValue, value: lp.length + 1, currentState: STATE_EVENT_STRING_PROPERTY_VALUE,
          nextState: STATE_PROPERTY_TYPE, 'type': STRING_VALUE_TYPE },
        );
    });
    return lp;
  }

  getQueryStates(projectEventNames) {
    const queryStates = {
      [STATE_EVENTS]: {
        'labels': (projectEventNames ? (
          Array.from(projectEventNames,
            eventName => ({
              'label': eventName, 'value': eventName,
              currentState: STATE_EVENTS, nextState: STATE_PROPERTY_TYPE, type: EVENT_NAME_TYPE
            }))) :
          []),
      },
      [STATE_PROPERTY_TYPE]: {
        'labels': [
          { label: 'Escape and Enter to close and search', isDisabled: true },
          { label: 'to', value: 1, currentState: STATE_PROPERTY_TYPE, nextState: STATE_EVENTS, onlyOnce: true, type: TO_TYPE },
          { label: 'with event property', value: 2, currentState: STATE_PROPERTY_TYPE, nextState: STATE_EVENT_PROPERTY_NAME, type: EVENT_PROPERTY_START_TYPE },
          { label: 'with user property', value: 3, currentState: STATE_PROPERTY_TYPE, nextState: STATE_USER_PROPERTY_NAME, type: USER_PROPERTY_STARTY_TYPE },
        ],
      },
      [STATE_EVENT_PROPERTY_NAME]: {
        'labels': [],
        [DYNAMIC_FETCH_EVENT_PROPERTIES]: true,
      },
      [STATE_USER_PROPERTY_NAME]: {
        'labels': [
          {
            label: 'country equals', value: 1, currentState: STATE_USER_PROPERTY_NAME, nextState: STATE_USER_STRING_PROPERTY_VALUE,
            type: USER_PROPERTY_NAME_TYPE, property: 'country', operator: OPERATOR_EQUALS
          },
          {
            label: 'age equals', value: 2, currentState: STATE_USER_PROPERTY_NAME, nextState: STATE_USER_NUMERIC_PROPERTY_VALUE,
            type: USER_PROPERTY_NAME_TYPE, property: 'age', operator: OPERATOR_EQUALS
          },
          {
            label: 'age greater than', value: 3, currentState: STATE_USER_PROPERTY_NAME, nextState: STATE_USER_NUMERIC_PROPERTY_VALUE,
            type: USER_PROPERTY_NAME_TYPE, property: 'age', operator: OPERATOR_GREATER_THAN
          },
          {
            label: 'age lesser than', value: 4, currentState: STATE_USER_PROPERTY_NAME, nextState: STATE_USER_NUMERIC_PROPERTY_VALUE,
            type: USER_PROPERTY_NAME_TYPE, property: 'age', operator: OPERATOR_LESSER_THAN
          },
        ],
      },
      [STATE_EVENT_NUMERIC_PROPERTY_VALUE]: {
        'labels': [
          { label: 'Enter number', 'value': 0, isDisabled: true, 'type': NUMERICAL_VALUE_TYPE },
        ],
        [ALLOW_NUMBER_CREATE]: true,
        'nextState': STATE_PROPERTY_TYPE,
      },
      [STATE_EVENT_STRING_PROPERTY_VALUE]: {
        'labels': [],
        [ALLOW_STRING_CREATE]: true,
        [DYNAMIC_FETCH_EVENT_PROPERTY_VALUES]: true,
        'nextState': STATE_PROPERTY_TYPE,
      },
      [STATE_USER_NUMERIC_PROPERTY_VALUE]: {
        'labels': [
          { label: 'Enter number', 'value': 0, isDisabled: true, 'type': NUMERICAL_VALUE_TYPE },
        ],
        [ALLOW_NUMBER_CREATE]: true,
        'nextState': STATE_PROPERTY_TYPE,
      },
      [STATE_USER_STRING_PROPERTY_VALUE]: {
        'labels': [
          { label: 'India', value: 1, currentState: STATE_USER_STRING_PROPERTY_VALUE, nextState: STATE_PROPERTY_TYPE, 'type': STRING_VALUE_TYPE },
          { label: 'United States', value: 2, currentState: STATE_USER_STRING_PROPERTY_VALUE, nextState: STATE_PROPERTY_TYPE, 'type': STRING_VALUE_TYPE },
          { label: 'France', value: 3, currentState: STATE_USER_STRING_PROPERTY_VALUE, nextState: STATE_PROPERTY_TYPE, 'type': STRING_VALUE_TYPE }
        ],
        [ALLOW_STRING_CREATE]: true,
        'nextState': STATE_PROPERTY_TYPE,
      },
    };
    return queryStates;
  }

  toggle() {
    this.setState({ collapse: !this.state.collapse });
  }

  toggleFade() {
    this.setState((prevState) => { return { fadeIn: !prevState } });
  }

  factor = (queryElements) => {
    console.log('Factor ' + JSON.stringify(this.state.values));

    var query = {
      eventsWithProperties: [],
    }

    var nextExpectedTypes = [EVENT_NAME_TYPE];

    queryElements.forEach(function (queryElement) {
      if (nextExpectedTypes.length > 0 &&
        nextExpectedTypes.indexOf(queryElement.type) < 0) {
        console.error("Invalid Query: " + JSON.stringify(query));
        return;
      }

      switch (queryElement.type) {
        case EVENT_NAME_TYPE:
          // Create a new event and add it to query.
          var newEvent = {}
          newEvent["name"] = queryElement.label;
          newEvent["properties"] = [];
          newEvent["userProperties"] = [];
          query.eventsWithProperties.push(newEvent);
          nextExpectedTypes = [];
          break;
        case EVENT_PROPERTY_START_TYPE:
          // Create a new event property condition.
          var newEventProperty = {}
          numEvents = query.eventsWithProperties.length;
          query.eventsWithProperties[numEvents - 1].properties.push(newEventProperty);
          nextExpectedTypes = [EVENT_PROPERTY_NAME_TYPE];
          break;
        case USER_PROPERTY_STARTY_TYPE:
          // Create a new user property condition type.
          var newUserProperty = {}
          var numEvents = query.eventsWithProperties.length;
          query.eventsWithProperties[numEvents - 1].userProperties.push(newUserProperty);
          nextExpectedTypes = [USER_PROPERTY_NAME_TYPE];
          break;
        case TO_TYPE:
          nextExpectedTypes = [EVENT_NAME_TYPE];
          break;
        case EVENT_PROPERTY_NAME_TYPE:
          var numEvents = query.eventsWithProperties.length;
          var currentEvent = query.eventsWithProperties[numEvents - 1];
          var numProperties = currentEvent.properties.length;
          var currentProperty = currentEvent.properties[numProperties - 1];
          currentProperty['property'] = queryElement.property;
          currentProperty['operator'] = queryElement.operator;
          nextExpectedTypes = [NUMERICAL_VALUE_TYPE, STRING_VALUE_TYPE];
          break;
        case USER_PROPERTY_NAME_TYPE:
          var numEvents = query.eventsWithProperties.length;
          var currentEvent = query.eventsWithProperties[numEvents - 1];
          var numProperties = currentEvent.userProperties.length;
          var currentUserProperty = currentEvent.userProperties[numProperties - 1];
          currentUserProperty['property'] = queryElement.property;
          currentUserProperty['operator'] = queryElement.operator;
          nextExpectedTypes = [NUMERICAL_VALUE_TYPE, STRING_VALUE_TYPE];
          break;
        case NUMERICAL_VALUE_TYPE:
          var numEvents = query.eventsWithProperties.length;
          var currentEvent = query.eventsWithProperties[numEvents - 1];
          var numProperties = currentEvent.userProperties.length;
          var currentUserProperty = currentEvent.userProperties[numProperties - 1];
          if (!currentUserProperty || currentUserProperty.hasOwnProperty('value')) {
            var numEvents = query.eventsWithProperties.length;
            var currentEvent = query.eventsWithProperties[numEvents - 1];
            var numEventProperties = currentEvent.properties.length;
            var currentEventProperty = currentEvent.properties[numEventProperties - 1];
            currentEventProperty['value'] = parseFloat(queryElement.label);
            currentEventProperty['type'] = "numerical"
          } else {
            currentUserProperty['value'] = parseFloat(queryElement.label);
            currentUserProperty['type'] = "numerical"
          }
          nextExpectedTypes = [];
          break;
        case STRING_VALUE_TYPE:
          var numEvents = query.eventsWithProperties.length;
          var currentEvent = query.eventsWithProperties[numEvents - 1];
          var numProperties = currentEvent.userProperties.length;
          var currentUserProperty = currentEvent.userProperties[numProperties - 1];
          if (!currentUserProperty || currentUserProperty.hasOwnProperty('value')) {
            var numEvents = query.eventsWithProperties.length;
            var currentEvent = query.eventsWithProperties[numEvents - 1];
            var numEventProperties = currentEvent.properties.length;
            var currentEventProperty = currentEvent.properties[numEventProperties - 1];
            currentEventProperty['value'] = queryElement.label;
            currentEventProperty['type'] = "categorical"
          } else {
            currentUserProperty['value'] = queryElement.label;
            currentUserProperty['type'] = "categorical"
          }
          nextExpectedTypes = [];
          break;
      }
    });

    if (nextExpectedTypes.length > 0) {
      console.error('Invalid Query: ' + JSON.stringify(query));
      return;
    }
    console.log('Fire Query: ' + JSON.stringify(query));
    this.props.fetchFactors(this.props.currentProjectId,
      { query: query }, this.props.location.search);
  }

  render() {
    if (!this.state.eventNames.loaded) return <div> Loading... </div>;

    var charts = [];
    let resultElements;
    if (!!this.props.factors.charts) {
      for (var i = 0; i < this.props.factors.charts.length; i++) {
        // note: we add a key prop here to allow react to uniquely identify each
        // element in this array. see: https://reactjs.org/docs/lists-and-keys.html
        var chartData = this.props.factors.charts[i];
        if (chartData.type === 'line') {
          charts.push(<Row style={chartCardRowStyle} key={i}><Col sm={cardColumnSetting}><LineChartCard chartData={chartData} /></Col></Row>)
        } else if (chartData.type === 'bar') {
          charts.push(<Row style={chartCardRowStyle} key={i}><Col sm={cardColumnSetting}><BarChartCard chartData={chartData} key={i} /></Col></Row>)
        } else if (chartData.type === 'funnel') {
          charts.push(<Row style={chartCardRowStyle} key={i}><Col sm={cardColumnSetting}><FunnelChartCard chartData={chartData} /></Col></Row>);
        }
      }
      resultElements = <Card className="fapp-card-border-none">{charts}</Card>;
    }


    return (
      <div className='animated fadeIn'>

        <div>
          <Row>
            <Col xs='12' md='12'>
              <QueryBuilderCard 
                getQueryStates={this.getQueryStates}
                getEventPropertiesOptions={this.getEventPropertiesOptions}
                getEventPropertyValueOptions={this.getEventPropertyValueOptions}
                onKeyDown={this.factor}
                holderText="Enter goal."
              />
            </Col>
          </Row>
        </div>

        {resultElements}

      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Factor);