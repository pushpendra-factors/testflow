import React, { Component } from 'react';
import Select from 'react-select';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Row, Col, Button, Table, ButtonGroup, ButtonToolbar, 
  ButtonDropdown, DropdownToggle, DropdownMenu, DropdownItem } from 'reactstrap'; 
import 'react-date-range/dist/styles.css';
import 'react-date-range/dist/theme/default.css';
import { DateRangePicker, createStaticRanges } from 'react-date-range'; 
import moment from 'moment';

import LineChart from './LineChart';
import BarChart from './BarChart';
import TableChart from './TableChart'
import { PRESENTATION_BAR, PRESENTATION_LINE, PRESENTATION_TABLE } from './common';
import { 
  fetchProjectEvents,
  runQuery,
} from '../../actions/projectsActions';
import { fetchDashboards, createDashboardUnit } from '../../actions/dashboardActions';
import Event from './Event';
import GroupBy from './GroupBy';
import { removeElementByIndex, getSelectedOpt, isNumber, createSelectOpts } from '../../util'
import TableBarChart from './TableBarChart';
import Loading from '../../loading';
import factorsai from '../../factorsaiObj';

const COND_ALL_GIVEN_EVENT = 'all_given_event';
const COND_ANY_GIVEN_EVENT = 'any_given_event'; 
const EVENTS_COND_OPTS = [
  { value: COND_ALL_GIVEN_EVENT, label: 'All given event' },
  { value: COND_ANY_GIVEN_EVENT, label: 'Any given event' }
];

