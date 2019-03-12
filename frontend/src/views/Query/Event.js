import { makeSelectOpts } from '../../util';
import React, { Component } from 'react';
import { Row, Col, Button } from 'reactstrap';
import Select from 'react-select';
import Property from './Property';



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
          key={key} 
          projectId={this.props.projectId}
          propertyState={this.props.eventState.properties[i]}
          eventName={this.props.eventState.name}
          
          onTypeChange={(option) => this.props.onPropertyTypeChange(this.props.index, i, option.value)}
          onNameChange={(value) => this.props.onPropertyNameChange(this.props.index, i, value)}
          onOpChange={(option) => this.props.onPropertyOpChange(this.props.index, i, option.value)}
          onValueChange={(value) => this.props.onPropertyValueChange(this.props.index, i, value)}
        />
      );
    }
    return properties
  }

  addPropertyDisplay() {
    return this.props.eventState.name != '' ? 'inline-block' : 'none';
  }

  render() {
    return (
      <div>
        <Row style={{marginBottom: '15px'}}>
          <Col xs='12' md='12' style={{marginLeft: '70px'}}>
            <div style={{display: 'inline-block', width: '15%'}}>
              <Select
                onChange={this.props.onNameChange}
                options={makeSelectOpts(this.props.nameOpts)}
                placeholder='Event name'
              />
            </div>
            <Button outline color='primary' style={{marginLeft: '10px', display: this.addPropertyDisplay()}} onClick={this.props.onAddProperty} >+ Property</Button>
          </Col>         
        </Row>
        { this.getProperties() }
      </div>
    );
  }
}

export default Event;