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
        paddingTop: '60px', 
        paddingBottom: '60px'
      }
      return <CardBody style={style}> Integration is disabled </CardBody>
    }

    let segmentPrivateToken = this.getPrivateToken();
    return (
      <CardBody style={{ padding: '2.5rem 2.5rem' }}>
        <p><strong>API Key: </strong>{segmentPrivateToken}</p>
        Copy the API Key and Follow the <a target='_blank' href='https://www.factors.ai/integrations/segment'>link</a> for instructions. 
      </CardBody>
    ); 
  }

  renderCard() {
    return (
      <Card className='fapp-bordered-card'>
        <div>
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
      <div className='fapp-content fapp-content-margin'>
        { this.renderCard() }
      </div>
    )
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Segment);