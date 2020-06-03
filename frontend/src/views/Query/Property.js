import React, { Component } from 'react';
import Select from 'react-select';
import CreatableSelect from 'react-select/lib/Creatable';
import { Row, Col, Input, Button } from 'reactstrap';
import moment from 'moment';

import ClosableDateRangePicker from '../../common/ClosableDatePicker';
import { 
  fetchProjectEventProperties,
  fetchProjectEventPropertyValues,
  fetchProjectUserProperties,
  fetchProjectUserPropertyValues,
} from '../../actions/projectsActions';
import { makeSelectOpts, createSelectOpts, getSelectedOpt, 
  makeSelectOpt, QUERY_TYPE_ANALYTICS } from "../../util";
import { PROPERTY_VALUE_NONE, PROPERTY_TYPE_OPTS, PROPERTY_LOGICAL_OP_OPTS,
  getDateRangeFromStoredDateRange, DEFAULT_DATE_RANGE, DEFINED_DATE_RANGES 
} from "./common";

const TYPE_NUMERICAL = 'numerical';
const TYPE_CATEGORICAL = 'categorical';
const TYPE_DATETIME = 'datetime';

const OP_CONTAINS = 'contains';
const OP_NOT_CONTAINS = 'notContains';

const LABEL_STYLE = { marginRight: '10px', fontWeight: '600', color: '#777' };

const NUMERICAL_OPERATOR_OPTS = { 
  'equals': '=',
  'notEqual': '!=',
  'lesserThan': '<' ,
  'lesserThanOrEqual': '<=',
  'greaterThan': '>',
  'greaterThanOrEqual': '>=',
};

const CATEGORICAL_OPERATORS_OPTS = {
  'equals': '=',
  'notEqual': '!=',
  'contains': 'contains',
  'notContains': 'not contains',
}

class Property extends Component {
  constructor(props) {
    super(props);

    this.state = {
      nameOpts: [],
      valueOpts: [],
      valueType: null,
      isNameOptsLoading: false,
      isValueOptsLoading: false,

      showDatePicker: false,
    }
  }

  addToNameOptsState(props) {
    let opts = [];
    for(let type in props) {
      for(let i in props[type]) {
        let v = props[type][i];
        // type: categorical, numerical.
        opts.push({value: v, label: v, type: type}); 
      }
    }
    opts.sort((a, b)=>a.label > b.label ? 1:-1 );
    this.setState({ nameOpts: opts, isNameOptsLoading: false });
  }

  addToValueOptsState(values) {
    let valuesOpts = makeSelectOpts(values);
    valuesOpts.unshift(makeSelectOpt(PROPERTY_VALUE_NONE));

    if (values != undefined && values != null)
      this.setState({ valueOpts: valuesOpts, isValueOptsLoading: false });
  }

  fetchPropertyKeys = () => {
    this.setState({ nameOpts: [], isNameOptsLoading: true }); // reset opts.

    if (this.props.propertyState.entity == 'event') {
      fetchProjectEventProperties(this.props.projectId, this.props.eventName, "", false)
        .then((r) => this.addToNameOptsState(r.data))
        .catch(r => console.error("Failed fetching event property keys.", r));
    }

    if (this.props.propertyState.entity == 'user') {
      fetchProjectUserProperties(this.props.projectId, QUERY_TYPE_ANALYTICS, "", false)
      .then((r) => this.addToNameOptsState(r.data))
      .catch(r => console.error("Failed fetching user property keys.", r));
    }
  }

  fetchPropertyValues = () => { 
    this.setState({ valueOpts: [], isValueOptsLoading: true }); // reset opts.

    if (this.props.propertyState.entity == 'event') {
      fetchProjectEventPropertyValues(this.props.projectId, 
        this.props.eventName, this.props.propertyState.name, false)
        .then(r => this.addToValueOptsState(r.data))
        .catch(r => console.error("Failed fetching event property values.", r));
    }
    
    if (this.props.propertyState.entity == 'user') {
      fetchProjectUserPropertyValues(this.props.projectId, 
        this.props.propertyState.name, false)
        .then(r => this.addToValueOptsState(r.data))
        .catch(r => console.error("Failed fetching user property values.", r));
    }
  }

