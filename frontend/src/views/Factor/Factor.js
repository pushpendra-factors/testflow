import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Card, Col, Row } from 'reactstrap';
import moment from 'moment';

import { fetchFactors, resetFactors } from "../../actions/factorsActions";
import { fetchProjectEvents, fetchProjectModels } from "../../actions/projectsActions";
import BarChartCard from './BarChartCard.js';
import LineChartCard from './LineChartCard.js';
import FunnelChartCard from './FunnelChartCard.js';
import QueryBuilderCard from './QueryBuilderCard';
import {
  ALLOW_NUMBER_CREATE, ALLOW_STRING_CREATE,
  DYNAMIC_FETCH_EVENT_PROPERTIES, DYNAMIC_FETCH_EVENT_PROPERTY_VALUES,
  DYNAMIC_FETCH_USER_PROPERTIES, DYNAMIC_FETCH_USER_PROPERTY_VALUES,
  NUMERICAL_VALUE_TYPE, STRING_VALUE_TYPE, SUBMIT_QUERY_TYPE,
  STATE_EVENTS, STATE_PROPERTY_TYPE,
  STATE_EVENT_PROPERTY_NAME, STATE_USER_PROPERTY_NAME,
  STATE_EVENT_NUMERIC_PROPERTY_VALUE, STATE_EVENT_STRING_PROPERTY_VALUE,
  STATE_USER_NUMERIC_PROPERTY_VALUE, STATE_USER_STRING_PROPERTY_VALUE
} from './QueryBuilderCard';
import Loading from '../../loading';
import loadingImage from '../../assets/img/loading_g1.gif';


import Select from 'react-select';
import factorsai from '../../factorsaiObj';

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
  marginTop: '25px',
  marginBottom: '2px',
  marginRight: '2px',
  marginLeft: '2px',
};

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    factors: store.factors.factors,
    intervals: store.projects.intervals,
    defaultModelInterval: store.projects.defaultModelInterval,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ fetchFactors, resetFactors, fetchProjectEvents, fetchProjectModels }, dispatch);
}

const LOADING_DEFAULT = -1, LOADING_INIT = 0, LOADING_DONE = 1;

class Factor extends Component {
  constructor(props) {
    super(props);

    this.state = {
      collapse: true,
      fadeIn: true,
      timeout: 300,

      eventNames: {
        loaded: false,
        error: null
      },

      models: {
        loaded: false,
        error: null
      },

      factors: {
        loading: LOADING_DEFAULT
      },

      isModelSelectorDropdownOpen: false,
      selectedModelInterval: null
    }

  }

  componentWillMount() {
    if(!this.props.currentProjectId){
      return;
    }

    this.props.fetchProjectEvents(this.props.currentProjectId)
      .then(() => {
        this.setState({ eventNames: { loaded: true } });
      })
      .catch((r) => {
        this.setState({ eventNames: { loaded: true, error: r.payload } });
      });
      
      this.props.fetchProjectModels(this.props.currentProjectId)
        .then(() => this.setState({ models: { loaded: true } }));
  }

  componentDidUpdate() {
    if (this.props.defaultModelInterval != null && this.state.selectedModelInterval == null){ 
      this.setState({selectedModelInterval: this.props.defaultModelInterval});
    }
  }

  getPropertiesOptions(properties, isEventType) {
    var lp = [];

    if (properties.categorical == undefined && properties.numerical == undefined) {
      console.warn("Warning: No properties returned.")
      return lp;
    }

    var categoricalProperties = properties["categorical"];
    categoricalProperties.forEach(function (categoricalProperty) {
      lp.push(
        {
          label: categoricalProperty + " equals", value: lp.length + 1,
          currentState: isEventType ? STATE_EVENT_PROPERTY_NAME: STATE_USER_PROPERTY_NAME,
          nextState: isEventType ? STATE_EVENT_STRING_PROPERTY_VALUE: STATE_USER_STRING_PROPERTY_VALUE,
          type: isEventType ? EVENT_PROPERTY_NAME_TYPE: USER_PROPERTY_NAME_TYPE,
          property: categoricalProperty,
          operator: OPERATOR_EQUALS
        });
    });
    var numericalProperties = properties["numerical"];
    numericalProperties.forEach(function (numericalProperty) {
      lp.push(
        {
          label: numericalProperty + " equals",
          value: lp.length + 1,
          currentState: isEventType ? STATE_EVENT_PROPERTY_NAME: STATE_USER_PROPERTY_NAME,
          nextState: isEventType ? STATE_EVENT_NUMERIC_PROPERTY_VALUE: STATE_USER_NUMERIC_PROPERTY_VALUE,
          type: isEventType ? EVENT_PROPERTY_NAME_TYPE: USER_PROPERTY_NAME_TYPE,
          property: numericalProperty,
          operator: OPERATOR_EQUALS
        });
      lp.push(
        {
          label: numericalProperty + " greater than",
          value: lp.length + 1,
          currentState: isEventType ? STATE_EVENT_PROPERTY_NAME: STATE_USER_PROPERTY_NAME,
          nextState: isEventType ? STATE_EVENT_NUMERIC_PROPERTY_VALUE: STATE_USER_NUMERIC_PROPERTY_VALUE,
          type: isEventType ? EVENT_PROPERTY_NAME_TYPE: USER_PROPERTY_NAME_TYPE,
          property: numericalProperty,
          operator: OPERATOR_GREATER_THAN
        });
      lp.push(
        {
          label: numericalProperty + " lesser than",
          value: lp.length + 1,
          currentState: isEventType ? STATE_EVENT_PROPERTY_NAME: STATE_USER_PROPERTY_NAME,
          nextState: isEventType ? STATE_EVENT_NUMERIC_PROPERTY_VALUE: STATE_USER_NUMERIC_PROPERTY_VALUE,
          type: isEventType ? EVENT_PROPERTY_NAME_TYPE: USER_PROPERTY_NAME_TYPE,
          property: numericalProperty,
          operator: OPERATOR_LESSER_THAN
        });
    });
    return lp;
  }

