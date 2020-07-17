import React, {Component} from 'react';
import Select from 'react-select';
import {
  Button,
  ButtonDropdown,
  Col,
  DropdownItem,
  DropdownMenu,
  DropdownToggle,
  Modal,
  ModalFooter,
  ModalHeader,
  Row
} from 'reactstrap';
import {connect} from 'react-redux';
import {bindActionCreators} from 'redux';
import 'react-date-range/dist/styles.css';
import 'react-date-range/dist/theme/default.css';

import {fetchProjectEvents, runAttributionQuery} from '../../actions/projectsActions';
import {
  DASHBOARD_TYPE_WEB_ANALYTICS,
  DEFAULT_DATE_RANGE,
  DEFINED_DATE_RANGES,
  getQueryPeriod,
  PRESENTATION_TABLE,
  QUERY_CLASS_ATTRIBUTION,
  readableDateRange,
  sameDay
} from '../Query/common';
import ClosableDateRangePicker from '../../common/ClosableDatePicker';
import {getReadableKeyFromSnakeKey, makeSelectOpt, makeSelectOpts, removeElementByIndex} from '../../util';
import TableChart from '../Query/TableChart';
import Loading from '../../loading';
import mt from "moment-timezone";
import moment from 'moment';
import {createDashboardUnit} from "../../actions/dashboardActions";

const SOURCE = "Source";
const CAMPAIGN = "Campaign";
const ATTRIBUTION_KEYS = [
  {label: SOURCE, value: SOURCE},
  {label: CAMPAIGN, value: CAMPAIGN}
];

const FIRST_TOUCH = "First_Touch";
const LAST_TOUCH = "Last_Touch";
const ATTRIBUTION_METHODOLOGY = [
  {value: FIRST_TOUCH, label: "First Touch"},
  {value: LAST_TOUCH, label: "Last Touch"}
];

const IMPRESSIONS = "Impressions";
const CLICKS = "Clicks";
const SPEND = "Spend";

const CAMPAIGN_METRICS = [IMPRESSIONS, CLICKS, SPEND];

const NONE_OPT = {label: 'None', value: 'none'};

const LABEL_STYLE = {marginRight: '10px', fontWeight: '600', color: '#777'};

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    dashboards: store.dashboards.dashboards,
    eventNames: store.projects.currentProjectEventNames,
  };
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({
    fetchProjectEvents,
    createDashboardUnit,
  }, dispatch)
}

class AttributionQuery extends Component {
  constructor(props) {
    super(props);

    this.state = {
      duringDateRange: [DEFAULT_DATE_RANGE],
      linkedEventNames: [],
      isPresentationLoading: false,
      present: false,
      resultMetricsBreakdown: null,
      resultMeta: null,

      converisonEventName: null,
      loopbackDays: "",
      attributionMethodology: NONE_OPT,
      attributionKey: NONE_OPT,

      showDashboardsList: false,
      showAddToDashboardModal: false,
      addToDashboardMessage: null,
      selectedDashboardId: null,
      eventNamesLoaded: false,
      eventNamesLoadError: null,

      timeZone: null
    }
  }

  componentWillMount() {
    this.props.fetchProjectEvents(this.props.currentProjectId)
      .then(() => {
        this.setState({eventNamesLoaded: true, timeZone: this.getCurrentTimeZone()});
      })
      .catch((r) => {
        this.setState({eventNamesLoaded: true, eventNamesLoadError: r.paylaod});
      });
  }

  getDisplayMetricsBreakdown(metricsBreakdown) {
    if (!metricsBreakdown) return;

    let result = {...metricsBreakdown};
    for (let i = 0; i < result.headers.length; i++)
      result.headers[i] = getReadableKeyFromSnakeKey(result.headers[i]);

    return result;
  }

  getCurrentTimeZone() {
    return mt.tz.guess();
  }

  isLoaded() {
    return this.state.eventNamesLoaded;
  }

  validateQuery() {
    if (this.state.converisonEventName == null || this.state.converisonEventName == "") {
      this.props.showError("No conversion event provided.")
      return false;
    }

    for (let i = 0; i < this.state.linkedEventNames.length; i++) {
      if (this.state.linkedEventNames[i] == "" || this.state.linkedEventNames[i] == null) {
        this.props.showError("Invalid linked funnel event provided.")
        return false
      }
    }

    if (this.state.attributionKey.value != SOURCE && this.state.attributionKey.value != CAMPAIGN) {
      this.props.showError("No attribution key provided.")
      return false
    }

    if (this.state.attributionMethodology.value != FIRST_TOUCH && this.state.attributionMethodology.value != LAST_TOUCH) {
      this.props.showError("No attribution methodology provided.")
      return false
    }

    return true;
  }