  onNameChange = (option) => {
    if (option.type != null && 
      option.type != TYPE_NUMERICAL && 
      option.type != TYPE_CATEGORICAL && 
      option.type != TYPE_DATETIME) {
      
      throw new Error('Unknown property value type.');
    }
    this.setState({ valueType: option.type });
    this.props.onNameChange(option.value);
  }

  getDateRangeAsStr(range, overridePeriod=false) {
    return (
      JSON.stringify({
        fr: moment(range.startDate).unix(), 
        to: moment(range.endDate).unix(),
        ovp: overridePeriod,
      })
    );
  }

  getDateRangeFromStr(rangeStr) {
    let dateRange = JSON.parse(rangeStr);
    return getDateRangeFromStoredDateRange(dateRange);
  }

  onValueChange = (v) => {
    if (this.state.valueType == TYPE_DATETIME) {
      let isEndDateToday = moment(v.selected.endDate).isSame(moment(), 'day');
      this.props.onValueChange(this.getDateRangeAsStr(v.selected, isEndDateToday), this.state.valueType);
      return
    }

    if (this.state.valueType == TYPE_NUMERICAL) {
      this.props.onValueChange(v.target.value.trim(), this.state.valueType);
    }
    
    if (this.state.valueType == TYPE_CATEGORICAL) {
      if (v.target == undefined) {
        this.props.onValueChange(v.value, this.state.valueType);
      }
      else {
        this.props.onValueChange(v.target.value.trim(), this.state.valueType);
      }
    }
  }

  toggleDatePickerDisplay = () => {
    this.setState({ showDatePicker: !this.state.showDatePicker });
  }

  closeDatePicker = () => {
    this.setState({ showDatePicker: false });
  }

  readableDateRange(range) {
    // Use label for default date range.
    if(range.startDate ==  DEFAULT_DATE_RANGE.startDate 
      && range.endDate == DEFAULT_DATE_RANGE.endDate)
      return DEFAULT_DATE_RANGE.label;

    return moment(range.startDate).format('MMM DD, YYYY') + " - " +
      moment(range.endDate).format('MMM DD, YYYY');
  }

  getDateRangeFromPropertyState() {
    if (this.props.propertyState.value == '') {
      // set default property state.
      this.props.onValueChange(this.getDateRangeAsStr(DEFAULT_DATE_RANGE, true), this.state.valueType);
      return [DEFAULT_DATE_RANGE];
    }
    
    return this.getDateRangeFromStr(this.props.propertyState.value);  
  }

  getInputValueElement() {
    let input = null;

    if (this.state.valueType == TYPE_DATETIME) {
      return (
        <div style={{display: "inline-block", marginLeft: "10px"}}>
          <Button outline style={{ border: '1px solid #ccc', color: 'grey', padding: '8px 12px', marginBottom: '3px' }} onClick={this.toggleDatePickerDisplay}>
            <i className="fa fa-calendar" style={{marginRight: '10px'}}></i>
            {this.readableDateRange(this.getDateRangeFromPropertyState()[0])}
          </Button>
          <div className='fapp-date-picker' style={{ display: 'block', marginTop: '10px' }} hidden={!this.state.showDatePicker}>
            <ClosableDateRangePicker
              ranges={this.getDateRangeFromPropertyState()}
              onChange={this.onValueChange}
              staticRanges={ DEFINED_DATE_RANGES }
              inputRanges={[]}
              minDate={new Date('01 Jan 2000 00:00:00 GMT')} // range starts from given date.
              maxDate={new Date()}
              closeDatePicker={this.closeDatePicker}
            />
            <button className='fapp-close-round-button' style={{float: 'right', marginLeft: '0px', borderLeft: 'none'}} onClick={this.toggleDatePickerDisplay}>x</button>
          </div>
        </div>
      );
    }

    if (this.state.valueType == TYPE_NUMERICAL) {
      return (
        <div style={{display: "inline-block", width: "240px", marginLeft: "10px"}}>
          <Input
            type="text"
            onChange={this.onValueChange}
            placeholder="Enter a value"
            value={this.props.propertyState.value}
            style={{ border: "1px solid #ddd", color: "#444444" }}
          />
        </div>
      );
    }
    
    if (this.state.valueType == TYPE_CATEGORICAL) {
      if (!(this.props.propertyState.op == OP_CONTAINS || this.props.propertyState.op == OP_NOT_CONTAINS)) {
        return  (
          <div style={{display: "inline-block", width: "240px", marginLeft: "10px"}} className='fapp-select light'>
            <CreatableSelect
              onChange={this.onValueChange}
              onFocus={this.fetchPropertyValues}
              options={this.state.valueOpts}
              value={getSelectedOpt(this.props.propertyState.value)}
              placeholder="Select a value"
              formatCreateLabel={(value) => (value)}
              isLoading={this.state.isValueOptsLoading}
            />
          </div>
        );
      } 
      else {
        return  (
          <div style={{display: "inline-block", width: "240px", marginLeft: "10px"}} className='fapp-select light'>
            <Input
              type="text"
              onChange={this.onValueChange}
              placeholder="Enter a value"
              value={this.props.propertyState.value}
              style={{ border: "1px solid #ddd", color: "#444444" }}
            />
          </div>
        );
      }
    }

    if (this.state.valueType != null) {
      throw new Error('Failed to get input element. Unknown property value type.');
    }
  }

