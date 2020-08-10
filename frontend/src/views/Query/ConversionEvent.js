import React, { Component } from 'react';
import { Row, Col, Button } from 'reactstrap';
import CreatableSelect from 'react-select/lib/Creatable';
import Property from './Property';
import { makeSelectOpts, getSelectedOpt } from '../../util';

class ConversionEvent extends Component {
  constructor(props) {
    super(props);
  }

  getProperties() {
    let properties = [];
    for(let i=0 ; i<this.props.eventState.properties.length; i++) {
      let key = 'event_prop_'+i;
      properties.push(
        <Property 
          index={i}
          key={key} 
          projectId={this.props.projectId}
          propertyState={this.props.eventState.properties[i]}
          eventName={this.props.eventState.name}
          remove={() => this.props.removeProperty(i)}
          
          onEntityChange={(option) => this.props.onPropertyEntityChange(i, option.value)}
          onLogicalOpChange={(option) => this.props.onPropertyLogicalOpChange(i, option.value)}
          onNameChange={(value) => this.props.onPropertyNameChange(i, value)}
          onOpChange={(option) => this.props.onPropertyOpChange(i, option.value)}
          onValueChange={(value, type) => this.props.onPropertyValueChange(i, value, type)}
        />
      );
    }
    return properties
  }

  render() {
    return (
      <div>

        <Row style={{marginBottom: '15px', marginTop: "-8px"}}>

          <Col xs='2' md='2' style={{paddingTop: "5px"}}>
            <span style={{marginRight: '10px', fontWeight: '600', color: '#777'}}> Select Conversion Event</span>
          </Col>
          <Col xs='8' md='8'>
            <div className='fapp-select light' style={{display: 'inline-block', width: '250px'}}>
              <CreatableSelect
                onChange={this.props.onNameChange}
                options={makeSelectOpts(this.props.nameOpts)} 
                placeholder='Select'
                value={getSelectedOpt(this.props.eventState.name)}
                formatCreateLabel={(value) => (value)}
              />
            </div>
            <Button outline color='primary' style={{marginLeft: '10px', display: 'inline-block', height: '100%'}} onClick={this.props.onAddProperty} >+ Filter</Button>
          </Col>
        </Row>

        { this.getProperties() }
      </div>
    );
  }
}

export default ConversionEvent;