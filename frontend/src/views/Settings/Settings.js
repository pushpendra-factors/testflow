import React, { Component } from 'react';
import { Redirect } from 'react-router-dom';
import {
    Row,
    Col,
    Card,
    CardBody,
    CardHeader,
} from 'reactstrap';

import projectUsersSvg from '../../assets/img/settings/project_users.svg';
import jsSvg from '../../assets/img/settings/js.svg';
import segmentSvg from '../../assets/img/integrations/segment.svg';
import androidSvg from '../../assets/img/settings/android.svg';
import iosSvg from '../../assets/img/settings/iOS.svg';
import adwordsSvg from '../../assets/img/integrations/adwords.svg';
import hubspotSvg from '../../assets/img/integrations/hubspot.svg'
import facebookSvg from '../../assets/img/integrations/facebook.svg'
import salesforceLogo from '../../assets/img/integrations/salesforce.svg';

class SettingsCard extends Component {
  constructor(props) {
    super(props);
    
    this.state = {
      clicked: false
    }
  }

  handleClick = () => {
    this.setState({ clicked: true });
  }

  render() {
    if (this.state.clicked) return <Redirect to={this.props.href} />;
    let size = this.props.size ? this.props.size : '65px';

    return (
      <Col xs='12' md={{size: 2}} sm={{size: 6}} style={{marginBottom: '1rem', marginTop: '1rem'}} onClick={this.handleClick}>
        <div style={{border: '2px solid #ddd'}} className='setting-card'> 
          <div style={{width: size, height: size, margin: '25px auto'}}>
            <img src={this.props.img} style={{ width: '100%', height:'100%'}} />
          </div>
          <strong style={{textAlign: 'center', display: 'inherit', fontSize: '15px', fontWeight: 500, color: '#484848', paddingBottom: '25px'}}>{this.props.title}</strong>
        </div>
      </Col>
    );
  } 
}

class Settings extends Component {
  constructor(props) {
    super(props);
  }

  shouldComponentUpdate(nextProps, nextState) {
    // decide to render or not.
    return true
  }

  componentDidUpdate(prevProps, prevState) {
    // set state based on action or prevProps.
    // use conditions, like this.props.prop1 == prevProps.prop1;
  }

  render() {
    return (
      <div className='animated fadeIn fapp-content fapp-content-margin'>
        <Row>
          <Col xs='12' md='12'>
            <Card className="fapp-card">
              <CardHeader>
                <strong>General Settings</strong>
              </CardHeader>
              <CardBody style={{padding: '0 10px'}}>
                <Row>
                  <SettingsCard title='Project Users' img={projectUsersSvg} size='75px' href='/settings/agents' />
                </Row>
              </CardBody>
            </Card>
            <Card className="fapp-card">
              <CardHeader>
                <strong>SDKs</strong>
              </CardHeader>
              <CardBody style={{padding: '0 10px'}}>
                <Row>
                  <SettingsCard title='Javascript SDK' img={jsSvg} href='/settings/jssdk' />
                  <SettingsCard title='Android SDK' img={androidSvg} href='/settings/androidsdk' />
                  <SettingsCard title='IOS SDK' img={iosSvg} href='/settings/iossdk'/>
                </Row>
              </CardBody>
            </Card>
            <Card className="fapp-card">
              <CardHeader>
                <strong>Integrations</strong>
              </CardHeader>
              <CardBody style={{padding: '0 10px'}}>
                <Row>
                  <SettingsCard title='Segment' img={segmentSvg} href='/settings/segment' />
                  <SettingsCard title='Adwords' img={adwordsSvg} href='/settings/adwords' />
                  <SettingsCard title='Hubspot' img={hubspotSvg} href='/settings/hubspot' />
                  <SettingsCard title='Facebook' img={facebookSvg} href='/settings/facebook' />
                  <SettingsCard title='Salesforce' img={salesforceLogo} href='/settings/Salesforce' />
                </Row>
              </CardBody>
            </Card>
          </Col>
        </Row>
      </div>
    );
  }
}

export default Settings;