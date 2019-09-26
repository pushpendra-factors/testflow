import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Button,
  Card,
  CardBody,
  Col,
  FormGroup,
  Row,
} from 'reactstrap';
import CreatableSelect from 'react-select/lib/Creatable';
import Select from 'react-select';
import { 
  fetchProjectEventProperties, 
  fetchProjectEventPropertyValues,
  fetchProjectUserProperties, 
  fetchProjectUserPropertyValues
} from '../../actions/projectsActions';
import { deepEqual, QUERY_TYPE_FACTOR } from '../../util';

const queryBuilderStyles = {
  multiValue: () => ({
    background: 'none',
    fontSize: '1.3em',
  }),
  multiValueRemove: () => ({
    display: 'none',
  }),
  option: (base, state) => {
    let _style = {
      ...base
    }

    if(state.data.type == SUBMIT_QUERY_TYPE) {
      _style.backgroundColor = "#FFF"
      _style.fontSize = "0px"
      _style.padding = "0"
    }

    return _style;
  },
}

export const ALLOW_NUMBER_CREATE = "allowNumberCreate";
export const ALLOW_STRING_CREATE = "allowStringCreate";
export const DYNAMIC_FETCH_EVENT_PROPERTIES = "dynamicFetchEventProperties";
export const DYNAMIC_FETCH_EVENT_PROPERTY_VALUES = "dynamicFetchEventPropertyValues";
export const DYNAMIC_FETCH_USER_PROPERTIES = "dynamicFetchUserProperties";
export const DYNAMIC_FETCH_USER_PROPERTY_VALUES = "dynamicFetchUserPropertyValues";
export const NUMERICAL_VALUE_TYPE = "numericalValue";
export const STRING_VALUE_TYPE = "stringValue"
export const SUBMIT_QUERY_TYPE = "submitQuery"

export const STATE_EVENTS = 0;
export const STATE_PROPERTY_TYPE = 1;
export const STATE_EVENT_PROPERTY_NAME = 2;
export const STATE_USER_PROPERTY_NAME = 3;
export const STATE_EVENT_NUMERIC_PROPERTY_VALUE = 4;
export const STATE_EVENT_STRING_PROPERTY_VALUE = 5;
export const STATE_USER_NUMERIC_PROPERTY_VALUE = 6;
export const STATE_USER_STRING_PROPERTY_VALUE = 7;

const TYPE_EVENT_OCCURRENCE = 'events_occurrence';
const TYPE_UNIQUE_USERS = 'unique_users';
const ANALYSIS_TYPE_OPTS = [
  { value: TYPE_EVENT_OCCURRENCE, label: 'Number of occurrences of' },
  { value: TYPE_UNIQUE_USERS, label: 'Number of users with action' }
];

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    currentProjectEventNames: store.projects.currentProjectEventNames,
    eventPropertiesMap: store.projects.eventPropertiesMap,
    eventPropertyValuesMap: store.projects.eventPropertyValuesMap,
    userProperties: store.projects.userProperties,
    userPropertyValuesMap: store.projects.userPropertyValuesMap,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchProjectEventProperties, 
    fetchProjectEventPropertyValues,
    fetchProjectUserProperties, 
    fetchProjectUserPropertyValues
  }, dispatch);
}

class QueryBuilderCard extends Component {
  // Instance variables.
  latestSelectedEventName = null;
  latestSelectedEventProperty = null;
  latestSelectedUserProperty = null;

  constructor(props) {
    super(props);

    var queryStates;
    queryStates = this.props.getQueryStates(this.props.currentProjectEventNames)
    this.state = {
      type: ANALYSIS_TYPE_OPTS[1], // TYPE_UNIQUE_USERS as default.
      menuIsOpen: false,
      queryStates: queryStates,
      currentQueryState: STATE_EVENTS,
      currentOptions: queryStates[STATE_EVENTS]['labels'],
      allowNewOption: this.allowNewOptionMethod(queryStates[STATE_EVENTS]),
      noOptionsMessage: queryStates[STATE_EVENTS][ALLOW_NUMBER_CREATE] ? this.enterNumberMessage : this.noOptionsMessage,
      values: null,
      isLoadingOptions: false,
    };
    this.creatableSelect = React.createRef();
    this.latestSelectedEventName = null;
    this.latestSelectedEventProperty = null;
    this.latestSelectedUserProperty = null;
  }