  getQuery = () => {
    let query = {};
    query.cl = QUERY_CLASS_ATTRIBUTION;
    query.cm = CAMPAIGN_METRICS;
    query.ce = this.state.converisonEventName.value;
    query.lfe = this.state.linkedEventNames;
    query.attribution_key = this.state.attributionKey.value;
    query.attribution_methodology = this.state.attributionMethodology.value;
    query.lbw = Number(this.state.loopbackDays) || 0;
    let period = getQueryPeriod(this.state.duringDateRange[0]);
    query.from = period.from;
    query.to = period.to;

    return query;
  }

  runQuery = () => {
    let valid = this.validateQuery();
    if (!valid) return
    // Enable add to dashboard here.
    this.setState({present: true})

    this.props.resetError()
    this.setState({isPresentationLoading: true});
    let query = this.getQuery();
    runAttributionQuery(this.props.currentProjectId, query)
      .then(r => {
        this.setState({
          result: r.data,
          resultMeta: r.data.result.meta,
          isResultLoading: false, isPresentationLoading: false,
          resultMetricsBreakdown: this.getDisplayMetricsBreakdown(r.data.result)
        });
      })
      .catch(err => {
        console.log("error occurred while running query: ", err);
      });
  }

  getReadableAttributionMetricValue(key, value, meta) {
    if (value == null || value == undefined) return 0;
    if (typeof (value) != "number") return value;

    let rValue = value;
    let isFloat = (value % 1) > 0;
    if (isFloat) rValue = value >= 1 ? value.toFixed(0) : value.toFixed(2);
    // no decimal points for value >= 1 and 2 decimal points < 1.
    if (meta && meta.currency && key.toLowerCase().indexOf('spend') > -1)
      rValue = rValue + ' ' + meta.currency;
    return rValue;
  }

  renderAttributionResultAsTable() {
    if (!this.state.resultMetricsBreakdown || !this.state.resultMetricsBreakdown.headers ||
      !this.state.resultMetricsBreakdown.rows) return;

    let resultMetricsBreakdown = {...this.state.resultMetricsBreakdown};
    for (let ri = 0; ri < resultMetricsBreakdown.rows.length; ri++) {
      for (let ci = 0; ci < resultMetricsBreakdown.rows[ri].length; ci++) {
        let key = resultMetricsBreakdown.headers[ci];
        resultMetricsBreakdown.rows[ri][ci] = this.getReadableAttributionMetricValue(key,
          resultMetricsBreakdown.rows[ri][ci], this.state.resultMeta);
      }
    }
    return <Col md={12} style={{marginTop: '50px'}}>
      <Row><Col md={12}><TableChart sort bigWidthUptoCols={1} queryResult={resultMetricsBreakdown}/></Col></Row>
    </Col>;
  }

  handleDuringDateRangeSelect = (range) => {
    range.selected.label = null; // set null on custom range.
    if (sameDay(range.selected.endDate, new Date()) && !sameDay(range.selected.startDate, new Date())) {
      return
    }
    this.setState({duringDateRange: [range.selected]});
  }

  closeDatePicker = () => {
    this.setState({showDatePicker: false});
  }

  toggleDatePickerDisplay = () => {
    this.setState((state) => ({showDatePicker: !state.showDatePicker}));
  }

  onEventStateChange(option, index) {
    this.setState((prevState) => {
      let state = {...prevState};
      state.events = [...prevState.linkedEventNames];
      state.events[index].name = option.value;
      return state;
    })
  }

  addEvent = () => {
    this.setState((prevState) => {
      let state = {...prevState};
      state.linkedEventNames = [...prevState.linkedEventNames];
      // init with default state for each event row.
      state.linkedEventNames.push(null);
      return state;
    });
  }

  onEventNameChange = (eventIndex, option) => {
    this.setState((prevState) => {
      let state = {...prevState};
      state.linkedEventNames[eventIndex] = option.value;
      return state;
    })
  }

  removeEvent = (eventIndex) => {
    this.setState(() => {
      let state = {...this.state};
      state.linkedEventNames = removeElementByIndex(state.linkedEventNames, eventIndex);
      return state;
    })
  }

