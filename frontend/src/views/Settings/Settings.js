import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import Toggle from 'react-toggle'
import {
    Row,
    Col,
    Card,
    CardBody,
    CardHeader
} from 'reactstrap';

import { 
  fetchCurrentProjectSettings, 
  udpateCurrentProjectSettings 
} from '../../actions/projectsActions';


const mapStateToProps = store => {
  return {
    projects: store.projects.projects,
    currentProjectId: store.projects.currentProjectId,
    currentProjectSettings: store.projects.currentProjectSettings
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ fetchCurrentProjectSettings, udpateCurrentProjectSettings }, dispatch)
}

class Settings extends Component {
  constructor(props) {
    super(props);

    this.state = {
      settings: {
        loaded: false,
        error: null,
      }
    }
  }

  componentWillMount() {
    this.props.fetchCurrentProjectSettings(this.props.currentProjectId)
      .then((response) => {
        this.setState({ settings: { loaded: true } });
      })
      .catch((response) => {
        this.setState({ settings: { loaded: true, error: response.payload } });
      });
  }

  isAutoTrackEnabled() {
    return this.props.currentProjectSettings && this.props.currentProjectSettings.auto_track;
  }

  handleAutoTrackToggle = () =>  {
    this.props.udpateCurrentProjectSettings(
      this.props.currentProjectId, { 'auto_track': !this.isAutoTrackEnabled() });
  }

  getToken() {
    return this.props.projects[this.props.currentProjectId].token;
  }

  getSDKScript() {
    // Todo(Dinesh): https://github.com/orgs/Slashbit-Technologies/projects/1#card-15042473
    let token = this.getToken();
    let assetURL = BUILD_CONFIG.sdk_asset_url; // resolved on build.
    return '(function(c){var s=document.createElement("script");s.type="text/javascript";if(s.readyState){s.onreadystatechange=function(){if(s.readyState=="loaded"||s.readyState=="complete"){s.onreadystatechange=null;c()}}}else{s.onload=function(){c()}}s.src="'+assetURL+'";d=!!document.body?document.body:document.head;d.appendChild(s)})(function(){factors.init("'+token+'")})';
  }

  render() {
    if (!this.state.settings.loaded) return <div> Loading... </div>;

    return (
        <div className='animated fadeIn'>
          <div>
            <Row>
              <Col xs='12' md='12'>
                <Card className="fapp-card"> 
                  <CardHeader className="fapp-card-header">
                    <strong>SDK Code</strong>
                  </CardHeader>
                  <CardBody>
                    <Row>
                      <Col md={{ size: '10' }}>                                            
                          <span id="code-snippet" className="sdk-code">{this.getSDKScript()}</span>
                      </Col>
                    </Row>
                  </CardBody>
                </Card>
                <Card className="fapp-card">
                  <CardHeader className="fapp-card-header">
                    <strong>Configure SDK</strong>
                  </CardHeader>
                  <CardBody>
                    <Toggle
                      checked={this.isAutoTrackEnabled()}
                      icons={false}
                      onChange={this.handleAutoTrackToggle} />
                    <span className="fapp-toggle-label">Auto-track</span>
                  </CardBody>
                </Card>
              </Col>
            </Row>
          </div>      
        </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Settings);