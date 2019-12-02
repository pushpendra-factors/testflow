import React, { Component } from 'react';
import Select from 'react-select';
import CreatableSelect from 'react-select/lib/Creatable';
import { Button, Row, Col } from 'reactstrap';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import 'react-date-range/dist/styles.css';
import 'react-date-range/dist/theme/default.css';
import moment from 'moment';

import { runChannelQuery, fetchChannelFilterValues } from '../../actions/projectsActions';
import { DEFAULT_DATE_RANGE, DEFINED_DATE_RANGES, 
  readableDateRange } from '../Query/common';
import ClosableDateRangePicker from '../../common/ClosableDatePicker';
import { makeSelectOpts } from '../../util';
import TableChart from '../Query/TableChart';

const CHANNEL_GOOGLE_ADS = { label: 'Google Ads', value: 'google_ads' }
const CHANNEL_OPTS = [CHANNEL_GOOGLE_ADS]

const FILTER_KEY_CAMPAIGN = { label: 'Campaigns', value: 'campaign' }
const FILTER_KEY_AD = { label: 'Ads', value: 'ad' }
const FILTER_KEY_KEYWORD = { label: 'Keywords', value: 'keyword' }
const FILTER_KEY_OPTS = [ FILTER_KEY_CAMPAIGN, FILTER_KEY_AD, FILTER_KEY_KEYWORD ]

const ALL_OPT = { label: 'All', value: 'all' }
const NONE_OPT = { label: 'None', value: 'none' }

const STATUS_OPTS = [ ALL_OPT ]
const MATCH_TYPE_OPTS = [ ALL_OPT ]

const LABEL_STYLE = { marginRight: '10px', fontWeight: '600', color: '#777' };

