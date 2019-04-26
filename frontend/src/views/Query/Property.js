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
import { makeSelectOpts, createSelectOpts, getSelectedOpt } from "../../util";

const PROPERTY_ENTITY_OPTS = {
  'user': 'User Property',
  'event': 'Event Property'
};

const TYPE_NUMERICAL = 'numerical';
const TYPE_CATEGORICAL = 'categorical';

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
    if (values != undefined && values != null)
      this.setState({ valueOpts: makeSelectOpts(values), isValueOptsLoading: false });
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
        <div style={{display: "inline-block", width: "18%", marginLeft: "10px"}}>
          <Input
            type="text"
            onChange={this.onValueChange}
            placeholder="Enter a value"
            value={this.props.propertyState.value}
            style={{border: "1px solid #8f9ba6", color: "#444444"}} 
          />
        </div>
      );
    }
    
    if (this.state.valueType == TYPE_CATEGORICAL) {
      return  (
        <div style={{display: "inline-block", width: "18%", marginLeft: "10px"}} className='fapp-select'>
          <CreatableSelect
            onChange={this.onValueChange}
            onFocus={this.fetchPropertyValues}
            options={this.state.valueOpts}
            value={getSelectedOpt(this.props.propertyState.value)}
            placeholder="Enter a value"
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
      <div style={{display: "inline-block", width: "115px", marginLeft: "10px"}} className='fapp-select'>
        <Select 
          onChange={this.props.onOpChange}
          options={createSelectOpts(optSrc)}
          placeholder="Operator"
          value={getSelectedOpt(this.props.propertyState.op, optSrc) }
        />
      </div>
    );
  }

  nameSelectorDisplay() {
    return this.props.propertyState.entity != '' ? 'inline-block' : 'none';
  }

  render() {
    return <Row style={{marginBottom: "15px"}}>
      <Col xs='12' md='12' style={{marginLeft: "80px"}}>
        <span style={{marginRight: "10px"}}>with</span>
        <div style={{display: "inline-block", width: "15%"}} className='fapp-select'>
          <Select
            onChange={this.props.onEntityChange}
            options={createSelectOpts(PROPERTY_ENTITY_OPTS)}
            placeholder="Property Type"
            value={getSelectedOpt(this.props.propertyState.entity, PROPERTY_ENTITY_OPTS)}
          />
        </div>
        <div style={{display: this.nameSelectorDisplay(), width: "18%", marginLeft: "10px"}} className='fapp-select'>
          <Select
            onChange={this.onNameChange}
            onFocus={this.fetchPropertyKeys}
            options={this.state.nameOpts}
            placeholder="Property Name"
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