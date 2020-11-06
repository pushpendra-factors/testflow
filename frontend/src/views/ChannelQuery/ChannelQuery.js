import React, { Component } from 'react';
import Select from 'react-select';
import CreatableSelect from 'react-select/lib/Creatable';
import { Button, Row, Col, DropdownItem, ButtonDropdown, 
  DropdownToggle, DropdownMenu, Modal, ModalHeader, 
  ModalBody, ModalFooter, Input } from 'reactstrap';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import 'react-date-range/dist/styles.css';
import 'react-date-range/dist/theme/default.css'; 

import { runChannelQuery, fetchChannelFilterValues } from '../../actions/projectsActions';
import { createDashboardUnit } from '../../actions/dashboardActions'
import { DEFAULT_DATE_RANGE, DEFINED_DATE_RANGES, 
  readableDateRange, PRESENTATION_CARD, PRESENTATION_TABLE,
  QUERY_CLASS_CHANNEL, getQueryPeriod, DASHBOARD_TYPE_WEB_ANALYTICS} from '../Query/common';
import ClosableDateRangePicker from '../../common/ClosableDatePicker';
import { makeSelectOpts, getReadableKeyFromSnakeKey } from '../../util';
import TableChart from '../Query/TableChart';
import { getReadableChannelMetricValue } from './common';
import Loading from '../../loading';

const CHANNEL_METRIC_ORDER = [ "clicks", "impressions", "conversions", 
"conversion_rate", "total_cost", "cost_per_click", "cost_per_conversion" ]

const CHANNEL_GOOGLE_ADS = { label: 'Google Ads', value: 'google_ads' }
const CHANNEL_FACEBOOK_ADS = { label: "Facebook Ads", value: "facebook_ads"}
const CHANNEL_OPTS = [CHANNEL_GOOGLE_ADS, CHANNEL_FACEBOOK_ADS]

const FILTER_KEY_CAMPAIGN = { label: 'Campaigns', value: 'campaign' }
const FILTER_KEY_AD = { label: 'Ads', value: 'ad' }
const FILTER_KEY_AD_SET = {label: "Adsets", value: "adset"}
const FILTER_KEY_AD_GROUP = {label:"Adgroups", value: "ad_group"}
const FILTER_KEY_PLATFORM = {label: "Platforms", value: "platform"}
const FILTER_KEY_KEYWORD = { label: 'Keywords', value: 'keyword' }
const FILTER_KEY_OPTS = [ FILTER_KEY_CAMPAIGN, FILTER_KEY_AD, FILTER_KEY_AD_GROUP, FILTER_KEY_KEYWORD ];
const FACEBOOK_FILTER_KEY_OPTS = [FILTER_KEY_CAMPAIGN, FILTER_KEY_AD,FILTER_KEY_AD_SET]

// supported breakdown opts for each filter key.
const BREAKDOWN_KEY_OPT_MAP = {
  [FILTER_KEY_CAMPAIGN.value]: [FILTER_KEY_CAMPAIGN],
  [FILTER_KEY_AD.value]: [FILTER_KEY_AD],
  [FILTER_KEY_KEYWORD.value]: [FILTER_KEY_KEYWORD],
  [FILTER_KEY_AD_SET.value] : [FILTER_KEY_AD_SET],
  [FILTER_KEY_AD_GROUP.value] : [FILTER_KEY_AD_GROUP],
  [FILTER_KEY_PLATFORM.value] : [FILTER_KEY_PLATFORM]
}

const ALL_OPT = { label: 'All', value: 'all' }
const NONE_OPT = { label: 'None', value: 'none' }

const STATUS_OPTS = [ ALL_OPT ]
const MATCH_TYPE_OPTS = [ ALL_OPT ]

const LABEL_STYLE = { marginRight: '10px', fontWeight: '600', color: '#777' };

