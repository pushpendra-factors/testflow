import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Col,
  Input,
  Card,
  CardBody,
  CardHeader
} from 'reactstrap';
import Toggle from 'react-toggle';

import Loading from '../../loading';
import { 
  fetchProjectSettings,
  udpateProjectSettings,
} from '../../actions/projectsActions';
import NoContent from '../../common/NoContent';

const INT_SEGMENT_URI="/integrations/segment";

const mapStateToProps = store => {
  return {
    projects: store.projects.projects,
    currentProjectId: store.projects.currentProjectId,
    currentProjectSettings: store.projects.currentProjectSettings,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchProjectSettings,
    udpateProjectSettings,
  }, dispatch)
}

class Segment extends Component {
  constructor(props) {
    super(props);

    this.state = {
      loaded: false,
      error: null
    }
  }

  componentWillMount() {
    this.props.fetchProjectSettings(this.props.currentProjectId)
      .then((r) => {
        this.setState({ loaded: true });
      })
      .catch((r) => {
        this.setState({loaded: true, error: r.payload });
      });
  }

  isIntSegmentEnabled() {
    return this.props.currentProjectSettings && 
      this.props.currentProjectSettings.int_segment;
  }

  toggleIntSegment = () =>  {
    this.props.udpateProjectSettings(this.props.currentProjectId, 
      { 'int_segment': !this.isIntSegmentEnabled() });
  }

  getSegmentWebhookURL() {
    return BUILD_CONFIG.backend_host+INT_SEGMENT_URI;
  }

  getPrivateToken() {
    return this.props.projects[this.props.currentProjectId].private_token;
  }

  isLoaded() {
    return this.state.loaded;
  }

  renderSegmentConfig() {
    if (!this.isIntSegmentEnabled()) {
      let style = { 
        fontWeight: 700, 
        color: '#BBB', 
        fontSize: '20px', 
        textAlign: 'center', 
        paddingTop: '110px', 
        paddingBottom: '130px'
      }
      return <CardBody style={style}> Integration is disabled </CardBody>
    }

    let segmentWebhookURL = this.getSegmentWebhookURL();
    let segmentPrivateToken = this.getPrivateToken();
    return (
      <CardBody>
        <div style={{marginBottom: '25px'}}>
          <span className='fapp-label'>Webhook URL</span>
          <Input className='fapp-input' defaultValue={segmentWebhookURL}/>
        </div>
        <div>
          <span className='fapp-label'>API Key</span>
          <Input className='fapp-input' defaultValue={segmentPrivateToken}/>
        </div>
      </CardBody>
    ); 
  }

  renderCard() {
    return (
      <Card className='fapp-bordered-card'>
        <div className={!this.props.cardOnly ? 'fapp-block-shadow' : null}>
          <CardHeader className='fapp-button-header' style={{ marginBottom: '0' }}>
            <strong>Segment</strong>
            <div style={{display: 'inline-block', float: 'right'}}>
              <Toggle
                checked={this.isIntSegmentEnabled()}
                icons={false}
                onChange={this.toggleIntSegment} 
              />
            </div>
          </CardHeader>
          { this.renderSegmentConfig() }
        </div>
      </Card>
    );
  }
  
  render() {
    if (!this.isLoaded()) return <Loading />;

    if (this.props.cardOnly) return this.renderCard();

    return (
      <Col md={{ size:6, offset:3 }} className='fapp-content fapp-content-margin' style={{ padding: '5rem' }}>
        { this.renderCard() }
      </Col>
    )
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Segment);