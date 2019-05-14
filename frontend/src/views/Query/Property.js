import React, { Component } from 'react';
import Select from 'react-select';
import CreatableSelect from 'react-select/lib/Creatable';
import { connect } from 'react-redux';
// import { bindActionCreators } from 'redux';
import { Row, Col, Input } from 'reactstrap';

import { 
  fetchProjectEventProperties,
  fetchProjectEventPropertyValues,
  fetchProjectUserProperties,
  fetchProjectUserPropertyValues,
} from '../../actions/projectsActions';
import { makeSelectOpts, createSelectOpts, getSelectedOpt, makeSelectOpt } from "../../util";
import { PROPERTY_VALUE_NONE, PROPERTY_TYPE_OPTS } from "./common";

const TYPE_NUMERICAL = 'numerical';
const TYPE_CATEGORICAL = 'categorical';

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
      fetchProjectEventProperties(this.props.projectId, this.props.eventName, false)
        .then((r) => this.addToNameOptsState(r.data))
        .catch(r => console.error("Failed fetching event property keys.", r));
    }

    if (this.props.propertyState.entity == 'user') {
      fetchProjectUserProperties(this.props.projectId, false)
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
      option.type != TYPE_CATEGORICAL) {
      
      throw new Error('Unknown property value type.');
    }
    this.setState({ valueType: option.type });
    this.props.onNameChange(option.value);
  }

  onValueChange = (v) => {
    if (this.state.valueType == TYPE_NUMERICAL) {
      this.props.onValueChange(v.target.value.trim(), this.state.valueType);
    }
    
    if (this.state.valueType == TYPE_CATEGORICAL) {
      this.props.onValueChange(v.value, this.state.valueType);
    }
  }

  getInputValueElement() {
    let input = null;

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

    if (this.state.valueType != null) {
      throw new Error('Failed to get input element. Unknown property value type.');
    }
  }

  getOpSelector() {
    if (this.state.valueType == null) {
      return;
    }
    
    // categorical_operator_opts as default.
    let optSrc = this.state.valueType == TYPE_NUMERICAL ? NUMERICAL_OPERATOR_OPTS : CATEGORICAL_OPERATORS_OPTS;

    return (
      <div style={{display: "inline-block", width: "65px", marginLeft: "10px"}} className='fapp-select light'>
        <Select 
          onChange={this.props.onOpChange}
          options={createSelectOpts(optSrc)}
          value={getSelectedOpt(this.props.propertyState.op, optSrc) }
        />
      </div>
    );
  }

  isAndJoin() {
    return this.props.index > 0
  }

  getJoinStr() {
    return this.isAndJoin() ? "and" : "with";
  }

  render() {
    return <Row style={{marginBottom: "15px"}}>
      <Col xs='12' md='12'>
        <span style={LABEL_STYLE}> { this.getJoinStr() } </span>
        <div style={{display: "inline-block", width: "150px", marginLeft: this.isAndJoin() ? "5px" : null}} className='fapp-select light'>
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