const TYPE_EVENT_OCCURRENCE = 'events_occurrence';
const TYPE_UNIQUE_USERS = 'unique_users';
const ANALYSIS_TYPE_OPTS = [
  { value: TYPE_EVENT_OCCURRENCE, label: 'Events occurrence' },
  { value: TYPE_UNIQUE_USERS, label: 'Unique users' }
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

const PROPERTY_TYPE_OPTS = {
  'user': 'User Property',
  'event': 'Event Property'
};

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

      condition: EVENTS_COND_OPTS[0],
      type: ANALYSIS_TYPE_OPTS[0], // 1st type as default.
      events: [],
      groupBys: [],
      resultDateRange: [DEFAULT_DATE_RANGE],

      result: null,
      resultError: null,
      isResultLoading: true,
      selectedPresentation: null,

      showPresentation: false,
      showDatePicker: false,
      topError: null,
      showDashboardsList: false,
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
    return { entity: '',  name: '', op: '', value: '', valueType: '' }; 
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
    return { type: '', name: '' };
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

  setQueryDuration(query) {
    let selectedRange = this.state.resultDateRange[0];
    let isEndDateToday = selectedRange.endDate.getDate() == (new Date()).getDate();
    let from =  moment(selectedRange.startDate).unix();
    let to = moment(selectedRange.endDate).unix();

    // Adjust the duration window respective to current time.
    if (isEndDateToday) {
      let diff = to - from;
      to =  moment(new Date()).unix();
      from = to - diff;
    }

    query.fr = from; // in utc.
    query.to = to; // in utc.
  }

  getQuery(groupByDate=false) {
    let query = {};
    query.ty = this.state.type.value;
    query.ec = this.state.condition.value;
    // event_occurrence supports only any_given_event.
    if (query.ty == TYPE_EVENT_OCCURRENCE) {
      query.ec = COND_ANY_GIVEN_EVENT;
    }
    
    if (this.state.resultDateRange.length == 0)
      throw new Error('Invalid date range. No default range given.')
    
    this.setQueryDuration(query);

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
    return <TableChart queryResult={this.state.result} />;
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

  getPresentableResult() {
    if (this.state.isResultLoading) return <Loading paddingTop='10%' />;

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
    if (this.state.type.value == TYPE_UNIQUE_USERS) {
      return createSelectOpts({'user': PROPERTY_TYPE_OPTS['user']});
    } else {
      return createSelectOpts(PROPERTY_TYPE_OPTS);
    }
  }

  addToDashboard = (event) => {
    let dashboardId = event.currentTarget.getAttribute('value');

    if (this.state.selectedPresentation == null) {
      console.error('Invalid presentation');
      return;
    }
    let presentation = this.state.selectedPresentation;
    let query = this.getQuery(presentation === PRESENTATION_LINE);
    let payload = {
      presentation: presentation,
      query: query,
      title: "Chart Title", // Use modal to get chart title from user.
    };
    this.props.createDashboardUnit(this.props.currentProjectId, dashboardId, payload)
      .then(console.log)
      .catch(console.error);
  }

  render() {
    if (!this.isLoaded()) return <Loading />;

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

    console.debug('Query State : ', this.state);

    let dashboardsDropdown = [];
    for(let i=0; i<this.props.dashboards.length; i++){
      let dashboard = this.props.dashboards[i];
      if (dashboard) {
        dashboardsDropdown.push(<DropdownItem onClick={this.addToDashboard} value={dashboard.id}>{dashboard.name}</DropdownItem>)
      }
    }

    return (
      <div className='fapp-content' style={{marginLeft: '2rem', marginRight: '2rem'}}>
        <div className='fapp-error' style={{marginBottom: '15px'}} hidden={!this.state.topError}>
            <span>{ this.state.topError }</span>
        </div>

        {/* Query */}
        <div>
          <Row style={{marginBottom: '15px'}}>
            <Col xs='12' md='12'>        
              <span style={{marginRight: '10px'}}> Get </span>
              <div style={{display: 'inline-block', width: '15%', marginRight: '10px'}} className='fapp-select'>
                <Select
                  value={this.state.type}
                  onChange={this.handleTypeChange}
                  options={ANALYSIS_TYPE_OPTS}
                  placeholder='Type'
                />
              </div>
              <span style={{marginRight: '10px'}} hidden={this.state.type.value == TYPE_EVENT_OCCURRENCE}> who performed </span>
              <div style={{display: 'inline-block', width: '15%', marginRight: '10px'}} className='fapp-select' hidden={this.state.type.value == TYPE_EVENT_OCCURRENCE}>
                <Select
                  value={this.state.condition}
                  onChange={this.handleEventsConditionChange}
                  options={EVENTS_COND_OPTS}
                  placeholder='Condition'
                />
              </div>
            </Col>
          </Row>
          { events }
          <Row style={{marginBottom: '15px'}}>
            <Col xs='12' md='12' style={{marginLeft: '70px'}}>
              <Button outline color='primary' onClick={this.addEvent}>+ Event</Button>
            </Col>
          </Row>
          <Row style={{marginBottom: '15px'}}>
            <Col xs='12' md='12'>
              <span style={{marginRight: '10px'}}> During </span>
              <Button outline style={{border: '1px solid grey', color: 'grey', marginRight: '10px' }} onClick={this.toggleDatePickerDisplay}><i class="fa fa-calendar" style={{marginRight: '10px'}}></i>{this.readableDateRange(this.state.resultDateRange[0])}</Button>
              <div class='fapp-date-picker' hidden={!this.state.showDatePicker}>
                <DateRangePicker
                  ranges={this.state.resultDateRange}
                  onChange={this.handleResultDateRangeSelect}
                  staticRanges={ DEFINED_DATE_RANGES }
                  inputRanges={[]}
                  minDate={new Date('01 Jan 2000 00:00:00 GMT')} // range starts from given date.
                  maxDate={new Date()}
                />
                <button className='fapp-close-round-button' style={{float: 'right', marginLeft: '0px', borderLeft: 'none'}} onClick={this.toggleDatePickerDisplay}>x</button>
              </div>
            </Col>
          </Row>
          <Row style={{marginBottom: '15px'}}>
            <Col xs='12' md='12'>
              <span style={{marginRight: '10px'}}>Group by</span>
              <Button outline color='primary' onClick={this.addGroupBy}>+ Group By</Button>
              {groupBys}
            </Col>  
          </Row>
          <Row style={{marginBottom: '15px'}}>
            <div style={{width:'100%', textAlign: 'center'}}>
              <Button color='primary' style={{fontSize: '0.9rem', padding: '8px 18px', fontWeight: '500'}} onClick={() =>  this.run(DEFAULT_PRESENTATION)}>Run Query</Button>
            </div>  
          </Row>
        </div>
        

        <div style={{borderTop: '1px solid rgb(221, 221, 221)', paddingTop: '20px', marginTop: '25px', marginLeft: '-60px', marginRight: '-60px'}} hidden={ !this.state.showPresentation }></div>

        {/* Presentation */}
        <div hidden={ !this.state.showPresentation }>
          <Row>
            <Col xs='12' md='12'>
              <ButtonToolbar class='pull-right' style={{ marginBottom: '10px' }}>
                <ButtonGroup style={{ marginRight: '10px' }}>
                  <button className={this.getPresentationSelectorClass(PRESENTATION_TABLE)} style={{fontWeight: 500}} onClick={() => this.run(PRESENTATION_TABLE)}>Table</button>
                  <button className={this.getPresentationSelectorClass(PRESENTATION_BAR)}  style={{fontWeight: 500}} onClick={() => this.run(PRESENTATION_BAR)}>Bar</button>
                  <button className={this.getPresentationSelectorClass(PRESENTATION_LINE)}  style={{fontWeight: 500}} onClick={() => this.run(PRESENTATION_LINE)}>Line</button>
                </ButtonGroup>
                <ButtonDropdown isOpen={this.state.showDashboardsList} toggle={this.toggleDashboardsList} >
                  <DropdownToggle caret outline color="primary">
                    Dashboard
                  </DropdownToggle>
                  <DropdownMenu right>
                    { dashboardsDropdown }
                  </DropdownMenu>
                </ButtonDropdown>
              </ButtonToolbar>
            </Col>
          </Row>
          <Row>
            <Col xs='12' md='12' style={{marginTop: '20px'}}>
                { this.getPresentableResult() }
            </Col>
          </Row>
        </div>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Query);