import React, { Component } from 'react';
import Select from 'react-select';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Row, Col, Button, Table } from 'reactstrap';
import 'react-date-range/dist/styles.css';
import 'react-date-range/dist/theme/default.css';
import { DateRangePicker, createStaticRanges } from 'react-date-range'; 
import moment from 'moment';

import LineChart from './LineChart';
import BarChart from './BarChart';
import { 
  fetchProjectEvents,
  runQuery,
} from '../../actions/projectsActions';
import Event from './Event';
import GroupBy from './GroupBy';
import { trimQuotes, removeElementByIndex, firstToUpperCase } from '../../util'
import TableBarChart from './TableBarChart';

const ANALYSIS_TYPE_OPTS = [
  { value: 'events_occurrence', label: 'Events occurrence' },
  { value: 'unique_users', label: 'Unique users' }
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

const PRESENTATION_TABLE = 'table';
const PRESENTATION_LINE =  'line';
const PRESENTATION_BAR = 'bar';

const DEFAULT_PRESENTATION = PRESENTATION_TABLE;

const HEADER_COUNT = "count";
const HEADER_DATE = "date";

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
      selectedPresentation: null,

      showDatePicker: false,
      topError: null
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

  getQuery(groupByDate=false) {
    let query = {};
    query.type = this.state.type.value;
    query.eventsCondition = 'all'; // Todo(Dinesh): Add a selector. Make it part of the query.

    if (this.state.resultDateRange.length == 0)
      throw new Error('Invalid date range. No default range given.')
    
    let from = null, to = null;
    let selRange = this.state.resultDateRange[0];
    if (selRange.label !== DEFAULT_DATE_RANGE_LABEL) {
      from = selRange.startDate;
      to = selRange.endDate;
    } else {
      // Resets the range based on current timestamp.
      from = moment(new Date()).subtract(7, 'days').toDate();
      to = new Date();
    }
    
    query.from = moment(from).unix(); // in utc.
    query.to = moment(to).unix(); // in utc.

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

    if (groupByDate) {
      query.groupByTimestamp = true;
    }
  
    console.debug(query);
    return query
  }

  isResponseValid(result) {
    if (result.error) {
      this.setState({ resultError: result.error });
      return false;
    }

    return true;
  }

  validateQuery() {
    if (this.state.events.length == 0) {
      return ERROR_NO_EVENT;
    }
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
  
    runQuery(this.props.currentProjectId, this.getQuery(presentation === PRESENTATION_LINE))
      .then((r) => {
        if(this.isResponseValid(r.data)) 
          this.setState({ result: r.data, selectedPresentation: presentation });
      })
      .catch(console.error);
  }

  getResultAsTable() {
    let result = this.state.result;
    let headers = result.headers.map((h, i) => { return <th key={'header_'+i}>{ h }</th> });
    let rows = [];

    for(let i=0; i<Object.keys(result.rows).length; i++) {
      let cols = result.rows[i.toString()];
      if (cols != undefined) {
        let tds = cols.map((c) => { return <td> { trimQuotes(c) } </td> });
        rows.push(<tr>{tds}</tr>);
      }
    }

    return (
      <Table className='fapp-table'> 
        <thead>
          <tr> { headers } </tr>
        </thead>
        <tbody>
          { rows }
        </tbody>
      </Table>
    );
  }

  getLinesByGroupsIfExist(rows, countIndex, dateIndex) {
    let lines = {}
    let keySep = " / ";

    for(let i=0; i<Object.keys(rows).length; i++) {
      let row = rows[i.toString()];
      if (row == undefined) continue;

      // All group properties joined together 
      // with a seperator is a key.
      let key = "";
      for(let c=0; c < row.length; c++) {
        if(c != countIndex && c != dateIndex) {
          let prop = trimQuotes(row[c]);
          if (key === "") {
            key = prop;
            continue;
          }
          key = key + keySep + prop;
        }
      }
      
      // init.
      if (!(key in lines)) {
        lines[key] = { counts: [], timestamps: [] }
      }
      
      lines[key].counts.push(row[countIndex]);
      lines[key].timestamps.push(moment(row[dateIndex]).format('MMM DD, YYYY'));
    }
    
    return lines;
  }

  getResultAsLineChart() {
    let result = this.state.result;
    let lines = [];

    let countIndex = result.headers.indexOf(HEADER_COUNT);
    if (countIndex == -1) { 
      throw new Error('No counts to plot as lines.');
    }
  
    let dateIndex = result.headers.indexOf(HEADER_DATE);
    if (dateIndex == -1) { 
      throw new Error('No dates to plot as lines.');
    }
      
    let pLines = this.getLinesByGroupsIfExist(result.rows, countIndex, dateIndex);
    for(let key in pLines) {
      let line = { title: key, xAxisLabels: pLines[key].timestamps, yAxisLabels: pLines[key].counts };
      lines.push(line);
    }
    
    return <div style={{height: '450px'}}> <LineChart lines={lines} /> </div>;
  }
  

  getResultAsVerticalBarChart() {
    let result = this.state.result;
    let bars = {};

    let countIndex = result.headers.indexOf(HEADER_COUNT);
    // Need a count and a group col for bar.
    if (countIndex == -1) { 
      throw new Error('Invalid query result for bar chart.');
    }
    
    let data = [], labels = [];
    if (result.headers.length == 2) {
      // Other col apart from count is group col.
      let groupIndex = countIndex == 0 ? 1 : 0;
      for(let i=0; i<Object.keys(result.rows).length; i++) {
        let cols = result.rows[i.toString()];
        if (cols != undefined && cols[countIndex] != undefined) {
          data.push(cols[countIndex]);
          labels.push(trimQuotes(cols[groupIndex]));
        }
      }
      bars.x_label = firstToUpperCase(result.headers[groupIndex]);
    } else if (result.headers.length == 1) {
      let col = result.rows["0"];
      data.push(col[countIndex]);
      bars.x_label = "";
    } else {
      throw new Error("Invalid no.of result columns for vertical bar.");
    }

    bars.datasets = [{ data: data  }];
    bars.labels = labels;
    bars.y_label = firstToUpperCase(result.headers[countIndex]);

    return <div style={{height: '450px'}}> <BarChart bars={bars} legend={false} /> </div>;
  }

  getResultAsTabularBarChart() {
    let result = this.state.result;
    return <TableBarChart data={result} />;
  }

  getPresentableResult() {
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
          remove={() => this.remove('events', i)}
          removeProperty={(propertyIndex) => this.removeEventProperty(i, propertyIndex)}
          // event handlers.
          onNameChange={(value) => this.onEventStateChange(value, i)} 
          // property handlers.
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
          remove={() => this.remove('groupBys', i)}
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
      <div className='fapp-content' style={{marginLeft: '2rem', marginRight: '2rem'}}>
        <div className='fapp-error' hidden={!this.state.topError}>
            <span>{ this.state.topError }</span>
        </div>

        {/* Query */}
        <div style={{ margin: '' }}>
          <Row style={{marginBottom: '15px'}}>
            <Col xs='12' md='12'>        
              <span style={{marginRight: '10px'}}> Get </span>
              <div style={{display: 'inline-block', width: '15%'}} className='fapp-select'>
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
              <div class='fapp-date-picker' hidden={!this.state.showDatePicker}>
                <DateRangePicker
                  ranges={this.state.resultDateRange}
                  onChange={this.handleResultDateRangeSelect}
                  staticRanges={ DEFINED_DATE_RANGES }
                  inputRanges={[]}
                  minDate={new Date('01 Jan 2000 00:00:00 GMT')} // range starts from given date.
                  maxDate={new Date()}
                />
                <button className='fapp-close-round-button' style={{float: 'right'}} onClick={this.toggleDatePickerDisplay}>x</button>
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
        
        {/* Presentation */}
        <div hidden={ this.state.result == null }>
          <Row>
            <Col xs='12' md='12'>
                <div class='pull-right'>
                  <div style={{ marginBottom: '10px'}}>
                    <button className={this.getPresentationSelectorClass(PRESENTATION_TABLE)} style={{marginRight: '10px', fontWeight: 500}} onClick={() => this.run(PRESENTATION_TABLE)}>Table</button>
                    <button className={this.getPresentationSelectorClass(PRESENTATION_BAR)}  style={{marginRight: '10px', fontWeight: 500}} onClick={() => this.run(PRESENTATION_BAR)}>Bar</button>
                    <button className={this.getPresentationSelectorClass(PRESENTATION_LINE)}  style={{marginRight: '10px', fontWeight: 500}} onClick={() => this.run(PRESENTATION_LINE)}>Line</button>
                  </div>
                </div>
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