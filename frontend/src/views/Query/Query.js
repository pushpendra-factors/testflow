import React, { Component } from 'react';
import Select from 'react-select';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Row, Col, Button, ButtonGroup, ButtonToolbar, 
  ButtonDropdown, DropdownToggle, DropdownMenu, 
  Modal, ModalHeader, ModalBody, Form, DropdownItem,
  ModalFooter, Input } from 'reactstrap'; 
import 'react-date-range/dist/styles.css';
import 'react-date-range/dist/theme/default.css';
import moment from 'moment';
import mt from "moment-timezone"
import queryString from 'query-string';

import TableChart from './TableChart'
import LineChart from './LineChart';
import BarChart from './BarChart';
import TableBarChart from './TableBarChart';
import FunnelChart from './FunnelChart';
import { makeSelectOpts } from '../../util';

// Channel query is a different kind of component linked to Query.
import ChannelQuery from '../ChannelQuery/ChannelQuery';

import AttributionQuery from '../AttributionQuery/AttributionQuery'


import { PRESENTATION_BAR, PRESENTATION_LINE, PRESENTATION_TABLE, 
  PRESENTATION_CARD, PRESENTATION_FUNNEL, PROPERTY_TYPE_EVENT,
  getDateRangeFromStoredDateRange, PROPERTY_LOGICAL_OP_OPTS,
  DEFAULT_DATE_RANGE, DEFINED_DATE_RANGES, getGroupByTimestampType, 
  getQueryPeriod, convertFunnelResultForTable
} from './common';
import ClosableDateRangePicker from '../../common/ClosableDatePicker';
import { fetchProjectEvents, runQuery } from '../../actions/projectsActions';
import { fetchDashboards, createDashboardUnit } from '../../actions/dashboardActions';
import Event from './Event';
import GroupBy from './GroupBy';
import { 
  removeElementByIndex, getSelectedOpt, isNumber, createSelectOpts, 
  isSingleCountResult, slideUnixTimeWindowToCurrentTime,
  getLabelByValueFromOpts, getTimezoneString,
} from '../../util'
import Loading from '../../loading';
import factorsai from '../../common/factorsaiObj';
import { PROPERTY_TYPE_OPTS, USER_PREF_PROPERTY_TYPE_OPTS, 
  PROPERTY_VALUE_TYPE_DATE_TIME } from './common';
import insightsSVG from '../../assets/img/analytics/insights.svg';
import funnelSVG from '../../assets/img/analytics/funnel.svg';
import channelSVG from '../../assets/img/analytics/channel.svg';
import attributionSVG from '../../assets/img/analytics/attribution.svg'
import { del } from '../../actions/request';

const COND_ALL_GIVEN_EVENT = 'all_given_event';
const COND_ANY_GIVEN_EVENT = 'any_given_event'; 
const EVENTS_COND_OPTS = [
  { value: COND_ANY_GIVEN_EVENT, label: 'any' },
  { value: COND_ALL_GIVEN_EVENT, label: 'all' }
];
const LABEL_STYLE = { marginRight: '10px', fontWeight: '600', color: '#777' };

const QUERY_CLASS_INSIGHTS = 'insights';
const QUERY_CLASS_FUNNEL = 'funnel';
const QUERY_CLASS_CHANNEL = 'channel'
const QUERY_CLASS_ATTRIBUTION = 'attribution'
const QUERY_CLASS_OPTS = [
  { value: QUERY_CLASS_INSIGHTS, label: 'Insights' },
  { value: QUERY_CLASS_FUNNEL, label: 'Funnel' },
  { value: QUERY_CLASS_CHANNEL, label: 'Channel' },
  { value: QUERY_CLASS_ATTRIBUTION, label: 'attribution'}
];

const TYPE_EVENT_OCCURRENCE = 'events_occurrence';
const TYPE_UNIQUE_USERS = 'unique_users';

const QUERY_TYPE_OPTS = [
  { value: TYPE_EVENT_OCCURRENCE, label: 'events occurrence' },
  { value: TYPE_UNIQUE_USERS, label: 'unique users' },
];
const INSIGHTS_QUERY_TYPE_OPTS = QUERY_TYPE_OPTS;
const FUNNEL_QUERY_TYPE_OPTS = [
  { value: TYPE_UNIQUE_USERS, label: 'unique users' },
];

const AGGR_COUNT_OPT = {label: 'count', value: 'count'};
const AGGR_OPTS = [
  AGGR_COUNT_OPT, 
]

const ERROR_NO_EVENT = 'No events given. Please add atleast one event by clicking +Event button.';
const ERROR_FUNNEL_EXCEEDED_EVENTS = 'Funnel queries supports upto 4 events. Please ensure that you have the same.';

const DEFAULT_INSIGHTS_QUERY_PRESENTATION = PRESENTATION_TABLE;
const DEFAULT_FUNNEL_QUERY_PRESENTATION = PRESENTATION_FUNNEL;

const HEADER_COUNT = "count";
const HEADER_DATE = "date";

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    currentAgent: store.agents.agent,
    projects: store.projects.projects,
    viewQuery: store.projects.viewQuery,
    eventNames: store.projects.currentProjectEventNames,
    dashboards: store.dashboards.dashboards,

    eventPropertiesMap: store.projects.queryEventPropertiesMap,
    eventPropertyValuesMap: store.projects.queryEventPropertyValuesMap
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchProjectEvents,
    fetchDashboards,
    createDashboardUnit,
  }, dispatch)
}

