import React, { Component } from 'react';
import { Row, Col, Button } from 'reactstrap';
import Select from 'react-select';
import Property from './Property';

import { makeSelectOpts, getSelectedOpt } from '../../util';

class Event extends Component {
  constructor(props) {
    super(props);
  }

  getProperties() {
    let properties = [];
    for(let i=0 ; i<this.props.eventState.properties.length; i++) {
      let key = 'event_'+this.props.index+'_prop_'+i;
      properties.push(
        <Property 
          index={i}
          key={key} 
          projectId={this.props.projectId}
          propertyState={this.props.eventState.properties[i]}
          eventName={this.props.eventState.name}
          remove={() => this.props.removeProperty(i)}
          
          onEntityChange={(option) => this.props.onPropertyEntityChange(this.props.index, i, option.value)}
          onNameChange={(value) => this.props.onPropertyNameChange(this.props.index, i, value)}
          onOpChange={(option) => this.props.onPropertyOpChange(this.props.index, i, option.value)}
          onValueChange={(value, type) => this.props.onPropertyValueChange(this.props.index, i, value, type)}
        />
      );
    }
    return properties
  }

  getAddPropertyText() {
    return this.props.eventState.properties.length == 0 ? "with property" : "+ Property";
  }

  render() {
    return (
      <div>
        <Row style={{marginBottom: '15px'}}>
          <Col xs='12' md='12'>
            <div style={{display: 'inline-block', width: '275px'}} className='fapp-select light'>
              <Select
                onChange={this.props.onNameChange}
                options={makeSelectOpts(this.props.nameOpts)} 
                placeholder='Select an event'
                value={getSelectedOpt(this.props.eventState.name)}
              />
            </div>
            <Button outline color='primary' style={{marginLeft: '10px', display: 'inline-block', height: '100%'}} onClick={this.props.onAddProperty} >{ this.getAddPropertyText() }</Button>
            <button className='fapp-close-button' onClick={this.props.remove}>x</button>
          </Col>         
        </Row>
        { this.getProperties() }
      </div>
    );
  }
}

export default Event;