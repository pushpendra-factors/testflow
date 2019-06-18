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
import { DateRangePicker, createStaticRanges } from 'react-date-range'; 
import moment from 'moment';

import TableChart from './TableChart'
import LineChart from './LineChart';
import BarChart from './BarChart';
import TableBarChart from './TableBarChart';
import Funnel from './Funnel';
import { PRESENTATION_BAR, PRESENTATION_LINE, PRESENTATION_TABLE, PRESENTATION_CARD, isLineResultWithGroupBy } from './common';
import { 
  fetchProjectEvents,
  runQuery,
} from '../../actions/projectsActions';
import { fetchDashboards, createDashboardUnit } from '../../actions/dashboardActions';
import Event from './Event';
import GroupBy from './GroupBy';
import { 
  removeElementByIndex, getSelectedOpt, isNumber, createSelectOpts, 
  isSingleCountResult, slideUnixTimeWindowToCurrentTime,
} from '../../util'
import Loading from '../../loading';
import factorsai from '../../common/factorsaiObj';
import { PROPERTY_TYPE_OPTS } from './common';

const COND_ALL_GIVEN_EVENT = 'all_given_event';
const COND_ANY_GIVEN_EVENT = 'any_given_event'; 
const EVENTS_COND_OPTS = [
  { value: COND_ALL_GIVEN_EVENT, label: 'all' },
  { value: COND_ANY_GIVEN_EVENT, label: 'any' }
];
const LABEL_STYLE = { marginRight: '10px', fontWeight: '600', color: '#777' };

const QUERY_CLASS_INSIGHTS = 'insights';
const QUERY_CLASS_FUNNEL = 'funnel';
const QUERY_CLASS_OPTS = [
  { value: QUERY_CLASS_INSIGHTS, label: 'Insights' },
  { value: QUERY_CLASS_FUNNEL, label: 'Funnel' }
];

const TYPE_EVENT_OCCURRENCE = 'events_occurrence';
const TYPE_UNIQUE_USERS = 'unique_users';
const INSIGHTS_QUERY_TYPE_OPTS = [
  { value: TYPE_EVENT_OCCURRENCE, label: 'events occurrence' },
  { value: TYPE_UNIQUE_USERS, label: 'unique users' },
];
const FUNNEL_QUERY_TYPE_OPTS = [
  { value: TYPE_UNIQUE_USERS, label: 'unique users' },
];

const DEFAULT_DATE_RANGE_LABEL = 'Last 7 days';
const DEFAULT_DATE_RANGE = {
  startDate: moment(new Date()).subtract(7, 'days').toDate(),
  endDate: new Date(),
  label: DEFAULT_DATE_RANGE_LABEL,
  key: 'selected'
}
const DEFINED_DATE_RANGES = createStaticRanges([
  {
    label: 'Last 24 hours',
    range: () => ({
      startDate: moment(new Date()).subtract(24, 'hours').toDate(),
      endDate: new Date(),
    }),
  },
  {
    label: DEFAULT_DATE_RANGE_LABEL,
    range: () => ({
      startDate: DEFAULT_DATE_RANGE.startDate,
      endDate: DEFAULT_DATE_RANGE.endDate
    }),
  },
  {
    label: 'Last 30 days',
    range: () => ({
      startDate: moment(new Date()).subtract(30, 'days').toDate(),
      endDate: new Date(),
    })
  },
]);

const ERROR_NO_EVENT = 'No events given. Please add atleast one event by clicking +Event button.';

const DEFAULT_PRESENTATION = PRESENTATION_TABLE;

