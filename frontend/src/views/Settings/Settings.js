import React, { Component } from 'react';
import { connect } from 'react-redux'
import {
    Row,
    Col,
    Card,
    CardBody,
    CardHeader,
    FormGroup,
    Label,
    Input,
    Button
} from 'reactstrap';
import { udpateCurrentProjectSettings } from '../../actions/projectsActions';

@connect((store) => {
    return {
      currentProject: store.projects.currentProject,
      currentProjectSettings: store.projects.currentProjectSettings
    };
})
class Settings extends Component {
  constructor(props) {
    super(props);
    this.state = {
      settings: {
        auto_track: false
      }
    };
  }

  isAutoTrackEnabled() {
    return this.props.currentProjectSettings 
      && this.props.currentProjectSettings.auto_track == 1;
  }

  handleAutoTrackToggle = () =>  {
    this.props.dispatch(udpateCurrentProjectSettings(
      this.props.currentProject, {'auto_track': !this.isAutoTrackEnabled()}));  
  }
  
  getAutoTrackToggleText() {
    return this.isAutoTrackEnabled() ? "Disable" : "Enable"
  }

  getSDKScript() {
    // Todo(Dinesh): https://github.com/orgs/Slashbit-Technologies/projects/1#card-15042473
    let token = 'YOUR_TOKEN';
    return '(function(c){var s=document.createElement("script");s.type="text/javascript";if(s.readyState){s.onreadystatechange=function(){if(s.readyState=="loaded"||s.readyState=="complete"){s.onreadystatechange=null;c()}}}else{s.onload=function(){c()}}s.src="/dist/factors.prod.js";d=!!document.body?document.body:document.head;d.appendChild(s)})(function(){factors.init("'+token+'")})';
  }

  render() {
    return (
        <div className='animated fadeIn'>
          <div>
            <Row>
              <Col xs='12' md='12'>
                <Card> 
                  <CardHeader>
                    <div style={{ position: "absolute", paddingTop: "0.4rem" }}><strong>Configure</strong></div>
                    <Button onClick={this.handleAutoTrackToggle} color="success" style={{ position: "relative", float:"right", padding: "7px 35px" }}><strong>{this.getAutoTrackToggleText()}</strong></Button>
                  </CardHeader>
                  <CardBody>
                    <Row>
                      <Col md={{ size: '6' }} style={{ paddingLeft: "3rem" }}>                                            
                        <FormGroup>
                          <Label for="code-snippet">Javascript SDK </Label>
                          <Input style={{ height: "17vh" }} type="textarea" name="text" id="code-snippet" value={this.getSDKScript()}/>
                        </FormGroup>
                      </Col>
                    </Row>
                  </CardBody>
                </Card>
              </Col>
            </Row>
          </div>      
        </div>
    );
  }


}

export default Settings;