  renderEvents() {
    let events = [...this.state.linkedEventNames];
    events = events.map((v, i) => {
      return (
        <div style={{marginBottom: "8px"}} key={"event_" + i}>
          <div style={{display: 'inline-block', width: '250px'}} className='fapp-select light'>
            <Select
              index={i}
              onChange={(value) => this.onEventNameChange(i, value)}
              options={makeSelectOpts(this.props.eventNames)}
              placeholder='Select an event'
              value={v != null ? makeSelectOpt(v) : null}
            />
          </div>
          <button className='fapp-close-button' onClick={() => this.removeEvent(i)}>x</button>
        </div>)
    });

    return events
  }

  handleMethodologyChange = (option) => {
    this.setState({attributionMethodology: option});
  }

  handleAttributionKeyChange = (option) => {
    this.setState({attributionKey: option});
  }

  handleConversionEventNameChange = (option) => {
    this.setState({converisonEventName: option});
  }

  handleLookbackWindowChange = (event) => {
    let days = event.target.value;
    if (Number(days) && days > 0) {
      this.setState({
        loopbackDays: days
      })
      return
    }

    if (days == 0) {
      this.setState({
        loopbackDays: ""
      })
    }
  }

  toggleDashboardsList = () => {
    this.setState((state) => ({showDashboardsList: !state.showDashboardsList}));
  }

  toggleAddToDashboardModal = () => {
    this.setState((state) => ({showAddToDashboardModal: !state.showAddToDashboardModal, addToDashboardMessage: null}));
  }

  selectDashboardToAdd = (event) => {
    let dashboardId = event.currentTarget.getAttribute('value');
    this.setState({selectedDashboardId: dashboardId})
    this.toggleAddToDashboardModal();
  }

  renderDashboardDropdownOptions() {
    let dashboardsDropdown = [];
    for (let i = 0; i < this.props.dashboards.length; i++) {
      let dashboard = this.props.dashboards[i];
      if (dashboard && dashboard.name !== DASHBOARD_TYPE_WEB_ANALYTICS) {
        dashboardsDropdown.push(
          <DropdownItem onClick={this.selectDashboardToAdd}
                        value={dashboard.id}>{dashboard.name}</DropdownItem>
        )
      }
    }
    return dashboardsDropdown;
  }

  renderAddToDashboardModal() {
    return (
      <Modal isOpen={this.state.showAddToDashboardModal} toggle={this.toggleAddToDashboardModal}
             style={{marginTop: '10rem'}}>

        <ModalHeader toggle={this.toggleAddToDashboardModal}>Confirm add to Dashboard</ModalHeader>

        <ModalFooter style={{borderTop: 'none', paddingBottom: '30px', paddingRight: '35px'}}>
          <Button outline color="success" onClick={this.addToDashboard}>Add</Button>
          <Button outline color='danger' onClick={this.toggleAddToDashboardModal}>Cancel</Button>
        </ModalFooter>
      </Modal>
    );
  }

  addToDashboard = () => {
    let queryUnit = {};
    queryUnit.cl = QUERY_CLASS_ATTRIBUTION;
    queryUnit.query = this.getQuery();

    let metricBreakdownQueryUnit = {...queryUnit};
    metricBreakdownQueryUnit.meta = {metrics_breakdown: true};

    let title = "Attribution - Metrics by " + queryUnit.query.attribution_key;
    let payload = {
      presentation: PRESENTATION_TABLE,
      query: metricBreakdownQueryUnit,
      title: title,
    };

    this.props.createDashboardUnit(this.props.currentProjectId,
      this.state.selectedDashboardId, payload)
      .catch(() => console.error("Failed adding to attribution metrics breakdown to dashboard."))

    // close modal.
    this.toggleAddToDashboardModal();
  }