  isOpRequired() {
    // Op not required for DATETIME.
    return this.state.valueType != TYPE_DATETIME;
  }

  getOpSelector() {
    if (!this.isOpRequired()) return;
    if (this.props.propertyState.valueType == '' && this.state.valueType == null) return;

    // state updated for view query. where valueType updated
    // without using the property component.
    if (this.state.valueType == null) {
      this.setState({ valueType: this.props.propertyState.valueType });
    }

    // categorical_operator_opts as default.
    let optSrc = this.state.valueType == TYPE_NUMERICAL ? NUMERICAL_OPERATOR_OPTS : CATEGORICAL_OPERATORS_OPTS;

    return (
      <div style={{display: "inline-block", width: "130px", marginLeft: "10px"}} className='fapp-select light'>
        <Select 
          onChange={this.props.onOpChange}
          options={createSelectOpts(optSrc)}
          value={getSelectedOpt(this.props.propertyState.op, optSrc)}
        />
      </div>
    );
  }

  isCondRequired() {
    return this.props.index > 0
  }

  render() {
    return <Row style={{ marginBottom: "15px" }}>
      <Col xs='12' md='12'>
        {
          !this.isCondRequired() ? <span style={LABEL_STYLE}> with </span> : (
            <div style={{display: "inline-block", width: "75px"}} className='fapp-select light'>
              <Select
                onChange={this.props.onLogicalOpChange}
                options={createSelectOpts(PROPERTY_LOGICAL_OP_OPTS)}
                placeholder="Select Op"
                value={getSelectedOpt(this.props.propertyState.logicalOp, PROPERTY_LOGICAL_OP_OPTS)}
              />
            </div>
          )
        }
        <div style={{display: "inline-block", width: "150px", marginLeft: this.isCondRequired() ? '10px' : null }} className='fapp-select light'>
          <Select
            onChange={this.props.onEntityChange}
            options={createSelectOpts(PROPERTY_TYPE_OPTS)} 
            placeholder="Select Type"
            value={getSelectedOpt(this.props.propertyState.entity, PROPERTY_TYPE_OPTS)}
          />
        </div>
        <div style={{display: "inline-block", width: "240px", marginLeft: "10px"}} className='fapp-select light'>
          <Select
            onChange={this.onNameChange}
            onFocus={this.fetchPropertyKeys}
            options={this.state.nameOpts}
            placeholder="Select Property"
            value={getSelectedOpt(this.props.propertyState.name)}
            isLoading={this.state.isNameOptsLoading}
          />
        </div>
        { this.getOpSelector() }       
        { this.getInputValueElement() }
        <button className='fapp-close-button' onClick={this.props.remove} >x</button>
      </Col>
    </Row>;
  }
}

export default Property;