const mapStateToProps = store => {
  return { 
    currentProjectId: store.projects.currentProjectId,
    dashboards: store.dashboards.dashboards,
    channelFilterValues: store.projects.channelFilterValues,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({
    fetchChannelFilterValues,
    createDashboardUnit,
  }, dispatch)
}

class ChannelQuery extends Component {
  constructor(props) {
    super(props);

    this.state = {
      channel: CHANNEL_GOOGLE_ADS,
      filterKey: FILTER_KEY_CAMPAIGN,
      filterValue: ALL_OPT,
      duringDateRange: [DEFAULT_DATE_RANGE],
      isFilterValuesLoading: false,
      breakdownKey: NONE_OPT,

      isPresentationLoading: false,
      present: false,
      resultMetrics: {},
      resultMetricsBreakdown: null,
      resultMeta: null,
      topError: null,

      showDashboardsList: false,
      showAddToDashboardModal: false,
      addToDashboardMessage: null,
      selectedDashboardId: null,
      addToDashboardMetricUnits: [],
      addToDashboardMeticBreakdown: false,
    }
  }

  getDisplayMetricsBreakdown(metricsBreakdown) {
    if (!metricsBreakdown) return;

    let result = { ...metricsBreakdown };
    for (let i=0; i<result.headers.length; i++)
      result.headers[i] = getReadableKeyFromSnakeKey(result.headers[i]);

    return result;
  }

  getQuery = () => {
    let query = {};
    query.channel = this.state.channel.value;
    query.filter_key = this.state.filterKey.value;
    query.filter_value = this.state.filterValue.value;

    let period = getQueryPeriod(this.state.duringDateRange[0]);
    query.from = period.from;
    query.to = period.to;

    if (this.state.breakdownKey.value != "none")
      query.breakdown = this.state.breakdownKey.value;

    return query
  }

  runQuery = () => {
    this.setState({ isPresentationLoading: true });

    let query = this.getQuery();
    runChannelQuery(this.props.currentProjectId, query)
      .then((r) => {
        if (!r.ok) {
          this.setState({ topError: 'Failed to run query.' });
          return
        }

        if (r.data.meta)
          this.setState({ resultMeta: r.data.meta });

        if (r.data.metrics)
          this.setState({ present: true, resultMetrics: r.data.metrics });
        else
          console.error('No metrics on channel query response.');

        if (r.data.metrics_breakdown)
          this.setState({ present: true,
            resultMetricsBreakdown: this.getDisplayMetricsBreakdown(r.data.metrics_breakdown) });

        this.setState({ isPresentationLoading: false });
      });

    // reset the add to dashbaord units as result changes.
    this.setState({ addToDashboardMetricUnits: [] });
  }

  onSelectMetricUnitAddToDashboard = (k) => {
    let selectedUnits = [ ...this.state.addToDashboardMetricUnits ];

    if (!this.isMetricUnitAddedToDashboard(k)) {
      // add if not exist.
      selectedUnits.push(k);
    } else {
      // remove if key exists.
      selectedUnits.splice(selectedUnits.indexOf(k), 1)
    }
    
    this.setState({ addToDashboardMetricUnits: selectedUnits });
  }

  isMetricUnitAddedToDashboard(k) {
    return this.state.addToDashboardMetricUnits.indexOf(k) > -1;
  }

  getMetricWidget(k, isAddToDashboardModal) {
    return <Col md={3} style={{ padding: '0 15px', marginTop: '30px'}}>
      <div style={{ border: '1px solid #AAA' }}>

        { isAddToDashboardModal ? <div style={{ textAlign: "right", padding: "5px" }} >
          <Input onChange={() => this.onSelectMetricUnitAddToDashboard(k)} 
          checked={this.isMetricUnitAddedToDashboard(k)} type="checkbox"/></div> : null } 

        <div style={{ padding: '35px' }}>
          <span style={{display: 'block', textAlign: 'center', fontSize: '18px', marginBottom: '15px'}}> 
            { getReadableKeyFromSnakeKey(k) } 
          </span>
          <span style={{display: 'block', textAlign: 'center', fontSize: '20px', fontWeight: '500' }}> 
            { getReadableChannelMetricValue(k, this.state.resultMetrics[k], this.state.resultMeta) } 
          </span>
        </div>
      </div>
    </Col>;
  }

  presentMetrics(isAddToDashboardModal) {
    let widgets = [];
    let addedMetrics = [];

    // add metrics by order pref.
    for (let i=0; i<CHANNEL_METRIC_ORDER.length; i++) {     
      if (this.state.resultMetrics.hasOwnProperty(CHANNEL_METRIC_ORDER[i])) { 
        widgets.push(this.getMetricWidget(CHANNEL_METRIC_ORDER[i], isAddToDashboardModal));
        addedMetrics.push(CHANNEL_METRIC_ORDER[i]);
      }
    }

    // add metrics without order pref at the end. 
    for (let k in this.state.resultMetrics) {
      if (addedMetrics.indexOf(k) == -1) {
        widgets.push(this.getMetricWidget(k, isAddToDashboardModal));
      }
    }

    return widgets;
  }

  onSelectMetricBreakdownAddToDashboard = () => {
    this.setState({ addToDashboardMeticBreakdown: !this.state.addToDashboardMeticBreakdown });
  }

  presentMetricsBreakdown(isAddToDashboardModal) {
    if (!this.state.resultMetricsBreakdown ||  !this.state.resultMetricsBreakdown.headers ||
      !this.state.resultMetricsBreakdown.rows || this.state.breakdownKey === NONE_OPT) return;

    let resultMetricsBreakdown = { ...this.state.resultMetricsBreakdown };
    for (let ri=0; ri < resultMetricsBreakdown.rows.length; ri++ ) {
      for (let ci=0; ci < resultMetricsBreakdown.rows[ri].length; ci++) {
        let key = resultMetricsBreakdown.headers[ci];
        resultMetricsBreakdown.rows[ri][ci] = getReadableChannelMetricValue(key, 
          resultMetricsBreakdown.rows[ri][ci], this.state.resultMeta);
      }
    }

    return <Col md={12} style={{ marginTop: '50px' }}>
      { 
        isAddToDashboardModal ? <Row >
          <Col md={12}>
            <div style={{ background: '#EBEDFD', width: '100%', height:'25px', 
              textAlign: 'right', paddingRight: '5px', paddingTop: '5px' }}>
              <Input onChange={this.onSelectMetricBreakdownAddToDashboard} 
                checked={this.state.addToDashboardMeticBreakdown} type='checkbox' />
            </div>
          </Col>
        </Row> : null
      }
      <Row><Col md={12}><TableChart sort bigWidthUptoCols={1} queryResult={resultMetricsBreakdown} /></Col></Row>
    </Col>;
  }
  handleChannelChange = (option) => {
    console.log(option)
    this.setState({channel: option, })
  }

  handleFilterKeyChange = (option) => {
    this.setState({ filterKey: option, filterValue: ALL_OPT });
  }

  handleBreakdownKeyChange = (option) => {
    this.setState({ breakdownKey: option });
  }

  getBreakdownKeysOpts(filterKey) {
    let opts
    if(this.state.channel.value == "google_ads")
    {
      opts = [NONE_OPT]
    }
    else {
      opts = [ NONE_OPT, FILTER_KEY_PLATFORM ];
    }
    if (!filterKey || filterKey == "") return opts;
    if (!BREAKDOWN_KEY_OPT_MAP.hasOwnProperty(filterKey)) {
      console.error("No breakdown key opts for selected filter key.");
      return opts;
    }
    
    return [...opts, ...BREAKDOWN_KEY_OPT_MAP[filterKey]];
  }

  handleDuringDateRangeSelect = (range) => {
    range.selected.label = null; // set null on custom range.
    this.setState({ duringDateRange: [range.selected] });
  }

  closeDatePicker = () => {
    this.setState({ showDatePicker: false }); 
  }

  toggleDatePickerDisplay = () => {
    this.setState({ showDatePicker: !this.state.showDatePicker });
  }

  isChannelFilterValuesExists() {
    return this.props.channelFilterValues[this.state.channel.value] &&
    this.props.channelFilterValues[this.state.channel.value][this.state.filterKey.value];
  }

  getChannelFilterValues = () => {
    // Do not fetch from remote if exists on store.
    if (this.isChannelFilterValuesExists()) return;

    this.setState({ isFilterValuesLoading: true });
    this.props.fetchChannelFilterValues(this.props.currentProjectId, 
      this.state.channel.value, this.state.filterKey.value)
      .then(() => { this.setState({ isFilterValuesLoading: false }); });
  }

  getChannelFilterValuesOpts() {
    if (!this.isChannelFilterValuesExists()) return [ALL_OPT];
    let valueOpts = makeSelectOpts(
      this.props.channelFilterValues[this.state.channel.value][this.state.filterKey.value]);
    valueOpts.unshift(ALL_OPT);
    return valueOpts;
  }

  onChannelFilterValueChange = (value) => {
    this.setState({ filterValue: value });
  }

  renderDashboardDropdownOptions() {
    let dashboardsDropdown = [];
    for(let i=0; i<this.props.dashboards.length; i++){
      let dashboard = this.props.dashboards[i];
      if (dashboard && dashboard.name != DASHBOARD_TYPE_WEB_ANALYTICS ) {
        dashboardsDropdown.push(
          <DropdownItem onClick={this.selectDashboardToAdd} 
            value={dashboard.id}>{dashboard.name}</DropdownItem>
        )
      }
    }
    
    return dashboardsDropdown;
  }

  toggleDashboardsList = () => {
    this.setState({ showDashboardsList: !this.state.showDashboardsList });
  }

  toggleAddToDashboardModal = () =>  {
    this.setState({ showAddToDashboardModal: !this.state.showAddToDashboardModal, addToDashboardMessage: null });
  }

  selectDashboardToAdd = (event) => {
    let dashboardId = event.currentTarget.getAttribute('value');
    this.setState({ selectedDashboardId: dashboardId })
    this.toggleAddToDashboardModal();
  }

  renderAddToDashboardModal() {
    return (
      <Modal isOpen={this.state.showAddToDashboardModal} toggle={this.toggleAddToDashboardModal} 
        style={{ marginTop: "3rem", minWidth: "80rem" }}>
        <ModalHeader toggle={this.toggleAddToDashboardModal}>Add to Dashboard</ModalHeader>
        <ModalBody style={{padding: '25px 35px'}}>
          <Row> { this.presentMetrics(true) } </Row>
          <Row> { this.presentMetricsBreakdown(true) } </Row>
        </ModalBody>
        <ModalFooter style={{borderTop: 'none', paddingBottom: '30px', paddingRight: '35px'}}>
          <Button outline color="success" onClick={this.addToDashboard}>Add</Button>
          <Button outline color='danger' onClick={this.toggleAddToDashboardModal}>Cancel</Button>
        </ModalFooter>
      </Modal>
    );
  }

  addToDashboard = () => {
    let queryUnit = {};
    queryUnit.cl = QUERY_CLASS_CHANNEL;
    queryUnit.query = this.getQuery();;

    // add individual dashboard unit for each selected key.
    for (let i=0; i < this.state.addToDashboardMetricUnits.length; i++) {
      let metricQueryUnit = { ...queryUnit };
      metricQueryUnit.meta = { metric: this.state.addToDashboardMetricUnits[i] };


      let title = getReadableKeyFromSnakeKey(this.state.addToDashboardMetricUnits[i]);
      let payload = {
        presentation: PRESENTATION_CARD,
        query: metricQueryUnit,
        title: title,
      };

      this.props.createDashboardUnit(this.props.currentProjectId, 
        this.state.selectedDashboardId, payload)
        .catch(() => console.error("Failed adding to channel metric to dashboard."))
    }

    // add metric breakdown to dashboard only if selected.
    if (this.state.addToDashboardMeticBreakdown) {
      let metricBreakdownQueryUnit = { ...queryUnit };
      metricBreakdownQueryUnit.meta = { metrics_breakdown: true };

      let title = "Google Ads: Metrics by " + queryUnit.query.breakdown;
      let payload = {
        presentation: PRESENTATION_TABLE,
        query: metricBreakdownQueryUnit,
        title: title, 
      };

      this.props.createDashboardUnit(this.props.currentProjectId, 
        this.state.selectedDashboardId, payload)
        .catch(() => console.error("Failed adding to channel metrics breakdown to dashboard."))
    }

    // close modal.
    this.toggleAddToDashboardModal();
  }

  render() {
    return <div>
        <Row style={{ marginBottom: "15px" }}>
          <Col xs='2' md='2' style={{ paddingTop: "5px" }}>
            <span style={LABEL_STYLE}>Channel</span>
          </Col>
          <Col xs='10' md='10' style={{ marginLeft: "-7%" }}>
            <div className='fapp-select light' style={{ display: 'inline-block', width: '150px' }}>
              <Select value={this.state.channel} options={CHANNEL_OPTS} onChange={this.handleChannelChange}
              placeholder='Channel'/>
            </div>
          </Col>
        </Row>

        <Row style={{ marginBottom: "15px" }}>
          <Col xs='2' md='2' style={{ paddingTop: "5px" }}>
            <span style={LABEL_STYLE}>Filter by</span>
          </Col>
          <Col xs='10' md='10' style={{ marginLeft: "-7%" }}>
            <div className='fapp-select light' style={{ display: 'inline-block', width: '200px', marginRight: '15px' }}>
              <Select value={this.state.filterKey} onChange={this.handleFilterKeyChange} 
                options={this.state.channel.value == "google_ads"? FILTER_KEY_OPTS: FACEBOOK_FILTER_KEY_OPTS} placeholder='Filter'/>
            </div>
            <div className='fapp-select light' style={{ display: 'inline-block', width: '275px' }}>
              <CreatableSelect 
                value={this.state.filterValue} 
                options={this.getChannelFilterValuesOpts()}
                placeholder='Filter Value'
                onChange={this.onChannelFilterValueChange}
                onFocus={this.getChannelFilterValues}
                isLoading={this.isFilterValuesLoading}
              />
            </div>
          </Col>
        </Row>

        <Row style={{ marginBottom: "15px" }}>
          <Col xs='2' md='2' style={{ paddingTop: "5px" }}>
            <span style={LABEL_STYLE}> During </span>
          </Col>
          <Col xs='10' md='10' style={{ marginLeft: "-7%" }}>
            <Button outline style={{ border: '1px solid #ccc', color: 'grey', marginRight: '10px' }} 
              onClick={this.toggleDatePickerDisplay}>
              <i className="fa fa-calendar" style={{marginRight: '10px'}}></i>
              { readableDateRange(this.state.duringDateRange[0]) } 
            </Button>
            <div className='fapp-date-picker' hidden={!this.state.showDatePicker}>
              <ClosableDateRangePicker
                ranges={this.state.duringDateRange}
                onChange={this.handleDuringDateRangeSelect}
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

        <Row style={{ marginBottom: "15px" }}>
          <Col xs='2' md='2' style={{ paddingTop: "5px" }}>
            <span style={LABEL_STYLE}>Breakdown by</span>
          </Col>
          <Col xs='10' md='10' style={{ marginLeft: "-7%" }}>
            <div className='fapp-select light' style={{ display: 'inline-block', width: '200px', marginRight: '15px' }}>
              <Select value={this.state.breakdownKey} onChange={this.handleBreakdownKeyChange} options={this.getBreakdownKeysOpts(this.state.filterKey.value)} placeholder='Breakdown'/>
            </div>
          </Col>
        </Row>

      <div style={{ width: '100%', textAlign: 'center', marginTop: '15px' }}>
        <Button 
          color='primary' style={{ fontSize: '0.9rem', padding: '8px 18px', fontWeight: 500 }}
          onClick = {this.runQuery}
        > Run Query 
        </Button>
      </div>

      <div hidden={!this.state.present} style={{borderTop: '1px solid rgb(221, 221, 221)', 
        marginTop: '30px', marginLeft: '-60px', marginRight: '-60px'}}></div>

      {/* presentation */}
      <div style={{ paddingLeft: '30px', paddingRight: '30px', paddingTop: '10px', minHeight: '500px' }}>
        <Row style={{ marginTop: '15px', marginRight: '10px' }} hidden={ !this.state.present }>
          <Col xs='12' md='12'>
            <ButtonDropdown style={{ float: 'right', marginRight: '-20px' }} 
              isOpen={this.state.showDashboardsList} toggle={this.toggleDashboardsList}> 
              <DropdownToggle caret outline color="primary">
                Add to dashboard
              </DropdownToggle>
              <DropdownMenu style={{ height: 'auto', maxHeight: '210px', overflowX: 'scroll' }} right>
                { this.renderDashboardDropdownOptions() }
              </DropdownMenu>
            </ButtonDropdown>
          </Col>
        </Row>

        { this.state.isPresentationLoading ? <Loading paddingTop='12%' /> : null }
        <div className='animated fadeIn' hidden={this.state.isPresentationLoading}>
          <Row> { this.presentMetrics() } </Row>
          <Row> { this.presentMetricsBreakdown() } </Row>
        </div>

        { this.renderAddToDashboardModal() }
      </div>

    </div>
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(ChannelQuery);