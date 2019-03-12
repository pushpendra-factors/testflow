import React, { Component } from 'react';
import Select from 'react-select';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Row, Col, Button, Table } from 'reactstrap';
import 'react-date-range/dist/styles.css';
import 'react-date-range/dist/theme/default.css';
import { DateRangePicker, defaultStaticRanges, createStaticRanges } from 'react-date-range'; 
import moment from 'moment';

import { 
  fetchProjectEvents,
  runQuery,
} from '../../actions/projectsActions';

import Event from './Event';
import GroupBy from './GroupBy';
import { trimQuotes } from '../../util'

const ANALYSIS_TYPE_OPTS = [
  { value: 'events_occurrence', label: 'Events occurrence' },
  { value: 'unique_users', label: 'Unique users' }
];

const DEFAULT_DATE_RANGE = {
  startDate: moment(new Date()).subtract(7, 'days').toDate(),
  endDate: new Date(),
  label: 'Last 7 days',
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
    label: 'Last 7 days',
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

const PRESENTATION_TABLE = 'table';
const PRESENTATION_LINE =  'line';

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    projects: store.projects.projects,
    eventNames: store.projects.currentProjectEventNames,

    eventPropertiesMap: store.projects.queryEventPropertiesMap,
    eventPropertyValuesMap: store.projects.queryEventPropertyValuesMap
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchProjectEvents
  }, dispatch)
}

class Query extends Component {
  constructor(props) {
    super(props);

    this.state = {
      eventNamesLoaded: false,
      eventNamesLoadError: null,

      type: ANALYSIS_TYPE_OPTS[0], // 1st type as default.
      events: [],
      groupBys: [],
      resultDateRange: [DEFAULT_DATE_RANGE],

      result: null,
      resultError: null,
      selectedPresentation: PRESENTATION_TABLE,

      showDatePicker: false,
    }
  }

  componentWillMount() {
    this.props.fetchProjectEvents(this.props.currentProjectId)
      .then(() => this.setState({ eventNamesLoaded: true }))
      .catch((r) => this.setState({ eventNamesLoaded: true, eventNamesLoadError: r.paylaod }));
  }