const mapStateToProps = store => {
  return { 
    currentProjectId: store.projects.currentProjectId,
    channelFilterValues: store.projects.channelFilterValues,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({
    fetchChannelFilterValues,
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

      present: false,
      resultMetrics: {},
      resultMetricsBreakdown: null,
      topError: null,
    }
  }
  // Returns: 20191026
  getDateOnlyTimestamp(datetime) {
    return parseInt(moment(datetime).format('YYYYMMDD'));
  }

  getDisplayMetricsBreakdown(metricsBreakdown) {
    if (!metricsBreakdown) return;

    let result = { ...metricsBreakdown };
    for (let i=0; i<result.headers.length; i++)
      result.headers[i] = this.getSnakeToReadableKey(result.headers[i]);

    return result;
  }

  runQuery = () => {
    let query = {};
    query.channel = this.state.channel.value;
    query.filter_key = this.state.filterKey.value;
    query.filter_value = this.state.filterValue.value;
    query.date_from = this.getDateOnlyTimestamp(this.state.duringDateRange[0].startDate);
    query.date_to = this.getDateOnlyTimestamp(this.state.duringDateRange[0].endDate);

    if (this.state.breakdownKey.value != "none") 
      query.breakdown = this.state.breakdownKey.value;

    runChannelQuery(this.props.currentProjectId, query)
      .then((r) => {
        if (!r.ok) {
          this.setState({ topError: 'Failed to run query.' });
          return
        }

        if (r.data.metrics)
          this.setState({ present: true, resultMetrics: r.data.metrics });

        if (r.data.metrics_breakdown)
          this.setState({ present: true, 
            resultMetricsBreakdown: this.getDisplayMetricsBreakdown(r.data.metrics_breakdown) });
      });
  }

  getSnakeToReadableKey(k) { 
    let kSplits = k.split('_');

    let key = '';
    for (let i=0; i<kSplits.length; i++)
      key = key + ' ' + kSplits[i].charAt(0).toUpperCase() + kSplits[i].slice(1);
    
    return key
  }

  presentMetrics() {
    let widgets = [];

    for (let k in this.state.resultMetrics) {
      let value = (this.state.resultMetrics[k] == null || this.state.resultMetrics[k] == undefined) ? 
        'NA' : this.state.resultMetrics[k];
      
      widgets.push(
        <Col md={3} style={{ padding: '0 15px', marginTop: '30px'}}>
          <div style={{ border: '1px solid #AAA', padding: '35px' }}>
            <span style={{display: 'block', textAlign: 'center', fontSize: '18px', marginBottom: '15px'}}> 
              { this.getSnakeToReadableKey(k) } 
            </span>
            <span style={{display: 'block', textAlign: 'center', fontSize: '20px', fontWeight: '500' }}> 
              { value } 
            </span>
          </div>
        </Col>
      );
    }

    return widgets;
  }

  presentMetricsBreakdown() {
    if (!this.state.resultMetricsBreakdown || !this.state.resultMetricsBreakdown.rows) return;
    if (this.state.resultMetricsBreakdown.rows.length <= 1) return;

    return <Col md={12} style={{ marginTop: '50px' }}>
      <TableChart queryResult={this.state.resultMetricsBreakdown} />
    </Col>;
  }

  handleFilterKeyChange = (option) => {
    this.setState({ filterKey: option, filterValue: ALL_OPT });
  }

  handleBreakdownKeyChange = (option) => {
    this.setState({ breakdownKey: option });
  }

  getBreakdownKeysOpts() {
    return [NONE_OPT, ...FILTER_KEY_OPTS];
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
    let valueOpts = makeSelectOpts(this.props.channelFilterValues[this.state.channel.value][this.state.filterKey.value]);
    valueOpts.unshift(ALL_OPT);
    return valueOpts;
  }

  onChannelFilterValueChange = (value) => {
    this.setState({ filterValue: value });
  }

  render() {
    return <div>
      <Row style={{marginBottom: '15px'}}>
        <Col xs='12' md='12'>
          <span style={LABEL_STYLE}>Channel</span>
          <div className='fapp-select light' style={{ display: 'inline-block', width: '150px' }}>
            <Select value={this.state.channel} options={CHANNEL_OPTS} placeholder='Channel'/>
          </div>
        </Col>
      </Row>

      <Row style={{marginBottom: '15px'}}>
        <Col xs='12' md='12'>
          <span style={LABEL_STYLE}>Filter by</span>
          <div className='fapp-select light' style={{ display: 'inline-block', width: '200px', marginRight: '15px' }}>
            <Select value={this.state.filterKey} onChange={this.handleFilterKeyChange} options={FILTER_KEY_OPTS} placeholder='Filter'/>
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

      {/* 
      <Row style={{marginBottom: '15px'}}>
        <Col xs='12' md='12'>
          <span style={LABEL_STYLE}>Status</span>
          <div className='fapp-select light' style={{ display: 'inline-block', width: '150px' }}>
            <Select value={ALL_OPT} options={STATUS_OPTS} placeholder='Status'/>
          </div>
        </Col>
      </Row> 
      */}

      {/* 
      <Row style={{marginBottom: '15px'}}>
        <Col xs='12' md='12'>
          <span style={LABEL_STYLE}>Match Type</span>
          <div className='fapp-select light' style={{ display: 'inline-block', width: '150px' }}>
            <Select value={ALL_OPT} options={MATCH_TYPE_OPTS} placeholder='Match Type'/>
          </div>
        </Col>
      </Row> 
      */}

      <Row style={{marginBottom: '15px'}}>
        <Col xs='12' md='12'>
          <span style={LABEL_STYLE}> During </span>
          <Button outline style={{border: '1px solid #ccc', color: 'grey', marginRight: '10px' }} 
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

      <Row style={{marginBottom: '15px'}}>
        <Col xs='12' md='12'>
          <span style={LABEL_STYLE}>Breakdown by</span>
          <div className='fapp-select light' style={{ display: 'inline-block', width: '200px', marginRight: '15px' }}>
            <Select value={this.state.breakdownKey} onChange={this.handleBreakdownKeyChange} options={this.getBreakdownKeysOpts()} placeholder='Breakdown'/>
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

      <div hidden={!this.state.present} style={{borderTop: '1px solid rgb(221, 221, 221)', paddingTop: '20px', 
        marginTop: '30px', marginLeft: '-60px', marginRight: '-60px'}}></div>

      {/* presentation */}
      <div style={{ paddingLeft: '30px', paddingRight: '30px', paddingTop: '10px', minHeight: '500px' }}>
        <Row> { this.presentMetrics() } </Row>
        <Row> { this.presentMetricsBreakdown() } </Row>
      </div>

    </div>
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(ChannelQuery);