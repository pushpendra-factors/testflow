import React, { Component } from 'react';
import Select from 'react-select';
import { connect } from 'react-redux';
// import { bindActionCreators } from 'redux';
import { Row, Col, Input } from 'reactstrap';

import { 
  fetchProjectEventProperties,
  fetchProjectEventPropertyValues,
  fetchProjectUserProperties,
  fetchProjectUserPropertyValues,
} from '../../actions/projectsActions';
import { makeSelectOpts } from "../../util";

const PROPERTY_TYPE_OPTS = [
  { value: "user", label: "User Property" },
  { value: "event", label: "Event Property" }
];

const OPERATOR_EQUALS = { value: 'equals', label: '=' };

const NUMERICAL_OPERATOR_OPTS = [
  OPERATOR_EQUALS,
  { value: 'lesserThan', label: '<' },
  { value: 'lesserThanOrEqual', label: '<=' },
  { value: 'greaterThan', label: '>' },
  { value: 'greaterThanOrEqual', label: '>=' }
];

const CATEGORICAL_OPERATORS_OPTS = [
  OPERATOR_EQUALS
];

class Property extends Component {
  constructor(props) {
    super(props);

    this.state = {
      nameOpts: [],
      valueOpts: [],
      valueType: null
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
    this.setState({ nameOpts: opts });
  }

  addToValueOptsState(values) {
    if (values != undefined && values != null)
      this.setState({ valueOpts: makeSelectOpts(values)});
  }

  fetchPropertyKeys = () => {
    this.setState({ nameOpts: [] }); // reset opts

    if (this.props.propertyState.type == 'event') {
      fetchProjectEventProperties(this.props.projectId, this.props.eventName, false)
        .then((r) => this.addToNameOptsState(r.data))
        .catch(r => console.error("Failed fetching event property keys.", r));
    }

    if (this.props.propertyState.type == 'user') {
      fetchProjectUserProperties(this.props.projectId, false)
      .then((r) => this.addToNameOptsState(r.data))
      .catch(r => console.error("Failed fetching user property keys.", r));
    }
  }

  fetchPropertyValues = () => {
    this.setState({ valueOpts: [] }); // reset opts.

    if (this.props.propertyState.type == 'event') {
      fetchProjectEventPropertyValues(this.props.projectId, 
        this.props.eventName, this.props.propertyState.name, false)
        .then(r => this.addToValueOptsState(r.data))
        .catch(r => console.error("Failed fetching event property values.", r));
    }
    
    if (this.props.propertyState.type == 'user') {
      fetchProjectUserPropertyValues(this.props.projectId, 
        this.props.propertyState.name, false)
        .then(r => this.addToValueOptsState(r.data))
        .catch(r => console.error("Failed fetching user property values.", r));
    }
  }

  onNameChange = (option) => {
    if (option.type != null && 
      option.type != 'numerical' && 
      option.type != 'categorical') {
      
      throw new Error('Unknown property value type.');
    }
    this.setState({ valueType: option.type });
    this.props.onNameChange(option.value);
  }

  onValueChange = (v) => {
    if (this.state.valueType == 'numerical') {
      this.props.onValueChange(v.target.value.trim());
    }
    
    if (this.state.valueType == 'categorical') {
      this.props.onValueChange(v.value);
    }
  }

  getInputValueElement() {
    let input = null;

    if (this.state.valueType == 'numerical') {
      return <div style={{display: "inline-block", width: "15%", marginLeft: "10px"}}>
        <Input
          type="text"
          onChange={this.onValueChange}
          placeholder="Enter a value"
        />
      </div>;
    }
    
    if (this.state.valueType == 'categorical') {
      return  <div style={{display: "inline-block", width: "15%", marginLeft: "10px"}}>
        <Select
          onChange={this.onValueChange}
          onFocus={this.fetchPropertyValues}
          options={this.state.valueOpts}
          placeholder="Enter a value"
        />
      </div>;
    }

    if (this.state.valueType != null) {
      throw new Error('Failed to get input element. Unknown property value type.');
    }
  }

  getOpSelector() {
    if (this.state.valueType == null) {
      return;
    }

    let opts = [];
    if (this.state.valueType == 'numerical') {
      opts = [ ...NUMERICAL_OPERATOR_OPTS ];
    }
    
    if (this.state.valueType == 'categorical') {
      opts = [ ...CATEGORICAL_OPERATORS_OPTS ];
    }

    return (
      <div style={{display: "inline-block", width: "115px", marginLeft: "10px"}}>
        <Select
          onChange={this.props.onOpChange}
          options={opts}
          placeholder="Operator"
        />
      </div>
    );
  }

  nameSelectorDisplay() {
    return this.props.propertyState.type != '' ? 'inline-block' : 'none';
  }

  render() {
    return <Row style={{marginBottom: "15px"}}>
      <Col xs='12' md='12' style={{marginLeft: "80px"}}>
        <span style={{marginRight: "10px"}}>with</span>
        <div style={{display: "inline-block", width: "15%"}}>
          <Select
            onChange={this.props.onTypeChange}
            options={PROPERTY_TYPE_OPTS}
            placeholder="Property Type"
          />
        </div>
        <div style={{display: this.nameSelectorDisplay(), width: "15%", marginLeft: "10px"}}>
          <Select
            onChange={this.onNameChange}
            onFocus={this.fetchPropertyKeys}
            options={this.state.nameOpts}
            placeholder="Property Name"
          />
        </div>        
        { this.getOpSelector() }       
        { this.getInputValueElement() }
      </Col>
    </Row>;
  }
}

export default Property;