class Query extends Component {
  constructor(props) {
    super(props);

    this.state = {
      eventNamesLoaded: false,
      eventNamesLoadError: null,

      // add to resetQueryInterface to reset on
      // interface change.
      class: QUERY_CLASS_OPTS[0],
      aggr: AGGR_OPTS[0],
      condition: EVENTS_COND_OPTS[0],
      type: INSIGHTS_QUERY_TYPE_OPTS[0], // 1st type as default.
      events: [],
      groupBys: [],
      resultDateRange: [DEFAULT_DATE_RANGE],

      result: null,
      resultError: null,
      isResultLoading: false,
      selectedPresentation: DEFAULT_INSIGHTS_QUERY_PRESENTATION,

      showPresentation: false,
      showDatePicker: false,
      topError: null,

      showDashboardsList: false,
      showAddToDashboardModal: false,
      addToDashboardMessage: null,
      inputDashboardUnitTitle: null,
      selectedDashboardId: null,

      timeZone:null,
    }
  }

  getQueryTypeOptsByClass = () => {
    return this.state.class.value == QUERY_CLASS_FUNNEL ? FUNNEL_QUERY_TYPE_OPTS : INSIGHTS_QUERY_TYPE_OPTS;
  }

  getDefaultPresentationByClass() {
    return this.state.class.value == QUERY_CLASS_FUNNEL ? 
      DEFAULT_FUNNEL_QUERY_PRESENTATION : DEFAULT_INSIGHTS_QUERY_PRESENTATION;
  }
  
  resetQueryInterfaceOnClassChange() {
    if (this.isViewQuery()) {
      // reset presentation alone.
      this.setState({
        result: null,
        selectedPresentation: this.getDefaultPresentationByClass(),
        showPresentation: false
      });

      return;
    }

    this.setState({
      // reset query state.
      condition: EVENTS_COND_OPTS[0],
      type: this.getQueryTypeOptsByClass()[0],
      events: [],
      groupBys: [],
      resultDateRange: [DEFAULT_DATE_RANGE],
      // reset presentation.
      result: null,
      selectedPresentation: this.getDefaultPresentationByClass(),
      showPresentation: false,
    });

    this.initWithAnEventRow();
  }

  componentDidUpdate(prevProps, prevState) {
    if (prevState.class.value != this.state.class.value) {
      this.resetQueryInterfaceOnClassChange();
    }
  }

  isViewQuery() {
    let queryParams = queryString.parse(this.props.location.search);
    return queryParams && queryParams.view && queryParams.view != "" && 
          Object.keys(this.props.viewQuery).length > 0;
  }

  componentWillMount() {
    this.props.fetchProjectEvents(this.props.currentProjectId)
      .then(() => {
        this.setState({ eventNamesLoaded: true, timeZone: this.getCurrentTimeZone() });

        // init query builder.
        if (this.isViewQuery()) {
          this.initWithViewQuery();
        } else {
          this.initWithAnEventRow();
        }
      })
      .catch((r) => this.setState({ eventNamesLoaded: true, eventNamesLoadError: r.paylaod }));

    this.props.fetchDashboards(this.props.currentProjectId);
  }

  initWithViewQuery() {
    let storeQuery = this.props.viewQuery;

    let queryState = {};
    queryState.class = { 
      value: storeQuery.cl ? storeQuery.cl : QUERY_CLASS_INSIGHTS 
    }
    queryState.type = {
      value: storeQuery.ty,
      label: getLabelByValueFromOpts(QUERY_TYPE_OPTS, storeQuery.ty)
    }
    queryState.condition = {
      value: storeQuery.ec,
      label: getLabelByValueFromOpts(EVENTS_COND_OPTS, storeQuery.ec)
    }

    let events = [];
    for (let ei=0; ei<storeQuery.ewp.length; ei++) {
      let event = storeQuery.ewp[ei];

      let properties = [];
      for (let pi=0; pi<event.pr.length; pi++) {
        let prop = event.pr[pi];

        let vProp = {};
        vProp.logicalOp = prop.lop;
        vProp.entity = prop.en;
        vProp.name = prop.pr;
        vProp.op = prop.op;
        vProp.valueType = prop.ty;
        vProp.value = prop.va;

        properties.push(vProp);
      }

      events.push({ name: event.na, properties: properties });
    }
    queryState.events = events;

    let groupBys = [];
    for (let gi=0; gi<storeQuery.gbp.length; gi++) {
      let prop = storeQuery.gbp[gi];

      let group = {};
      group.type = prop.en;
      group.name = prop.pr;
      group.eventName = prop.ena;

      groupBys.push(group);
    }
    queryState.groupBys = groupBys;

    queryState.resultDateRange = getDateRangeFromStoredDateRange(storeQuery);

    console.log("Stored Query : ", storeQuery);
    console.log("View Query State : ", queryState);
    this.setState(queryState);
  }

  initWithAnEventRow() {
    this.addEvent();

    if (this.props.eventNames.length > 0) {
      this.onEventStateChange(getSelectedOpt(this.props.eventNames[0]), 0);
    } else {
      console.error('Query not initialized with an event row. zero events found.');
    }
  }

  handleClassChange = (option) => {
    this.setState({class: option});
  }

  handleTypeChange = (option) => {
    this.setState({type: option});
  }
  
  handleEventsConditionChange = (option) => {
    this.setState({condition: option});
  }

  getDefaultEventState() {
    return { name:'', properties:[] };
  }