  handleEventNamesChange(projectEventNames) {
    var queryStates;
    queryStates = this.props.getQueryStates(projectEventNames)
    this.setState({
      menuIsOpen: false,
      queryStates: queryStates,
      currentQueryState: STATE_EVENTS,
      currentOptions: queryStates[STATE_EVENTS]['labels'],
      allowNewOption: this.allowNewOptionMethod(queryStates[STATE_EVENTS]),
      noOptionsMessage: queryStates[STATE_EVENTS][ALLOW_NUMBER_CREATE] ? this.enterNumberMessage : this.noOptionsMessage,
      values: null,
      isLoadingOptions: false,
    });
    this.latestSelectedEventName = null;
    this.latestSelectedEventProperty = null;
    this.latestSelectedUserProperty = null;
  }

  allowNewOptionMethod(queryState) {
    if (queryState[ALLOW_STRING_CREATE]) {
      return this.allowAllNewOption;
    } else if (queryState[ALLOW_NUMBER_CREATE]) {
      return this.allowNumberOption
    } else {
      return this.disallowNewOption
    }
  }

  componentDidMount() {
    if (this.props.initEventName && this.props.initEventName != "") {
      let eventName = this.props.initEventName;
      let values = [{ label: eventName, value: eventName, currentState: STATE_EVENTS, nextState: STATE_PROPERTY_TYPE, type: "eventName" }];
      let actionMeta = { 
        action: "select-option", 
        option: { label: eventName, value: eventName, currentState: STATE_EVENTS, nextState: STATE_PROPERTY_TYPE, type: "eventName" } 
      }

      // initialize query with event name using
      // existing handle change method.
      this.handleChange(values, actionMeta);
      // trigger factor.
      this.props.onKeyDown(values, this.state.type.value);
    }
  }

  shouldComponentUpdate(nextProps, nextState) {
    // Deep equal is not required, if we don't copy eventNames to components state.
    if (!deepEqual(this.props.currentProjectEventNames, nextProps.currentProjectEventNames)) {
      this.handleEventNamesChange(nextProps.currentProjectEventNames);
    }
    if (this.state.queryStates[this.state.currentQueryState][DYNAMIC_FETCH_EVENT_PROPERTIES] &&
        this.state.currentOptions.length == 0 && 
        this.state.isLoadingOptions &&
        !!nextProps.eventPropertiesMap[this.latestSelectedEventName]) {
          var eventProperties = nextProps.eventPropertiesMap[this.latestSelectedEventName];
          var ops = this.props.getPropertiesOptions(eventProperties, true)
          var nextOptions = this.buildNewOptions(ops, this.state.values, this.state.values.length)
          this.setState({
            currentOptions: nextOptions,
            isLoadingOptions: false,
          })
    }
    if (this.state.queryStates[this.state.currentQueryState][DYNAMIC_FETCH_EVENT_PROPERTY_VALUES] &&
      this.state.currentOptions.length == 0 && 
      this.state.isLoadingOptions &&
      !!nextProps.eventPropertyValuesMap[this.latestSelectedEventName] &&
      !!nextProps.eventPropertyValuesMap[this.latestSelectedEventName][this.latestSelectedEventProperty]) {
        var eventPropertyValues = nextProps.eventPropertyValuesMap[this.latestSelectedEventName][this.latestSelectedEventProperty];
        var ops = this.props.getPropertyValueOptions(eventPropertyValues, true)
        var nextOptions = this.buildNewOptions(ops, this.state.values, this.state.values.length)
        this.setState({
          currentOptions: nextOptions,
          isLoadingOptions: false,
        })
    }
    if (this.state.queryStates[this.state.currentQueryState][DYNAMIC_FETCH_USER_PROPERTIES] &&
      this.state.currentOptions.length == 0 && 
      this.state.isLoadingOptions &&
      !!nextProps.userProperties) {
        var userProperties = nextProps.userProperties;
        var ops = this.props.getPropertiesOptions(userProperties, false)
        var nextOptions = this.buildNewOptions(ops, this.state.values, this.state.values.length)
        this.setState({
          currentOptions: nextOptions,
          isLoadingOptions: false,
        })
    }
    if (this.state.queryStates[this.state.currentQueryState][DYNAMIC_FETCH_USER_PROPERTY_VALUES] &&
      this.state.currentOptions.length == 0 && 
      this.state.isLoadingOptions &&
      !!nextProps.userPropertyValuesMap[this.latestSelectedUserProperty]) {
        var userPropertyValues = nextProps.userPropertyValuesMap[this.latestSelectedUserProperty];
        var ops = this.props.getPropertyValueOptions(userPropertyValues, false)
        var nextOptions = this.buildNewOptions(ops, this.state.values, this.state.values.length)
        this.setState({
          currentOptions: nextOptions,
          isLoadingOptions: false,
        })
    }
    return true
  }

