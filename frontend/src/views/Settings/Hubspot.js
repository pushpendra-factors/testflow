import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Input,
  Card,
  CardBody,
  CardHeader,
  Button,
} from 'reactstrap';
import Toggle from 'react-toggle';

import Loading from '../../loading';
import { 
  fetchProjectSettings,
  udpateProjectSettings,
} from '../../actions/projectsActions';


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

class Hubspot extends Component {
  constructor(props) {
    super(props);

    this.state = {
      apiKey: "",
      change: false,
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

  isIntHubspotEnabled() {
    return this.props.currentProjectSettings && 
      this.props.currentProjectSettings.int_hubspot;
  }

  toggleIntHubspot = () =>  {
    this.props.udpateProjectSettings(this.props.currentProjectId, 
      { 'int_hubspot': !this.isIntHubspotEnabled() });
  }

  isLoaded() {
    return this.state.loaded;
  }

  updateAPIKey = () => {
    this.props.udpateProjectSettings(this.props.currentProjectId, 
      { 'int_hubspot_api_key': this.state.apiKey });

    // on update api key.
    this.setState({ change: false });
  }

  changeAPIKey = () => {
    // copy key to local state and set change as true.
    this.setState({ apiKey: this.props.currentProjectSettings.int_hubspot_api_key, change: true });
  }

  renderHubspotConfig() {
    if (!this.isIntHubspotEnabled()) {
      let style = { 
        fontWeight: 700, 
        color: '#BBB', 
        fontSize: '20px', 
        textAlign: 'center', 
        paddingTop: '55px', 
        paddingBottom: '55px'
      }
      return <CardBody style={style}> Integration is disabled </CardBody>
    }

    let apiKey = this.props.currentProjectSettings.int_hubspot_api_key;
    if (apiKey && apiKey != "" && !this.state.change) {
      return (
          <CardBody style={{ padding: '3.5rem' }}>
            <div style={{ marginRight: "20px", display: "inline" }}><span style={{ fontWeight: "700" }}>API Key:</span> { this.props.currentProjectSettings.int_hubspot_api_key }</div>
            <Button color="primary" outline onClick={this.changeAPIKey} size="sm">Change</Button>
          </CardBody>
      ); 
    }

    
    return (
      <CardBody style={{ padding: '3.5rem' }}>
        <Input
          style={{ display: "inline-block", width: "300px", marginRight: "15px", height: "36px" }}
          placeholder="Your API Key."
          onChange={(e) => this.setState({ apiKey: e.target.value })}
          value={this.state.apiKey}
        />
        <Button 
          style={{ display: "inline-block", marginBottom: "1px" }} 
          color="primary" outline
          onClick={this.updateAPIKey}
        >
        { this.state.change ? "Update" : "Add" }
        </Button>
      </CardBody>
    );
  }

  renderCard() {
    return (
      <Card className='fapp-bordered-card'>
        <div>
          <CardHeader className='fapp-button-header' style={{ marginBottom: '0' }}>
            <strong>Hubspot</strong>
            <div style={{display: 'inline-block', float: 'right'}}>
              <Toggle
                checked={this.isIntHubspotEnabled()}
                icons={false}
                onChange={this.toggleIntHubspot} 
              />
            </div>
          </CardHeader>
          { this.renderHubspotConfig() }
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

export default connect(mapStateToProps, mapDispatchToProps)(Hubspot);