  handleTypeChange = (option) => {
    this.setState({type: option});
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
    return { type: '',  name: '', op: '', value: ''};
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

  onPropertyTypeChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'type', value)
  }

  onPropertyNameChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'name', value)
  }

  onPropertyOpChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'op', value)
  }

  onPropertyValueChange = (eventIndex, propertyIndex, value) => {
    this.setPropertyAttr(eventIndex, propertyIndex, 'value', value)
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

  getQuery() {
    let query = {};
    query.type = this.state.type.value;
    query.eventsCondition = 'all'; // Todo(Dinesh): Add a selector. Make it part of the query.

    if (this.state.resultDateRange.length == 0)
      throw new Error('Invalid date range. No default range given.')
    
    query.from = moment(this.state.resultDateRange[0].startDate).unix(); // in utc.
    query.to = moment(this.state.resultDateRange[0].endDate).unix(); // in utc.

    query.eventsWithProperties = []
    for(let ei=0; ei < this.state.events.length; ei++) {
      let event = this.state.events[ei];
      
      let ewp = {};
      ewp.name = event.name;
      ewp.properties = [];

      for(let pi=0; pi < event.properties.length; pi++) {
        let property = event.properties[pi];
        let cProperty = {}
        
        if (property.type != '' && property.name != '' &&
            property.operator != '' && property.value != '') {

            cProperty.entity = property.type;
            cProperty.property = property.name;
            cProperty.operator = property.op;
            cProperty.value = property.value;
            ewp.properties.push(cProperty);
        }
      }
      query.eventsWithProperties.push(ewp)
    }

    query.groupByProperties = [];
    for(let i=0; i < this.state.groupBys.length; i++) {
      let groupBy = this.state.groupBys[i];
      let cGroupBy = {};

      if (groupBy.name != '' && groupBy.type != '') {
        cGroupBy.property = groupBy.name;
        cGroupBy.entity = groupBy.type;
        cGroupBy.index = i;
        query.groupByProperties.push(cGroupBy)
      }
    }

    console.debug(query);
    return query
  }

  run = () => {
    runQuery(this.props.currentProjectId, this.getQuery())
      .then((r) => this.handleResultResponse(r.data))
      .catch(console.error);
  }

  handleResultResponse(result) {
    if (result.error) {
      this.setState({ resultError: result.error });
      return
    }
    
    if (!result.headers)
      throw new Error('Query result headers not found.')

    // Rewrite groupKeys with actual names.
    // Todo: Move this to backend. Susceptible to errors.
    for(let i=0; i < result.headers.length; i++) {
      let header = result.headers[i];
      if (header.indexOf('gk_') > -1) {
        let groupByIndex = parseInt(header.split('_')[1]);
        let groupBy = this.state.groupBys[groupByIndex]
        if (groupBy && groupBy.name) {
          result.headers[i] = groupBy.name;  
        } 
        else {
          throw new Error('Mismatch of groupKeys on state and result');
        }
      }
    }

    this.setState({ result: result });
  }

  displayPresentationPane() {
    return this.state.result != null ? 'block' : 'none';
  }

  setSelectedPresentation = (type) => {
    if (this.state.selectedPresentation != type)
      this.setState({ selectedPresentation: type });
  }

  displayDatePicker() {
    return this.state.showDatePicker ? 'inline-block' : 'none';
  }

  getResultAsTable() {
    if (this.state.result == null) return;

    let result = this.state.result;
    let headers = result.headers.map((h, i) => { return <th key={'header_'+i}>{h}</th> });
    let rows = [];

    for(let i=0; i<Object.keys(result.rows).length; i++) {
      let cols = result.rows[i.toString()];
      if (cols != undefined) {
        let tds = cols.map((c) => { return <td> { trimQuotes(c) } </td> });
        rows.push(<tr>{tds}</tr>);
      }
    }

    return (
      <Table className='fapp-table' style={{textAlign: 'center'}}> 
        <thead>
          <tr> { headers } </tr>
        </thead>
        <tbody>
          { rows }
        </tbody>
      </Table>
    );
  }

  getResultAsLineChart() {}
  
  getPresentableResultByType() {
    if (this.state.selectedPresentation == PRESENTATION_TABLE) {
      return this.getResultAsTable();
    }
  }

  getPresentationSelectorClass(type) {
    return this.state.selectedPresentation == type ? 'btn btn-primary' : 'btn btn-outline-primary';
  }

  getEventNames = () => {
    return this.state.events.map((e) => { return e.name; })
  }
  
  render() {
    let events = [];
    for(let i=0; i<this.state.events.length; i++) {
      events.push(
        <Event 
          index={i}
          key={'events_'+i} 
          projectId={this.props.currentProjectId} 
          nameOpts={this.props.eventNames} 
          eventState={this.state.events[i]}
          // event handlers
          onNameChange={(value) => this.onEventStateChange(value, i)} 
          // property handlers
          onAddProperty={() => this.addProperty(i)}
          onPropertyTypeChange={this.onPropertyTypeChange}
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
          projectId={this.props.currentProjectId}
          getSelectedEventNames={this.getEventNames}
          groupByState={this.state.groupBys[i]}
          onTypeChange={(option) => this.onGroupByTypeChange(i, option)}
          onNameChange={(option) => this.onGroupByNameChange(i, option)}
        />
      );
    }

    console.debug('Query State : ', this.state);

    return (
      <div>
        {/* Query */}
        <div>
          <Row style={{marginBottom: '15px'}}>
            <Col xs='12' md='12'>              
              <span style={{marginRight: '10px'}}> Get </span>
              <div style={{display: 'inline-block', width: '15%'}}>
                <Select
                  value={this.state.type}
                  onChange={this.handleTypeChange}
                  options={ANALYSIS_TYPE_OPTS}
                  placeholder='Type'
                />
              </div>
              <Button outline color='primary' style={{marginLeft: '10px'}} onClick={this.addEvent}>+ Event</Button>
            </Col>
          </Row>
          { events }
          <Row style={{marginBottom: '15px'}}>
            <Col xs='12' md='12'>
              <span style={{marginRight: '10px'}}> During </span>
              <Button outline style={{border: '1px solid grey', color: 'grey', marginRight: '10px' }} onClick={this.toggleDatePickerDisplay}><i class="fa fa-calendar" style={{marginRight: '10px'}}></i>{this.readableDateRange(this.state.resultDateRange[0])}</Button>
              <div  style={{ border: '1px solid rgb(239, 242, 247)', position: 'absolute', zIndex: '100', display: this.displayDatePicker() }}>
                <DateRangePicker
                  ranges={this.state.resultDateRange}
                  onChange={this.handleResultDateRangeSelect}
                  staticRanges={ DEFINED_DATE_RANGES }
                  inputRanges={[]}
                  minDate={new Date('01 Jan 2010 00:00:00 GMT')} // range starts from given date.
                />
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
            <Col sm='12' md={{ size: 'auto', offset: 5 }} style={{textAlign: 'center'}}>
              <Button color='primary' style={{fontSize: '0.9rem', padding: '8px 18px', fontWeight: '500'}} onClick={this.run}>Run Query</Button>
            </Col>  
          </Row>
        </div>
        
        {/* Presentation */}
        <div style={{ display: this.displayPresentationPane() }}>
          <Row>
            <Col xs='12' md='12'>
              <div>
                <div class='pull-right'>
                  <div style={{ marginBottom: '10px'}}>
                    <button className={this.getPresentationSelectorClass(PRESENTATION_TABLE)} style={{marginRight: '10px', fontWeight: 500}} onClick={() => this.setSelectedPresentation(PRESENTATION_TABLE)}>Table View</button>
                    <button className={this.getPresentationSelectorClass(PRESENTATION_LINE)}  style={{marginRight: '10px', fontWeight: 500}} onClick={() => this.setSelectedPresentation(PRESENTATION_LINE)}>Line Chart</button>
                  </div>
                </div>
              </div>
            </Col>
          </Row>
          <Row>
            <Col xs='12' md='12' style={{marginTop: '20px'}}>
              <div>
                { this.getPresentableResultByType() }
              </div>
            </Col>
          </Row>
        </div>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Query);