  addEvent = () => {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.events = [ ...prevState.events ];
      // init with default state for each event row.
      state.events.push(this.getDefaultEventState());
      return state;
    });
  }

  onEventStateChange(option, index) {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.events = [ ...prevState.events ];
      state.events[index]={name:option.value, properties:[]};
      return state;
    })
  }

  getDefaultPropertyState() {
    let entities = Object.keys(PROPERTY_TYPE_OPTS);
    let logicalOps = Object.keys(PROPERTY_LOGICAL_OP_OPTS);
    return { entity: entities[0],  name: '', op: 'equals', value: '', valueType: '', logicalOp: logicalOps[0] };
  }

  addProperty(eventIndex) {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.events = [ ...prevState.events ];
      // init with default state for each propety row by event index.
      state.events[eventIndex].properties.push(this.getDefaultPropertyState())
      return state;
    })
  }

  setPropertyAttr = (eventIndex, propertyIndex, attr, value) => {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.events[eventIndex].properties = [...prevState.events[eventIndex].properties]
      state.events[eventIndex]['properties'][propertyIndex][attr] = value
      return state;
    })
  }

  onPropertyEntityChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'entity', value)
    this.setPropertyAttr(eventIndex,propertyIndex,'name',"")
    this.setPropertyAttr(eventIndex,propertyIndex,'value',"")
    this.setPropertyAttr(eventIndex, propertyIndex, 'valueType', "");
  }

  onPropertyLogicalOpChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'logicalOp', value)
  }

  onPropertyNameChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'name', value)
    this.setPropertyAttr(eventIndex,propertyIndex,'value',"")
  }

  onPropertyOpChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'op', value)
    this.setPropertyAttr(eventIndex,propertyIndex,'value',"")
  }

  onPropertyValueChange = (eventIndex, propertyIndex, value, type) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'value', value);
    this.setPropertyAttr(eventIndex, propertyIndex, 'valueType', type);
  }

  getDefaultGroupByState() {
    let groupByOpts = this.getGroupByOpts();

    let defaultEventName = '';
    if (this.state.events.length > 0) 
      defaultEventName = this.state.events[0].name;

    return { type: groupByOpts[0].value, name: '', eventName: defaultEventName };
  }

  addGroupBy = () => {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.groupBys = [ ...prevState.groupBys ];
      state.groupBys.push(this.getDefaultGroupByState());
      return state;
    })
  }

  setGroupByAttr(groupByIndex, attr, value) {
    this.setState((prevState) => {
      let state = { ...prevState };
      state.groupBys = [ ...prevState.groupBys ];
      state.groupBys[groupByIndex][attr] = value;
      return state;
    })
  }

  onGroupByTypeChange = (groupByIndex, option) => {
    this.setGroupByAttr(groupByIndex, 'type', option.value);
    this.setGroupByAttr(groupByIndex, 'name', "");
  }

  onGroupByNameChange = (groupByIndex, option) => {
    this.setGroupByAttr(groupByIndex, 'name', option.value);
  }

  onGroupByEventNameChange = (groupByIndex, option) => {
    this.setGroupByAttr(groupByIndex, 'eventName', option.value);

  }

  handleResultDateRangeSelect = (range) => {
    range.selected.label = null; // set null on custom range.
    this.setState({ resultDateRange: [range.selected] });
  }

  readableDateRange(range) {
    // Use label for default date range.
    if(range.startDate ==  DEFAULT_DATE_RANGE.startDate 
      && range.endDate == DEFAULT_DATE_RANGE.endDate)
      return DEFAULT_DATE_RANGE.label;

    return moment(range.startDate).format('MMM DD, YYYY') + " - " +
      moment(range.endDate).format('MMM DD, YYYY');
  }

  toggleDatePickerDisplay = () => {
    this.setState({ showDatePicker: !this.state.showDatePicker });
  }
  

  getQuery(presentation, toSave=false) {
    let query = {};
    query.cl = this.state.class.value;
    query.ty = this.state.type.value;
    query.ec = this.state.condition.value;
    // event_occurrence supports only any_given_event.
    if (query.ty == TYPE_EVENT_OCCURRENCE) {
      query.ec = COND_ANY_GIVEN_EVENT;
    }
    
    if (this.state.resultDateRange.length == 0)
      throw new Error('Invalid date range. No default range given.')
    
    let period = getQueryPeriod(this.state.resultDateRange[0], this.state.timeZone)
    query.fr=period.from
    query.to=period.to
    query.ewp = []
    for(let ei=0; ei < this.state.events.length; ei++) {
      let event = this.state.events[ei];
      if (event.name == "") continue;
      
      let ewp = {};
      ewp.na = event.name;
      ewp.pr = [];

      for(let pi=0; pi < event.properties.length; pi++) {
        let property = event.properties[pi];
        let cProperty = {}
        
        if (property.entity != '' && property.name != '' &&
            property.operator != '' && property.value != '' &&
            property.valueType != '') {

            // Todo: show validation error.
            if (property.valueType == 'numerical' && 
              !isNumber(property.value))
              continue;

            cProperty.en = property.entity;
            cProperty.pr = property.name;
            cProperty.op = property.op;
            cProperty.va = property.value;
            cProperty.ty = property.valueType;
            cProperty.lop = property.logicalOp;

            // update datetime with current time window if ovp is true.
            if (property.valueType == PROPERTY_VALUE_TYPE_DATE_TIME) {
              let dateRange = JSON.parse(cProperty.va);
              if (dateRange.ovp) {
                let newRange = slideUnixTimeWindowToCurrentTime(dateRange.fr, dateRange.to);
                dateRange.fr = newRange.from;
                dateRange.to = newRange.to;
                cProperty.va = JSON.stringify(dateRange); 
              }
            }

            ewp.pr.push(cProperty);
        }
      }
      query.ewp.push(ewp)
    }

    query.gbp = [];
    for(let i=0; i < this.state.groupBys.length; i++) {
      let groupBy = this.state.groupBys[i];
      let cGroupBy = {};

      if (groupBy.name != '' && groupBy.type != '') {
        cGroupBy.pr = groupBy.name;
        cGroupBy.en = groupBy.type;

        // add group by event name.
        if (groupBy.type == PROPERTY_TYPE_EVENT && this.isEventNameRequiredForGroupBy() &&  
          groupBy.eventName != '') cGroupBy.ena = groupBy.eventName;
          
        query.gbp.push(cGroupBy)
      }
    }

    query.gbt = (presentation == PRESENTATION_LINE) ? 
      getGroupByTimestampType(query.fr, query.to) : '';

    let timezone = this.state.timeZone;
    query.tz = (!toSave && timezone && timezone != '') ? timezone : '';
  
    return query
  }

  isResponseValid(result) {
    if (result.error) {
      this.setState({ resultError: result.error });
      return false;
    }

    return !!result.headers;
  }

  validateQuery() {
    let hasEvent = false;
    for(let i=0; i<this.state.events.length; i++) {
      if (this.state.events[i].name !== "") {
        hasEvent = true;
      }
    }
    if (!hasEvent) return ERROR_NO_EVENT;
    if (this.state.class.value == QUERY_CLASS_FUNNEL && this.state.events.length > 4)
      return ERROR_FUNNEL_EXCEEDED_EVENTS;

    return "";
  }

  showTopError(error) {
    this.setState({ topError: error });
  }

  resetTopError() {
    this.setState({ topError: null });
  }
  
  resetResult() {
    this.setState({ result: null });
  }

  run = (presentation) => {
    if (presentation == "")
      throw new Error('Invalid presentation');

    let err = this.validateQuery();
    if (err != "") {
      this.showTopError(err);
      this.resetResult();
      return;
    } else {
      this.resetTopError();
    }

    this.scrollToBottom();
    this.setState({ 
      isResultLoading: true, 
      showPresentation: true, 
      selectedPresentation: presentation,
    });
    let query = this.getQuery(presentation);

    let eventProperties = { 
      projectId: this.props.currentProjectId,
      query: JSON.stringify(query),
      class: query.cl,
      type: query.ty,
      condition: query.ec,
      presentation: presentation,
    };
    let startTime = new Date().getTime();
    
    runQuery(this.props.currentProjectId, query)
      .then((r) => {
        if(this.isResponseValid(r.data)) {
          this.setState({ 
            result: r.data, 
            isResultLoading: false,
          });
        } else {
          console.log('Failed to run query. Invalid response.');
        }

        let endTime = new Date().getTime();
        eventProperties['time_taken_in_ms'] = endTime - startTime;
        eventProperties['request_failed'] = (!r.ok).toString();
        if (!r.ok) eventProperties['error'] = JSON.stringify(r.data);
        factorsai.track('run_query', eventProperties);
      })
      .catch((err) => {
        console.log(err);

        let endTime = new Date().getTime();
        eventProperties['time_taken_in_ms'] = endTime - startTime;
        eventProperties['error'] = err.message;
        eventProperties['request_failed'] = 'true';
        factorsai.track('run_query', eventProperties);
      });
  }

  getResultAsTable() {
    if (!isSingleCountResult(this.state.result)) 
      return <TableChart queryResult={this.state.result} />;
    
    return (
      <div style={{ marginTop: '150px', marginBottom: '100px' }} >
        <TableChart card noHeader bordered queryResult={this.state.result} />
      </div>
    );
  }

  getResultAsLineChart() {
    return <div style={{height: '450px'}} className='animated fadeIn'>
      <LineChart queryResult={this.state.result} /> 
    </div>;
  }
  
  getResultAsVerticalBarChart() {
    return <div style={{height: '450px'}} className='animated fadeIn'> 
      <BarChart queryResult={this.state.result} legend={false} />
    </div>;
  }

  getResultAsTabularBarChart() {
    let result = this.state.result;
    return <div className='animated fadeIn'><TableBarChart data={result} /></div>;
  }

  renderInsightsPresentation = () => {
    if (this.state.isResultLoading) return <Loading paddingTop='14%' />;

    if (this.state.result == null) return null;
    let selected = this.state.selectedPresentation;
    
    if (selected == PRESENTATION_TABLE) {
      return this.getResultAsTable();
    }
    if (selected == PRESENTATION_LINE) {
      return this.getResultAsLineChart();
    }
    if (selected == PRESENTATION_BAR) {      
        return (this.state.result.headers.length <= 2) ? this.getResultAsVerticalBarChart() : this.getResultAsTabularBarChart();
    }
  }

  getPresentationSelectorClass(type) {
    return this.state.selectedPresentation == type ? 'btn btn-primary' : 'btn btn-outline-primary';
  }

  getEventNames = () => {
    return this.state.events.map((e) => { return e.name; })
  }

  remove = (arrayKey, index) => {
    this.setState((pState) => {
      let state = { ...pState };
      state[arrayKey] = removeElementByIndex(state[arrayKey], index);
      return state
    })
  }

  removeEventProperty = (eventIndex, propertyIndex) => {
    this.setState((pState) => {
      let state = { ...pState };
      state['events'][eventIndex]['properties'] = removeElementByIndex(state['events'][eventIndex]['properties'], propertyIndex);
      return state;
    })
  }

  toggleDashboardsList = () => {
    this.setState({ showDashboardsList: !this.state.showDashboardsList });
  }

  isLoaded() {
    return this.state.eventNamesLoaded;
  }

  getGroupByOpts = () => {
    // user property on top.
    if (this.state.type.value == TYPE_UNIQUE_USERS || 
      this.state.class.value == QUERY_CLASS_FUNNEL) {
      
      return createSelectOpts(USER_PREF_PROPERTY_TYPE_OPTS);
    }

    return createSelectOpts(PROPERTY_TYPE_OPTS);
  }

  isEventNameRequiredForGroupBy = () => {
    return (this.state.type.value == TYPE_UNIQUE_USERS && 
      this.state.condition.value == COND_ALL_GIVEN_EVENT) || 
      this.state.class.value == QUERY_CLASS_FUNNEL;
  }

  showAddToDashboardFailure() {
    this.setState({ addToDashboardMessage: 'Failed to add chart to dashboard' });
  }

  addToDashboard = () => {
    if (this.state.selectedPresentation == null) {
      console.error('Invalid presentation');
      return;
    }
    let presentation = this.state.selectedPresentation;

    if (presentation === PRESENTATION_TABLE 
      && isSingleCountResult(this.state.result)) {
      presentation = PRESENTATION_CARD;
    }
    
    let query = this.getQuery(this.state.selectedPresentation, true);
    let payload = {
      presentation: presentation,
      query: query,
      title: this.state.inputDashboardUnitTitle,
    };

    if (this.state.selectedDashboardId == null) {
      throw new Error('Invalid dashboard to add.');
    }

    this.props.createDashboardUnit(this.props.currentProjectId, this.state.selectedDashboardId, payload)
      .then((r) => { 
        if (!r.ok) this.showAddToDashboardFailure();
        else this.toggleAddToDashboardModal(); 
      })
      .catch(() => { this.showAddToDashboardFailure(); });
  }

  toggleAddToDashboardModal = () =>  {
    this.setState({ showAddToDashboardModal: !this.state.showAddToDashboardModal, addToDashboardMessage: null });
  }

  setDashboardUnitTitle = (e) => {
    this.setState({ addToDashboardMessage: null });

    let title = e.target.value.trim();
    if (title == "") console.error("chart title cannot be empty");
    this.setState({ inputDashboardUnitTitle: title });
  }

  selectDashboardToAdd = (event) => {
    let dashboardId = event.currentTarget.getAttribute('value');
    this.setState({ selectedDashboardId: dashboardId })
    this.toggleAddToDashboardModal();
  }

  disableAddToDashboard() {
    return (
      // funnel presentation for class funnel with breakdown.
      this.state.class.value == QUERY_CLASS_FUNNEL && 
      this.state.selectedPresentation == PRESENTATION_FUNNEL && 
      this.state.groupBys.length > 0
    ) || (
      // tablular bar chart.
      this.state.selectedPresentation === PRESENTATION_BAR && 
      this.state.groupBys.length > 1
    ); 
  }

  scrollToBottom = () => {
    if (this.endOfPresentation != undefined)
      this.endOfPresentation.scrollIntoView({ behavior: "smooth" });
  }

  renderEventsWithProperties() {
    let events = [];
    for(let i=0; i<this.state.events.length; i++) {
      events.push(
        <Event 
          index={i}
          key={'events_'+i} 
          projectId={this.props.currentProjectId} 
          nameOpts={this.props.eventNames} 
          eventState={this.state.events[i]}
          remove={() => this.remove('events', i)}
          removeProperty={(propertyIndex) => this.removeEventProperty(i, propertyIndex)}
          // event handlers.
          onNameChange={(value) => this.onEventStateChange(value, i)} 
          // property handlers.
          onAddProperty={() => this.addProperty(i)}
          onPropertyEntityChange={this.onPropertyEntityChange}
          onPropertyLogicalOpChange={this.onPropertyLogicalOpChange}
          onPropertyNameChange={this.onPropertyNameChange}
          onPropertyOpChange={this.onPropertyOpChange}
          onPropertyValueChange={this.onPropertyValueChange}
        />
      )
    }

    let addEventButton = <Row style={{marginBottom: '15px'}}>
      <Col xs='12' md='12'>
        <Button outline color='primary' onClick={this.addEvent} style={{ marginTop: '3px' }}>+ Event</Button>
      </Col>
    </Row>

    return [events, addEventButton];
  }

  closeDatePicker = () => {
    this.setState({ showDatePicker: false });
  }



  changeTimeZone= ({value})=>{
    let isValidTimezone = mt.tz.zone(value)
    if(isValidTimezone){
      this.setState({timeZone: value});
    }
  }

  getCurrentTimeZone(){
    let timeZone = mt.tz.guess();
    return timeZone;
  }

  renderDateRangeSelector() {
    return (
      <Row style={{marginBottom: '15px'}}>
        <Col xs='12' md='12'>
          <span style={LABEL_STYLE}> during </span>
          <Button outline style={{border: '1px solid #ccc', color: 'grey', marginRight: '10px' }} 
            onClick={this.toggleDatePickerDisplay}>
            <i className="fa fa-calendar" style={{marginRight: '10px'}}></i>
            { this.readableDateRange(this.state.resultDateRange[0]) } 
          </Button>
          <span style={LABEL_STYLE}> timezone </span>
          <div style={{display: 'inline-block', width: '185px', marginRight: '10px'}} className='fapp-select light'>
          <Select 
          placeholder={this.state.timeZone} 
          options={makeSelectOpts(moment.tz.names())}
          onChange={this.changeTimeZone}
          />
          </div>
          <div className='fapp-date-picker' hidden={!this.state.showDatePicker}>
            <ClosableDateRangePicker
              ranges={this.state.resultDateRange}
              onChange={this.handleResultDateRangeSelect}
              staticRanges={ DEFINED_DATE_RANGES }
              inputRanges={[]}
              minDate={new Date('01 Jan 2000 00:00:00 GMT')} // range starts from given date.
              maxDate={new Date()}
              closeDatePicker={this.closeDatePicker}
            />
            <button className='fapp-close-round-button' style={{float: 'right', marginLeft: '0px', borderLeft: 'none'}} 
            onClick={this.toggleDatePickerDisplay}>x</button>
          </div>
        </Col>
      </Row>
    );
  }

  renderGroupBys() {
    let groupBys = [];
    for(let i=0; i<this.state.groupBys.length; i++) {
      groupBys.push(
        <GroupBy 
          key={'groupby_'+i}
          remove={() => this.remove('groupBys', i)}
          projectId={this.props.currentProjectId}
          getSelectedEventNames={this.getEventNames}
          groupByState={this.state.groupBys[i]}
          onTypeChange={(option) => this.onGroupByTypeChange(i, option)} 
          onNameChange={(option) => this.onGroupByNameChange(i, option)}
          onEventNameChange={(option) => this.onGroupByEventNameChange(i, option)}
          getOpts={this.getGroupByOpts}
          isEventNameRequired={this.isEventNameRequiredForGroupBy}
        />
      );
    }

    let addGroupByButton = <Button outline color='primary' 
      onClick={this.addGroupBy} style={{ marginTop: '3px' }}>
      + Breakdown
    </Button>;

    let groupBysRow = <Row style={{marginBottom: '15px'}}>
      <Col xs='12' md='12'>
        <div style={{ marginBottom: '15px' }} hidden={this.state.groupBys.length == 0}>
          <span style={LABEL_STYLE}> breakdown by </span>
        </div>
        { [ groupBys, addGroupByButton ] }
      </Col>  
    </Row>

    return groupBysRow;
  }

  renderDashboardDropdownOptions() {
    let dashboardsDropdown = [];
    for(let i=0; i<this.props.dashboards.length; i++){
      let dashboard = this.props.dashboards[i];
      if (dashboard) {
        dashboardsDropdown.push(<DropdownItem onClick={this.selectDashboardToAdd} 
          value={dashboard.id}>{dashboard.name}</DropdownItem>)
      }
    }
    
    return dashboardsDropdown;
  }

  renderRunQuery() {
    return (
      <Row style={{marginBottom: '15px'}}>
        <div style={{width:'100%', textAlign: 'center'}}>
          <Button color='primary' style={{fontSize: '0.9rem', padding: '8px 18px', fontWeight: '500'}} 
            onClick={() =>  this.run(this.state.selectedPresentation)}>Run Query</Button>
        </div>  
      </Row>
    );
  }

  renderInsightsQueryBuilder() {
    return (
      <div>
        <Row style={{marginBottom: '15px'}}>
          <Col xs='12' md='12'>        
            <span style={LABEL_STYLE}> Show </span>
            <div style={{display: 'inline-block', width: '85px', marginRight: '10px'}} className='fapp-select light'>
              <Select
                value={this.state.aggr}
                // onChange={}
                options={[AGGR_OPTS]}
                placeholder='Function'
              />
            </div>
            <span style={LABEL_STYLE}>of</span>
            <div style={{display: 'inline-block', width: '168px', marginRight: '10px'}} className='fapp-select light'>
              <Select
                value={this.state.type}
                onChange={this.handleTypeChange}
                options={this.getQueryTypeOptsByClass()}
                placeholder='Type'
              />
            </div>
            <span style={LABEL_STYLE} hidden={this.state.type.value == TYPE_UNIQUE_USERS}> matches the following events, </span>
            <span style={LABEL_STYLE} hidden={this.state.type.value == TYPE_EVENT_OCCURRENCE}> who performed </span>
            <div style={{display: 'inline-block', width: '80px', marginRight: '10px'}} className='fapp-select light' 
              hidden={this.state.type.value == TYPE_EVENT_OCCURRENCE}>
              <Select
                value={this.state.condition}
                onChange={this.handleEventsConditionChange}
                options={EVENTS_COND_OPTS}
                placeholder='Condition'
              />
            </div>
            <span style={LABEL_STYLE} hidden={this.state.type.value == TYPE_EVENT_OCCURRENCE}> 
              of the following events, 
            </span>
          </Col>
        </Row>

        { this.renderEventsWithProperties() }
        { this.renderDateRangeSelector() }
        { this.renderGroupBys() }
        { this.renderRunQuery() }
      </div>
    );
  }

  renderInsightsPresentationOptions = () => {
    return (
      <ButtonGroup style={{ marginRight: '10px' }}>
        <button className={this.getPresentationSelectorClass(PRESENTATION_TABLE)} style={{fontWeight: 500}} 
          onClick={() => this.run(PRESENTATION_TABLE)}>Table</button>
        <button className={this.getPresentationSelectorClass(PRESENTATION_BAR)}  style={{fontWeight: 500}} 
          onClick={() => this.run(PRESENTATION_BAR)}>Bar</button>
        <button className={this.getPresentationSelectorClass(PRESENTATION_LINE)}  style={{fontWeight: 500}}  
          onClick={() => this.run(PRESENTATION_LINE)}>Line</button>
      </ButtonGroup>
    );
  } 

  jsonToCSV = () => {
    const csvRows = [];
    let newJSON = {...this.state.result}
    if (newJSON.headers[0] == "_group_key_0" || newJSON.headers[0] == "datetime") {
      newJSON = this.convertLineJSON(newJSON)
    }
    csvRows.push((newJSON.headers).join(','));
    let jsonRows = (newJSON.rows).map((row)=> {
      let values = row.map((val)=>{
        const escaped = (''+val).replace(/"/g,'\\"');
        return `"${escaped}"`
      })
      return values.join(',');
    });
    jsonRows = jsonRows.join('\n')
    const csv = csvRows+ "\n"+ jsonRows
    this.downloadCSV(csv);
  }
  convertLineJSON = (data) => {
    let datetimeKey = 0;
    for(var i = 0; i<data.meta.query.gbp.length; i++) {
      data.headers[i] = data.meta.query.gbp[i].pr
    }
    for (var i=0; i<data.headers.length; i++) {
      if(data.headers[i] == "datetime") {
        datetimeKey = i;
        if(data.meta.query.gbt == "date"){
          data.headers[i] = "date(UTC)"
        } else {
          data.headers.splice(i,1,"date", "time")
        }

      }
    }
    data.rows = data.rows.map((row)=> {
      let dateTime= row[dateTime].split("T")
      if(data.meta.query.gbt == "date"){
        row[datetimeKey] = dateTime[0]
      }
      else {
        let time = (dateTime[1].split("+"))[0] +" GMT"
        row.splice(datetimeKey, 1, dateTime[0],time)
      }
      return row
    })
    return data
  }
  downloadCSV = (data) => {
    const blob = new Blob([data], {type: 'text/csv'})
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.setAttribute('hidden','')
    a.setAttribute('href', url)
    a.setAttribute('download', 'factors_insights.csv')
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
  }
  renderDownloadButton = () => {
    if(this.state.selectedPresentation != PRESENTATION_BAR) {
      return (
        <button className="btn btn-primary ml-1" style={{fontWeight: 500}} 
                  onClick={()=> this.jsonToCSV()}>Download</button>
      )
    }
      return (
        <button className="btn btn-primary ml-1" style={{display: "none"}} 
                  onClick={()=> this.jsonToCSV()}>Download</button>
      )
  }
  renderPresentationPane(presentationOptionsByClass=null, presentationByClass=null) {
    return (
      <div>
        <div style={{borderTop: '1px solid rgb(221, 221, 221)', paddingTop: '20px', marginTop: '30px', 
          marginLeft: '-60px', marginRight: '-60px'}} hidden={ !this.state.showPresentation }></div>
        <div style={{ minHeight: '530px' }}>
          <Row style={{ marginTop: '15px', marginRight: '10px' }} hidden={ !this.state.showPresentation }>
            <Col xs='12' md='12'>
              <ButtonToolbar className='pull-right'>
                { presentationOptionsByClass == null ? null : presentationOptionsByClass() }
                <ButtonDropdown isOpen={this.state.showDashboardsList} toggle={this.toggleDashboardsList} >
                  <DropdownToggle disabled={this.disableAddToDashboard()} caret outline color="primary">
                    Add to dashboard
                  </DropdownToggle>
                  <DropdownMenu style={{height: 'auto', maxHeight: '210px', overflowX: 'scroll'}} right>
                    { this.renderDashboardDropdownOptions() }
                  </DropdownMenu>
                </ButtonDropdown>
                {this.renderDownloadButton()}
              </ButtonToolbar>
            </Col>
          </Row>
          <Row style={{ marginTop: '60px' }}> 
            <Col xs='12' md='12' > { presentationByClass() } </Col>
          </Row>
        </div>
        <div ref={(el) => { this.endOfPresentation = el; }}></div>
      </div>
    );
  }

  renderGlobalError() {
    return (
      <div className='fapp-error' style={{marginBottom: '15px'}} hidden={!this.state.topError}>
        <span>{ this.state.topError }</span>
      </div>
    );
  }

  getInterfaceSelectorStyle(queryClass) {
    let style = { display: 'inline-block', fontSize: '15px', fontWeight: '600', 
    border: '1px solid', padding: '10px 20px', borderWidth: '0.1rem', borderRadius: '5px', 
    marginRight: '18px', cursor: 'pointer' }

    if (queryClass == this.state.class.value) {
      style.borderColor = '#20a8d8';
    } else {
      style.borderColor = '#DDD';
    }

    return style; 
  }

  renderInterfaceSelector() {
    return (
      <Row style={{ marginBottom: '16px' }}>
        <Col xs='12' md='12'>
          <div style={{ textAlign: 'center', marginBottom: '15px' }}>
            <div onClick={() => this.handleClassChange({ value: QUERY_CLASS_INSIGHTS })} 
              style={this.getInterfaceSelectorStyle(QUERY_CLASS_INSIGHTS)}> 
              <img src={insightsSVG} style={{ marginRight: '5px',  marginBottom: '4px', height: '25px' }} />  
              <span className='fapp-text'> Insights </span> 
            </div>
            <div onClick={() => this.handleClassChange({ value: QUERY_CLASS_FUNNEL })} 
              style={this.getInterfaceSelectorStyle(QUERY_CLASS_FUNNEL)}>
              <img src={funnelSVG} style={{ marginRight: '5px', marginBottom: '2px', height: '25px' }} /> 
              <span className='fapp-text'> Funnel </span> 
            </div>
            <div onClick={() => this.handleClassChange({ value: QUERY_CLASS_CHANNEL })} 
              style={this.getInterfaceSelectorStyle(QUERY_CLASS_CHANNEL)}>
              <img src={channelSVG} style={{ height: '26px' }} /> 
              <span style={{ marginLeft: '5px' }} className='fapp-text'> Channels </span> 
            </div>
            <div onClick={() => this.handleClassChange({ value: QUERY_CLASS_ATTRIBUTION })}
              style={this.getInterfaceSelectorStyle(QUERY_CLASS_ATTRIBUTION)}>
              <img src={attributionSVG} style={{ height: '26px' }} /> 
              <span style={{ marginLeft: '5px' }} className='fapp-text'> Attribution </span> 
            </div>
          </div>
        </Col>
      </Row>
    );
  }

  renderAddToDashboardModal() {
    return (
      <Modal isOpen={this.state.showAddToDashboardModal} toggle={this.toggleAddToDashboardModal} style={{marginTop: '10rem'}}>
        <ModalHeader toggle={this.toggleAddToDashboardModal}>Add to Dashboard</ModalHeader>
        <ModalBody style={{padding: '25px 35px'}}>
          <div style={{textAlign: 'center', marginBottom: '15px'}}>
            <span style={{display: 'inline-block'}} className='fapp-error' hidden={this.state.addToDashboardMessage == null}>
              { this.state.addToDashboardMessage }
            </span>
          </div>
          <Form>
            <span className='fapp-label'>Title</span>         
            <Input className='fapp-input' type="text" placeholder="Your Title" onChange={this.setDashboardUnitTitle} />
          </Form>
        </ModalBody>
        <ModalFooter style={{borderTop: 'none', paddingBottom: '30px', paddingRight: '35px'}}>
          <Button outline color="success" onClick={this.addToDashboard}>Add</Button>
          <Button outline color='danger' onClick={this.toggleAddToDashboardModal}>Cancel</Button>
        </ModalFooter>
      </Modal>
    );
  }

  renderInsightsQueryInterface = () => {
    return [
      this.renderInsightsQueryBuilder(),       
      this.renderPresentationPane(
        this.renderInsightsPresentationOptions, 
        this.renderInsightsPresentation,
      )
    ];
  }

  renderFunnelQueryBuilder() {
    return (
      <div>
        <Row style={{marginBottom: '15px'}}>
          <Col xs='12' md='12'>        
            <span style={LABEL_STYLE}> Show conversion of </span>
            <div style={{display: 'inline-block', width: '168px', marginRight: '10px'}} className='fapp-select light'>
              <Select
                value={this.state.type}
                onChange={this.handleTypeChange}
                options={this.getQueryTypeOptsByClass()}
                placeholder='Type'
              />
            </div>
            <span style={LABEL_STYLE} hidden={this.state.type.value != TYPE_UNIQUE_USERS}> who has perfomed events on the below given order, </span>
            <span style={LABEL_STYLE} hidden={this.state.type.value != TYPE_EVENT_OCCURRENCE}> on the below given order, </span>
          </Col>
        </Row>

        { this.renderEventsWithProperties() }
        { this.renderDateRangeSelector() }
        { this.renderGroupBys() }
        { this.renderRunQuery() }
      </div>
    );
  }

  getFunnelResultAsFunnel() {
    return <div style={{height: '450px'}} className='animated fadeIn'> 
      <FunnelChart queryResult={this.state.result} /> 
    </div>;
  }

  getFunnelResultAsTable() {
    return <div style={{height: '450px'}} className='animated fadeIn'> 
      <TableChart queryResult={ convertFunnelResultForTable(this.state.result) } /> 
    </div>;
  }

  renderFunnelPresentation = () => {
    if (this.state.isResultLoading) return <Loading paddingTop='14%' />;
    if (this.state.result == null) return null;

    let selected = this.state.selectedPresentation;
    
    if (selected == PRESENTATION_TABLE) {
      return this.getFunnelResultAsTable();
    }

    return this.getFunnelResultAsFunnel();
  }

  renderFunnelPresentationOptions = () => {
    return (
      <ButtonGroup style={{ marginRight: '10px' }}>
        <button className={this.getPresentationSelectorClass(PRESENTATION_TABLE)} style={{fontWeight: 500}} 
          onClick={() => this.run(PRESENTATION_TABLE)}>Table</button>
        <button className={this.getPresentationSelectorClass(PRESENTATION_FUNNEL)}  style={{fontWeight: 500}} 
          onClick={() => this.run(PRESENTATION_FUNNEL)}>Funnel</button>
      </ButtonGroup>
    );
  }

  renderFunnelQueryInterface = () => {
    return [
      this.renderFunnelQueryBuilder(),
      this.renderPresentationPane(
        this.renderFunnelPresentationOptions, 
        this.renderFunnelPresentation,
      )
    ];
  }

  renderChannelReportsInterface = () => {
    return <ChannelQuery />;
  }

  renderAttributionInterface = () =>{ 
    return <AttributionQuery showError={(err)=>{this.showTopError(err)}} resetError={()=>this.resetTopError()}/>
 }

  render() {
    if (!this.isLoaded()) return <Loading />;
    var renderQueryInterface = this.renderInsightsQueryInterface;

    if (this.state.class.value == QUERY_CLASS_FUNNEL) {
      renderQueryInterface = this.renderFunnelQueryInterface;
    }

    if (this.state.class.value == QUERY_CLASS_CHANNEL) {
      renderQueryInterface = this.renderChannelReportsInterface;
    }

    if (this.state.class.value == QUERY_CLASS_ATTRIBUTION){
      renderQueryInterface = this.renderAttributionInterface;
    }

    console.debug('Query State : ', this.state);
    return (
      <div className='fapp-content' style={{ marginLeft: '2rem', marginRight: '2rem', paddingTop: '30px' }}>
        {[ this.renderInterfaceSelector(), this.renderGlobalError(), renderQueryInterface(), this.renderAddToDashboardModal() ]}
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Query);