  buildNewOptions(labels, selectedValues, valueMultiplier) {
    // Create new options with updated values to unique number so that
    // repeatitions of the same options can be allowed.
    // Not expecting more than 1000 options.
    var nextOptions = [];
    labels.forEach(function (element) {
      var shouldSkip = false;
      if (element.onlyOnce) {
        // Do not add options that can occur only once and has already occurred.
        selectedValues.forEach(function (entry) {
          if (entry.currentState === element.currentState &&
            entry.label === element.label) {
            shouldSkip = true;
          }
        });
      }
      if (!shouldSkip) {
        let newElement = Object.assign({}, element);
        newElement['value'] = (1000 * valueMultiplier) + element['value'];
        nextOptions.push(newElement);
      }
    });
    return nextOptions
  }

  handleChange = (newValues, actionMeta) => {
    var nextState = 0;
    var numEnteredValues = newValues.length

    // reset charts on query change.
    this.props.resetCharts();

    if (!!newValues && numEnteredValues > 0) {
      var currentEnteredOption = newValues[numEnteredValues - 1];
      
      if(currentEnteredOption['type'] === SUBMIT_QUERY_TYPE){
        this.creatableSelect.current.select.select.blur(); // Hide options menu.
        this.props.onKeyDown(this.state.values, this.state.type.value); // Submit factor.
        return
      }

      if (!!currentEnteredOption['__isNew__'] && actionMeta.action === 'create-option') {
        if (this.state.queryStates[this.state.currentQueryState][ALLOW_STRING_CREATE]) {
          currentEnteredOption['value'] = (1000 * (numEnteredValues - 1));
          currentEnteredOption['type'] = STRING_VALUE_TYPE;
          currentEnteredOption['currentState'] = this.state.queryStates[this.state.currentQueryState];
          currentEnteredOption['nextState'] = this.state.queryStates[this.state.currentQueryState]['nextState'];
        } else if (this.state.queryStates[this.state.currentQueryState][ALLOW_NUMBER_CREATE]) {
          currentEnteredOption['value'] = (1000 * (numEnteredValues - 1));
          currentEnteredOption['type'] = NUMERICAL_VALUE_TYPE;
          currentEnteredOption['currentState'] = this.state.queryStates[this.state.currentQueryState];
          currentEnteredOption['nextState'] = this.state.queryStates[this.state.currentQueryState]['nextState'];
        } else {
          // Unexpected.
          return
        }
      }
      nextState = currentEnteredOption['nextState'];
    }
    console.group('Value Changed');
    console.log(newValues);
    console.log(`action: ${actionMeta.action}`);
    console.groupEnd();

    if (this.state.currentQueryState == STATE_EVENTS) {
      // Update this.latestSelectedEventName if selected.
      if (numEnteredValues > 0) {
        this.latestSelectedEventName = newValues[numEnteredValues - 1]['label'];
      }
    }
    if (this.state.currentQueryState == STATE_EVENT_PROPERTY_NAME) {
      // Update  if selected.
      this.latestSelectedEventProperty = newValues[numEnteredValues - 1]['property'];
    }
    if (this.state.currentQueryState == STATE_USER_PROPERTY_NAME) {
      // Update  if selected.
      this.latestSelectedUserProperty = newValues[numEnteredValues - 1]['property'];
    }
    if (this.state.queryStates[nextState][DYNAMIC_FETCH_EVENT_PROPERTIES]) {
      this.props.fetchProjectEventProperties(this.props.currentProjectId,
        this.latestSelectedEventName, this.props.selectedModelId);
      this.setState({
        currentOptions: [],
        currentQueryState: nextState,
        allowNewOption: this.allowNewOptionMethod(this.state.queryStates[nextState]),
        noOptionsMessage: this.state.queryStates[nextState][ALLOW_NUMBER_CREATE] ? this.enterNumberMessage : this.noOptionsMessage,
        values: newValues,
        isLoadingOptions: true,
      });
    } else if (this.state.queryStates[nextState][DYNAMIC_FETCH_EVENT_PROPERTY_VALUES]) {
      console.log("Fetch property: " + this.latestSelectedEventProperty);
      this.props.fetchProjectEventPropertyValues(this.props.currentProjectId,
        this.latestSelectedEventName, this.latestSelectedEventProperty);
      this.setState({
        currentOptions: [],
        currentQueryState: nextState,
        allowNewOption: this.allowNewOptionMethod(this.state.queryStates[nextState]),
        noOptionsMessage: this.state.queryStates[nextState][ALLOW_NUMBER_CREATE] ? this.enterNumberMessage : this.noOptionsMessage,
        values: newValues,
        isLoadingOptions: true,
      });
    } else if (this.state.queryStates[nextState][DYNAMIC_FETCH_USER_PROPERTIES]) {
      this.props.fetchProjectUserProperties(this.props.currentProjectId, QUERY_TYPE_FACTOR, this.props.selectedModelId);
      this.setState({
        currentOptions: [],
        currentQueryState: nextState,
        allowNewOption: this.allowNewOptionMethod(this.state.queryStates[nextState]),
        noOptionsMessage: this.state.queryStates[nextState][ALLOW_NUMBER_CREATE] ? this.enterNumberMessage : this.noOptionsMessage,
        values: newValues,
        isLoadingOptions: true,
      });
    } else if (this.state.queryStates[nextState][DYNAMIC_FETCH_USER_PROPERTY_VALUES]) {
      console.log("Fetch property: " + this.latestSelectedEventProperty);
      this.props.fetchProjectUserPropertyValues(this.props.currentProjectId,
        this.latestSelectedUserProperty);
      this.setState({
        currentOptions: [],
        currentQueryState: nextState,
        allowNewOption: this.allowNewOptionMethod(this.state.queryStates[nextState]),
        noOptionsMessage: this.state.queryStates[nextState][ALLOW_NUMBER_CREATE] ? this.enterNumberMessage : this.noOptionsMessage,
        values: newValues,
        isLoadingOptions: true,
      });
    } else {
      var nextOptions = this.buildNewOptions(
        this.state.queryStates[nextState]['labels'], newValues, numEnteredValues)
      this.setState({
        currentOptions: nextOptions,
        currentQueryState: nextState,
        allowNewOption: this.allowNewOptionMethod(this.state.queryStates[nextState]),
        noOptionsMessage: this.state.queryStates[nextState][ALLOW_NUMBER_CREATE] ? this.enterNumberMessage : this.noOptionsMessage,
        values: newValues,
      });
    }
  };

