import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { getHostURL } from '../../util';
import {
  Card,
  CardBody,
  CardHeader,
  Button,
} from 'reactstrap';

import Loading from '../../loading';
import { 
  fetchProjectSettings,
  udpateProjectSettings,
  fetchAdwordsCustomerAccounts,
  enableSalesforceIntegration,
} from '../../actions/projectsActions';
import salesforceLogo from '../../assets/img/settings/salesforce-logo.svg';

const mapStateToProps = store => {
  return {
    projects: store.projects.projects,
    currentProjectId: store.projects.currentProjectId,
    currentProjectSettings: store.projects.currentProjectSettings,
    adwordsCustomerAccounts: store.projects.adwordsCustomerAccounts,
    currentAgent: store.agents.agent,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchProjectSettings,
    udpateProjectSettings,
    fetchAdwordsCustomerAccounts,
    enableSalesforceIntegration,
  }, dispatch);
}

class Salesforce extends Component {
  constructor(props) {
    super(props);

    this.state = {
      loaded: false,
      redirect: false,
      error: null,

      modalOpen: false,
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

  isSalesforceEnabled() {
    return this.props.currentProjectSettings && 
      this.props.currentProjectSettings.int_salesforce_enabled_agent_uuid && 
      this.props.currentProjectSettings.int_salesforce_enabled_agent_uuid != "";
  }

  getRedirectURL() {
    let host = getHostURL();
    return host +"integrations/salesforce/auth"+"?pid="+this.props.currentProjectId+"&aid="+this.props.currentAgent.uuid;
  }

  isLoaded() {
    return this.state.loaded;
  }

  renderSettingInfo() {
    let style = { fontWeight: 700, color: '#BBB', fontSize: '20px', textAlign: 'center', 
      paddingTop: '60px', paddingBottom: '60px' }
      
    if (this.isSalesforceEnabled()) return <CardBody style={style}> Salesforce sync is enabled </CardBody>;
    return <CardBody style={style}> Salesforce sync is disabled </CardBody>
  }

  onClickEnableSalesforce = () => {
    this.props.enableSalesforceIntegration(this.props.currentProjectId.toString())
      .then((r) => {
        if (r.status == 304) {
          window.location = this.getRedirectURL();
          return
        }
      });
  }

  renderEnableButton(){
      let settingsText = this.isSalesforceEnabled() ? "Enabled": "Enable with Salesforce";
      return (
      <Button color='primary' style={{ marginTop: '-3px' }}
      outline onClick={!this.isSalesforceEnabled() ? this.onClickEnableSalesforce:null}  disabled={this.isSalesforceEnabled()}>
      <img src={salesforceLogo} style={{ marginRight: '6px', marginBottom: '3px', width: '15px' }}></img>
      {settingsText}
      </Button>
      );
  }

  render() {
    if (!this.isLoaded()) return <Loading />;

    return (
      <div className='fapp-content fapp-content-margin'>
        <Card className='fapp-bordered-card'>
          <div>
            <CardHeader  style={{ marginBottom: '0', padding: '15px 20px 25px' }}>
              <strong>Salesforce</strong>
              <div style={{float: 'right'}}>
                  {this.renderEnableButton()}
              </div>
            </CardHeader>
            {this.renderSettingInfo()}
          </div>
        </Card>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Salesforce);