  render() {
    if (!this.isLoaded()) return <Loading/>;
    return <div>
      <Row style={{marginBottom: "15px"}}>
        <Col xs='2' md='2' style={{paddingTop: "5px"}}>
          <span style={LABEL_STYLE}> Select Conversion Event</span>
        </Col>
        <Col xs='8' md='8'>
          <div className='fapp-select light' style={{display: 'inline-block', width: '250px'}}>
            <Select options={makeSelectOpts(this.props.eventNames)}
                    onChange={this.handleConversionEventNameChange}
                    placeholder='Select'/>
          </div>
        </Col>
      </Row>
      <Row style={{marginBottom: '15px'}}>
        <Col xs='2' md='2'>
          <Button outline style={{fontWeight: "bold", height: '38px', marginBottom: "8px"}} color='primary'
                  onClick={this.addEvent}>+ Linked Funnel Events</Button>
        </Col>
        <Col xs='8' md='8'>
          <Row style={{marginLeft: "0px"}}>{this.renderEvents()}
          </Row>
        </Col>
      </Row>
      <Row style={{marginBottom: "15px", marginTop: "-8px"}}>
        <Col xs='2' md='2' style={{paddingTop: "5px"}}>
          <span style={LABEL_STYLE}> Attribution Key</span>
        </Col>
        <Col xs='8' md='8'>
          <div className='fapp-select light' style={{display: 'inline-block', width: '250px'}}>
            <Select options={ATTRIBUTION_KEYS} onChange={this.handleAttributionKeyChange}
                    placeholder='Select'/>
          </div>
        </Col>
      </Row>

      <Row style={{marginBottom: "15px"}}>
        <Col xs='2' md='2' style={{paddingTop: "5px"}}>
          <span style={LABEL_STYLE}> Attribution Methodology</span>
        </Col>
        <Col xs='8' md='8'>
          <div className='fapp-select light' style={{display: 'inline-block', width: '250px'}}>
            <Select
              options={ATTRIBUTION_METHODOLOGY} onChange={this.handleMethodologyChange}
              placeholder='Select Event'/>
          </div>
        </Col>
      </Row>
      <Row style={{marginBottom: "15px"}}>
        <Col xs='2' md='2' style={{paddingTop: "5px"}}>
          <span style={LABEL_STYLE}>Lookback Window</span>
        </Col>
        <Col xs='8' md='8'>
          <input className="form-control" style={{
            height: "38px", width: "250px", borderRadius: "5px",
            border: "1px solid #bbb"
          }} type="text" value={this.state.loopbackDays} onChange={this.handleLookbackWindowChange}
                 placeholder="in days"/>
        </Col>
      </Row>

      <Row style={{marginBottom: '15px'}}>
        <Col xs='2' md='2' style={{paddingTop: "5px"}}>
          <span style={LABEL_STYLE}> Period </span>
        </Col>
        <Col xs="8" md="8">
          <Button outline style={{border: '1px solid #ccc', color: 'grey', marginRight: '10px'}}
                  onClick={this.toggleDatePickerDisplay}>
            <i className="fa fa-calendar" style={{marginRight: '10px'}}></i>
            {readableDateRange(this.state.duringDateRange[0])}
          </Button>

          <div className='fapp-date-picker' hidden={!this.state.showDatePicker}>
            <ClosableDateRangePicker
              ranges={this.state.duringDateRange}
              onChange={this.handleDuringDateRangeSelect}
              staticRanges={DEFINED_DATE_RANGES}
              inputRanges={[]}
              minDate={new Date('01 Jan 2000 00:00:00 GMT')} // range starts from given date.
              maxDate={moment(new Date()).subtract(1, 'days').endOf('day').toDate()}
              closeDatePicker={this.closeDatePicker}
            />
            <button className='fapp-close-round-button'
                    style={{float: 'right', marginLeft: '0px', borderLeft: 'none'}}
                    onClick={this.toggleDatePickerDisplay}>x
            </button>
          </div>
        </Col>
      </Row>
      <div style={{width: '100%', textAlign: 'center', marginTop: '15px'}}>
        <Button
          color='primary' style={{fontSize: '0.9rem', padding: '8px 18px', fontWeight: 500}}
          onClick={this.runQuery}
        > Run Query
        </Button>
      </div>

      <div hidden={!this.state.present} style={{
        borderTop: '1px solid rgb(221, 221, 221)',
        marginTop: '30px', marginLeft: '-60px', marginRight: '-60px'
      }}></div>

      {/*Dashboard*/}
      <div style={{paddingLeft: '30px', paddingRight: '30px', paddingTop: '10px', minHeight: '500px'}}>
        <Row style={{marginTop: '15px', marginRight: '10px'}} hidden={!this.state.present}>
          <Col xs='12' md='12'>
            <ButtonDropdown style={{float: 'right', marginRight: '-20px'}}
                            isOpen={this.state.showDashboardsList} toggle={this.toggleDashboardsList}>
              <DropdownToggle caret outline color="primary">
                Add to dashboard
              </DropdownToggle>
              <DropdownMenu style={{height: 'auto', maxHeight: '210px', overflowX: 'scroll'}} right>
                {this.renderDashboardDropdownOptions()}
              </DropdownMenu>
            </ButtonDropdown>
          </Col>
        </Row>

        {this.state.isPresentationLoading ? <Loading paddingTop='12%'/> : null}
        <div className='animated fadeIn' hidden={this.state.isPresentationLoading} style={{marginTop: "50px"}}>
          <Row> {this.renderAttributionResultAsTable()} </Row>
        </div>

        {this.renderAddToDashboardModal()}
      </div>

    </div>
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(AttributionQuery);