  handleKeyDown = (event) => {
    console.log(event)
    switch (event.key) {
      case 'Enter':
        if (!this.state.menuIsOpen) {
          console.log(event);
          console.log(this.state.values);
          event.preventDefault();
          this.props.onKeyDown(this.state.values, this.state.type.value);
        }
    }
  };

  handleTypeChange = (option) => {
    this.setState({type: option});
  };

  disallowNewOption = (inputOption, valueType, optionsType) => {
    return false;
  };

  allowNumberOption = (inputOption, valueType, optionsType) => {
    return !!inputOption && !isNaN(inputOption);
  };

  allowAllNewOption = (inputOption, valueType, optionsType) => {
    return true;
  };

  noOptionsMessage = (inputValue) => {
    return 'No Options';
  };

  enterNumberMessage = (inputValue) => {
    return 'Enter a valid number';
  };

  formatCreateLabel = (inputValue) => {
    return inputValue;
  };

  render() {
    return (
      <Card className="fapp-search-card fapp-select light">
        <CardBody>
          <Row>
            <Col md={{ size: '3', offset: 2 }}>
              <div style={{display: 'inline-block', width: '280px', marginRight: '10px', marginBottom: "10px"}} className='fapp-select light'>
              <Select
                value={this.state.type}
                onChange={this.handleTypeChange}
                options={ANALYSIS_TYPE_OPTS}
                placeholder='Type'
              />
              </div>
            </Col>
          </Row>
          <Row>
            <Col md={{ size: '8', offset: 2 }}>
              <FormGroup>
                <div>
                  <CreatableSelect
                    onChange={this.handleChange}
                    onKeyDown={this.handleKeyDown}
                    isValidNewOption={this.state.allowNewOption}
                    noOptionsMessage={this.state.noOptionsMessage}
                    formatCreateLabel={this.formatCreateLabel}
                    isMulti={true}
                    onMenuOpen={() => this.setState({ menuIsOpen: true })}
                    onMenuClose={() => this.setState({ menuIsOpen: false })}
                    closeMenuOnSelect={false}
                    styles={queryBuilderStyles}
                    options={this.state.currentOptions}
                    value={this.state.values}
                    isLoading={this.state.isLoadingOptions}
                    placeholder={this.props.holderText}
                    ref={this.creatableSelect}
                  />
                </div>
              </FormGroup>
            </Col>
          </Row>
          <Row>
            <Col md={{ size: 'auto', offset: 5 }}>
              <Button block color='primary' onClick={() => { this.props.onKeyDown(this.state.values, this.state.type.value) }}>Factor</Button>
            </Col>
            <Col md={{ size: 'auto' }} style={{ display: 'none' }}>
              <Button block color='primary'>Paths!</Button>
            </Col>
          </Row>
        </CardBody>
      </Card>
    )
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(QueryBuilderCard);