  getPropertyValueOptions(propertyValues, isEventType) {
    var lp = [];
    propertyValues.forEach(function (propertyValue) {
      lp.push(
        { label: propertyValue,
          value: lp.length + 1,
          currentState: isEventType ? STATE_EVENT_STRING_PROPERTY_VALUE : STATE_USER_STRING_PROPERTY_VALUE,
          nextState: STATE_PROPERTY_TYPE,
          'type': STRING_VALUE_TYPE },
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
          { label: '', isDisabled: false, type: SUBMIT_QUERY_TYPE },
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
        'labels': [],
        [DYNAMIC_FETCH_USER_PROPERTIES]: true,
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
        'labels': [],
        [ALLOW_STRING_CREATE]: true,
        [DYNAMIC_FETCH_USER_PROPERTY_VALUES]: true,
        'nextState': STATE_PROPERTY_TYPE,
      },
    };
    return queryStates;
  }

  toggle = () => {
    this.setState({ collapse: !this.state.collapse });
  }

  toggleModelSelector = () => {
    this.setState(prevState => ({
      isModelSelectorDropdownOpen: !prevState.isModelSelectorDropdownOpen
    }));
  }

  handleModelSelectorChange = (selectedOption) => {
    var clickedModelId = selectedOption.value;
    var selectedInterval;
    for (var i = 0; i < this.props.intervals.length; i++) {
      var interval = this.props.intervals[i];
      if (interval.mid ==  clickedModelId){
        selectedInterval = interval;
        break;
      }
    }
    if(!!selectedInterval.mid){
      this.setState({selectedModelInterval: selectedInterval});
    }
    
    this.props.resetFactors();
  }

  toggleFade = () => {
    this.setState((prevState) => { return { fadeIn: !prevState } });
  }

  factor = (queryElements) => {

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
          newEvent["user_properties"] = [];
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
          query.eventsWithProperties[numEvents - 1].user_properties.push(newUserProperty);
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
          var numProperties = currentEvent.user_properties.length;
          var currentUserProperty = currentEvent.user_properties[numProperties - 1];
          currentUserProperty['property'] = queryElement.property;
          currentUserProperty['operator'] = queryElement.operator;
          nextExpectedTypes = [NUMERICAL_VALUE_TYPE, STRING_VALUE_TYPE];
          break;
        case NUMERICAL_VALUE_TYPE:
          var numEvents = query.eventsWithProperties.length;
          var currentEvent = query.eventsWithProperties[numEvents - 1];
          var numProperties = currentEvent.user_properties.length;
          var currentUserProperty = currentEvent.user_properties[numProperties - 1];
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
          var numProperties = currentEvent.user_properties.length;
          var currentUserProperty = currentEvent.user_properties[numProperties - 1];
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

    this.setState({ factors: { loading: LOADING_INIT } });

    let eventProperties = {
      projectId: this.props.currentProjectId,
      modelId: this.state.selectedModelInterval.mid,
      interval: this.getReadableInterval(this.state.selectedModelInterval),
      query: JSON.stringify(query),
    };
    let startTime = new Date().getTime();

    this.props.fetchFactors(this.props.currentProjectId,
      this.state.selectedModelInterval.mid, { query: query }, this.props.location.search)
        .then((response) => {
          console.log('Factors completed');
          this.setState({ factors: { loading: LOADING_DONE } });

          let endTime = new Date().getTime();
          eventProperties['time_taken_in_ms'] = endTime - startTime;
          eventProperties['results_count'] = response.data.charts.length;
          eventProperties['request_failed'] = (!response.ok).toString();
          if (!response.ok) eventProperties['error'] = JSON.stringify(response.data);
          factorsai.track('factor', eventProperties);
        })
        .catch((err) => {
          console.error(err);

          let endTime = new Date().getTime();
          eventProperties['time_taken_in_ms'] = endTime - startTime;
          eventProperties['error'] = err.message;
          eventProperties['request_failed'] = 'true';
          factorsai.track('factor', eventProperties);
        })

  }

  readableTimstamp(unixTime) {
    return moment.unix(unixTime).utc().format('MMM DD, YYYY');
  }

  getReadableInterval = (interval) => {
    let prefix = ''
    if (interval.mt === 'w') { prefix = '[w]'; }
    else if (interval.mt === 'm') { prefix = '[m]'; }
    else { throw new Error('invalid model type'); }

    return { 
      label: prefix + ' ' + this.readableTimstamp(interval.st)+' - '+this.readableTimstamp(interval.et), 
      value: interval.mid
    };
  }

  getIntervalOptions(intervals){
    if(intervals.length==0){
      return [{ 
        label:"", 
        value: ""
      }]
    }
    return intervals.map(this.getReadableInterval);
  }

  getIntervalDisplayValue() {
    if(this.state.selectedModelInterval != null){
      return this.getReadableInterval(this.state.selectedModelInterval);
    }
    return null 
  }

  getChartContainerContent(charts=[]) {
    if (this.state.factors.loading == LOADING_DEFAULT || 
      this.state.factors.loading == LOADING_INIT) return null;
    
    if (charts.length == 0) {
      return (
        <Col md={{ size: 'auto', offset: 5 }} style={{paddingTop:'8%', color: '#c0c0c0'}}>
          <h2> No results </h2>
        </Col>
      )
    }

    return charts;
  }

  isLoaded() {
    if(!this.props.currentProjectId){
      return true;
    }
    return this.state.eventNames.loaded && 
      this.state.models.loaded;
  }

  render() {
    if (!this.isLoaded()) return <Loading />;
    // Render empty view
    if(!this.props.currentProjectId){
      return (
        <div></div>
      )
    }

    var charts = [];
    if (!!this.props.factors.charts) {
      for (var i = 0; i < this.props.factors.charts.length; i++) {
        // note: we add a key prop here to allow react to uniquely identify each
        // element in this array. see: https://reactjs.org/docs/lists-and-keys.html
        var chartData = this.props.factors.charts[i];
        if (chartData.type === 'line') {
          charts.push(<Row className="animated fadeIn" style={chartCardRowStyle} key={i}><Col sm={cardColumnSetting}><LineChartCard chartData={chartData} /></Col></Row>)
        } else if (chartData.type === 'bar') {
          charts.push(<Row className="animated fadeIn" style={chartCardRowStyle} key={i}><Col sm={cardColumnSetting}><BarChartCard chartData={chartData} key={i} /></Col></Row>)
        } else if (chartData.type === 'funnel') {
          charts.push(<Row className="animated fadeIn" style={chartCardRowStyle} key={i}><Col sm={cardColumnSetting}><FunnelChartCard chartData={chartData} /></Col></Row>);
        }
      }
    }

    return (
      <div class="fapp-content" style={{margin: 0}}>
        <div>
          <Row class="fapp-select">
            <Col xs={{size: 10, offset: 1}} md={{ size: 3, offset: 8 }}>
              <Select
                value={this.getIntervalDisplayValue()}
                onChange={this.handleModelSelectorChange}
                options={this.getIntervalOptions(this.props.intervals)}
                placeholder="No intervals"
              />
            </Col>
          </Row>  
        </div>
        <div className='animated fadeIn'>
            <div>
              <Row>
                <Col xs='12' md='12'>
                  <QueryBuilderCard 
                    getQueryStates={this.getQueryStates}
                    getPropertiesOptions={this.getPropertiesOptions}
                    getPropertyValueOptions={this.getPropertyValueOptions}
                    onKeyDown={this.factor}
                    holderText="Enter goal."
                    resetCharts={this.props.resetFactors}
                  />
                </Col>
              </Row>
            </div>
            <Col md={{ size: 'auto', offset: 5 }} style={{padding:'8% 43px', display: this.state.factors.loading === LOADING_INIT ? 'block' : 'none'}}>
              <img src={loadingImage} alt="Loading.." /> 
            </Col>
            <Card className="fapp-card-border-none">
              { this.getChartContainerContent(charts) }
            </Card>
        </div>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Factor);