const HEADER_COUNT = "count";
const HEADER_DATE = "date";

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    projects: store.projects.projects,
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
      condition: EVENTS_COND_OPTS[0],
      type: INSIGHTS_QUERY_TYPE_OPTS[0], // 1st type as default.
      events: [],
      groupBys: [],
      resultDateRange: [DEFAULT_DATE_RANGE],

      result: null,
      resultError: null,
      isResultLoading: false,
      selectedPresentation: null,

      showPresentation: false,
      showDatePicker: false,
      topError: null,

      showDashboardsList: false,
      showAddToDashboardModal: false,
      addToDashboardMessage: null,
      inputDashboardUnitTitle: null,
      selectedDashboardId: null,
    }
  }

  getQueryTypeOptsByClass = () => {
    return this.state.class.value == QUERY_CLASS_FUNNEL ? FUNNEL_QUERY_TYPE_OPTS : INSIGHTS_QUERY_TYPE_OPTS;
  }
  
  resetQueryInterfaceOnClassChange() {
    this.setState({
      // reset query state.
      condition: EVENTS_COND_OPTS[0],
      type: this.getQueryTypeOptsByClass()[0],
      events: [],
      groupBys: [],
      resultDateRange: [DEFAULT_DATE_RANGE],
      // reset presentation.
      result: null,
      showPresentation: false,
    });

    this.initWithAnEventRow();
  }

  componentDidUpdate(prevProps, prevState) {
    if (prevState.class.value != this.state.class.value) {
      this.resetQueryInterfaceOnClassChange();
    }
  }

  componentWillMount() {
    this.props.fetchProjectEvents(this.props.currentProjectId)
      .then(() => {
        this.setState({ eventNamesLoaded: true });
        this.initWithAnEventRow();
      })
      .catch((r) => this.setState({ eventNamesLoaded: true, eventNamesLoadError: r.paylaod }));

    this.props.fetchDashboards(this.props.currentProjectId);
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
      state.events[index].name = option.value;
      return state;
    })
  }

  getDefaultPropertyState() {
    let keys = Object.keys(PROPERTY_TYPE_OPTS)
    return { entity: keys[0],  name: '', op: 'equals', value: '', valueType: '' };
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
  }

  onPropertyNameChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'name', value)
  }

  onPropertyOpChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'op', value)
  }

  onPropertyValueChange = (eventIndex, propertyIndex, value, type) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'value', value);
    this.setPropertyAttr(eventIndex, propertyIndex, 'valueType', type);
  }

  getDefaultGroupByState() {
    let groupByOpts = this.getGroupByOpts();
    return { type: groupByOpts[0].value, name: '' };
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
  }

  onGroupByNameChange = (groupByIndex, option) => {
    this.setGroupByAttr(groupByIndex, 'name', option.value);
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

  setQueryPeriod(query, toSave=false) {
    let selectedRange = this.state.resultDateRange[0];
    let isEndDateToday = moment(selectedRange.endDate).isSame(moment(), 'day');
    let from =  moment(selectedRange.startDate).unix();
    let to = moment(selectedRange.endDate).unix();

    // Adjust the duration window respective to current time.
    if (isEndDateToday) {
      let newRange = slideUnixTimeWindowToCurrentTime(from, to)
      from = newRange.from;
      to = newRange.to;
    }

    if (toSave) query.ovp = isEndDateToday;
    query.fr = from; // in utc.
    query.to = to; // in utc.
  }

  getQuery(groupByDate=false, toSave=false) {
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
    
    this.setQueryPeriod(query, toSave);

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
        query.gbp.push(cGroupBy)
      }
    }

    if (groupByDate) {
      query.gbt = true;
    }
  
    console.debug(query);
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
  }

  showTopError(error) {
    if (!error) {
      this.setState({ topError: null });
      return;
    }

    this.setState({ topError: error });
  }

  run = (presentation) => {
    this.scrollToBottom();

    if (presentation == "")
      throw new Error('Invalid presentation');

    this.showTopError(this.validateQuery());
    
    this.setState({ isResultLoading: true, showPresentation: true });
    let query = this.getQuery(presentation === PRESENTATION_LINE);

    let eventProperties = { 
      projectId: this.props.currentProjectId,
      query: JSON.stringify(query),
      queryType: query.type,
      eventsCondition: query.eventsCondition,
      presentation: presentation,
    };
    let startTime = new Date().getTime();
    
    runQuery(this.props.currentProjectId, query)
      .then((r) => {
        if(this.isResponseValid(r.data)) {
          this.setState({ 
            result: r.data, 
            selectedPresentation: presentation,
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
    if (this.state.type.value == TYPE_UNIQUE_USERS || this.state.class.value == QUERY_CLASS_FUNNEL) {
      return createSelectOpts({'user': PROPERTY_TYPE_OPTS['user']});
    } else {
      return createSelectOpts(PROPERTY_TYPE_OPTS);
    }
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
    
    let groupByTimestamp = presentation === PRESENTATION_LINE;
    let query = this.getQuery(groupByTimestamp, true);
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
    return this.state.selectedPresentation === PRESENTATION_BAR 
      && this.state.groupBys.length > 1;
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

  renderDateRangeSelector() {
    return (
      <Row style={{marginBottom: '15px'}}>
        <Col xs='12' md='12'>
          <span style={LABEL_STYLE}> during </span>
          <Button outline style={{border: '1px solid #ccc', color: 'grey', marginRight: '10px' }} 
            onClick={this.toggleDatePickerDisplay}>
            <i className="fa fa-calendar" style={{marginRight: '10px'}}></i>
            {this.readableDateRange(this.state.resultDateRange[0])}
          </Button>
          <div className='fapp-date-picker' hidden={!this.state.showDatePicker}>
            <DateRangePicker
              ranges={this.state.resultDateRange}
              onChange={this.handleResultDateRangeSelect}
              staticRanges={ DEFINED_DATE_RANGES }
              inputRanges={[]}
              minDate={new Date('01 Jan 2000 00:00:00 GMT')} // range starts from given date.
              maxDate={new Date()}
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
          getOpts={this.getGroupByOpts}
        />
      );
    }

    let addGroupByButton = <Button outline color='primary' 
      onClick={this.addGroupBy} style={{ marginTop: '3px' }}>
      + Group
    </Button>;

    let groupBysRow = <Row style={{marginBottom: '15px'}}>
      <Col xs='12' md='12'>
        <div style={{ marginBottom: '15px' }} hidden={this.state.groupBys.length == 0}>
          <span style={LABEL_STYLE}> group by </span>
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
            onClick={() =>  this.run(DEFAULT_PRESENTATION)}>Run Query</Button>
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
                value={{label: 'count', value: 'count'}}
                // onChange={}
                options={[{label: 'count', value: 'count'}]}
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
              </ButtonToolbar>
            </Col>
          </Row>
          <Row style={{ marginTop: '45px' }}> 
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

  renderInterfaceSelector() {
    return (
      <Row style={{marginBottom: '16px'}}>
        <Col xs='12' md='12'>
          <span style={LABEL_STYLE}> Get </span>
          <div style={{display: 'inline-block', width: '100px', marginRight: '10px'}} className='fapp-select light'>
            <Select
              value={this.state.class}
              options={QUERY_CLASS_OPTS}
              onChange={this.handleClassChange}
            />
          </div>
          <span style={LABEL_STYLE}> for below query, </span>
        </Col>
      </Row>
    );
  }

  renderAddToDashboardModal() {
    return (
      <Modal isOpen={this.state.showAddToDashboardModal} toggle={this.toggleAddToDashboardModal} style={{marginTop: '10rem'}}>
        <ModalHeader toggle={this.toggleAddToDashboardModal}>Add to dashboard</ModalHeader>
        <ModalBody style={{padding: '25px 35px'}}>
          <div style={{textAlign: 'center', marginBottom: '15px'}}>
            <span style={{display: 'inline-block'}} className='fapp-error' hidden={this.state.addToDashboardMessage == null}>
              { this.state.addToDashboardMessage }
            </span>
          </div>
          <Form>
            <span className='fapp-label'>Chart title</span>         
            <Input className='fapp-input' type="text" placeholder="Your chart title" onChange={this.setDashboardUnitTitle} />
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

  getResultAsFunnel() {
    let stepsIndexes = [];
    let conversionIndexes = [];
    let conversionHeaders = [];
    let groupIndexes = [];
    let groupHeaders = [];
    
    for (let i=0; i<this.state.result.headers.length; i++) {
      if (this.state.result.headers[i].indexOf('step_') == 0)
        stepsIndexes.push(i);
      else if (this.state.result.headers[i].indexOf('conversion_') == 0) {
        conversionIndexes.push(i);
        conversionHeaders.push(this.state.result.headers[i]);
      }
      else {
        groupIndexes.push(i);
        groupHeaders.push(this.state.result.headers[i]);
      }
    }

    let rows = this.state.result.rows;
    let funnelData = [];
    for (let i=0; i<stepsIndexes.length; i++) {
      let data = null;
      if (i == 0) data = [rows[0][stepsIndexes[0]], 0];
      else data = [rows[0][stepsIndexes[i]], [rows[0][stepsIndexes[i-1]] - rows[0][stepsIndexes[i]]]];

      let comp = {};
      comp.conversion_percent = rows[0][conversionIndexes[i]];
      comp.data = data;
      funnelData.push(comp);
    }

    let showGroupsTable = groupIndexes.length > 0;
    let groupRows = [];
    if (showGroupsTable) {
      groupHeaders.push(...conversionHeaders)

      // Row 0 is $no_group.
      for(let i=1; i<this.state.result.rows.length; i++) {
        let row = [];
        // adds group values to row.
        for (let r=0; r<groupIndexes.length; r++) {
          row.push(this.state.result.rows[i][groupIndexes[r]]);
        }
        // adds conversions to row.
        for (let c=0; c<conversionIndexes.length; c++) {
          row.push(this.state.result.rows[i][conversionIndexes[c]] + '%');
        }
        groupRows.push(row);
      }
    }

    let tableGroupResult = { headers: groupHeaders, rows: groupRows };
    let present = [];
    present.push(<div style={{ marginTop: '30px' }}><Funnel data={funnelData} /></div>);
    if (showGroupsTable) present.push(<div style={{ marginTop: '50px' }}>
      <TableChart queryResult={tableGroupResult} /></div>);

    return <div style={{height: '450px'}} className='animated fadeIn'> { present } </div>;
  }

  renderFunnelPresentation = () => {
    if (this.state.isResultLoading) return <Loading paddingTop='14%' />;
    if (this.state.result == null) return null;

    return this.getResultAsFunnel();
  }

  renderFunnelQueryInterface = () => {
    return [
      this.renderFunnelQueryBuilder(),
      this.renderPresentationPane(
        null, this.renderFunnelPresentation,
      )
    ];
  }


  render() {
    if (!this.isLoaded()) return <Loading />;
    var renderQueryInterface = this.renderInsightsQueryInterface;

    if (this.state.class.value == QUERY_CLASS_FUNNEL) {
      renderQueryInterface = this.renderFunnelQueryInterface;
    }

    console.debug('Query State : ', this.state);
    return (
      <div className='fapp-content' style={{ marginLeft: '2rem', marginRight: '2rem' }}>
        {[ this.renderGlobalError(), this.renderInterfaceSelector(), renderQueryInterface(), this.renderAddToDashboardModal() ]}
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Query);