import React, { Component } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  Card,
  CardBody,
  CardHeader,
  Button,
  Table,
  Input,
  Modal,
  ModalBody,
  ModalHeader,
  Form,
  
} from 'reactstrap';

import Loading from '../../loading';
import { 
  fetchProjectSettings,
  udpateProjectSettings,
  fetchAdwordsCustomerAccounts,
  enableAdwordsIntegration,
} from '../../actions/projectsActions';
import googleSvg from '../../assets/img/settings/google_sso.svg';
import { getAdwordsHostURL } from '../../util';

const ADWORDS_REDIRECT_URI="/adwords/auth/redirect";

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
    enableAdwordsIntegration,
  }, dispatch);
}

class Adwords extends Component {
  constructor(props) {
    super(props);

    this.state = {
      loaded: false,
      redirect: false,
      error: null,

      customerAccountsLoaded: false,
      selectedAdwordsAccounts: [],
      modalOpen: false,
      accountId: '',
      manualAccounts: [],
      addNewAccount: false
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

  isIntAdwordsEnabled() {
    return this.props.currentProjectSettings && 
      this.props.currentProjectSettings.int_adwords_enabled_agent_uuid && 
      this.props.currentProjectSettings.int_adwords_enabled_agent_uuid != "";
  }

  getRedirectURL() {
    let host = getAdwordsHostURL();
    return host + ADWORDS_REDIRECT_URI+"?pid="+this.props.currentProjectId+"&aid="+this.props.currentAgent.uuid;
  }

  isLoaded() {
    return this.state.loaded;
  }

  onAccountSelect = (e) => {
    let selectedAdwordsAccounts = [...this.state.selectedAdwordsAccounts]
    if(e.target.checked) {
      selectedAdwordsAccounts.push(e.target.value)
    } else {
      let index = selectedAdwordsAccounts.indexOf(e.target.value)
      if(index>-1) selectedAdwordsAccounts.splice(index,1)
    }
    this.setState({
      selectedAdwordsAccounts
    })
  }

  onClickFinishSetup = () => {
    let selectedAdwordsAccounts = this.state.selectedAdwordsAccounts.join(",")
    this.props.udpateProjectSettings(this.props.currentProjectId, 
      { 'int_adwords_customer_account_id':  selectedAdwordsAccounts});
    this.setState({
        addNewAccount: false,
        selectedAdwordsAccounts: []
      })
  }

  renderAccountsList = () => {
    let accountRows = [];

    if (!this.props.adwordsCustomerAccounts) return;

    for (let i=0; i<this.props.adwordsCustomerAccounts.length; i++) {
      let account = this.props.adwordsCustomerAccounts[i];

      accountRows.push(
        <tr>
          <td style={{ border: 'none', padding: '5px'  }}>
            <Input type="checkbox" value={account.customer_id} onChange={this.onAccountSelect} />
          </td>
          <td style={{ border: 'none', padding: '5px'  }}>{ account.customer_id }</td>
          <td style={{ border: 'none', padding: '5px' }}>{ account.name }</td>
        </tr>
      )
    }
    for (let i=0;i<this.state.manualAccounts.length; i++) {
      let account = this.state.manualAccounts[i];

      accountRows.push(
        <tr>
          <td style={{ border: 'none', padding: '5px'  }}>
            <Input type="checkbox" value={account.customer_id} onChange={this.onAccountSelect} />
          </td>
          <td style={{ border: 'none', padding: '5px'  }}>{ account.customer_id }</td>
          <td style={{ border: 'none', padding: '5px' }}>{ account.name }</td>
        </tr>
      )
    }
    
    return (
      <CardBody style={{paddingLeft: '50px', maxWidth: '50%'}}> 
        <div style={{ paddingBottom: '20px', fontSize: '15px', color: '#444', fontWeight: '500'}}>Choose an adwords account</div>
        
        <Table>
          <thead>
            <tr>
              <td style={{ border: 'none', padding: '5px' }}></td>
              <td style={{ fontWeight: '700', border: 'none', padding: '5px' }}>Customer Id</td>
              <td style={{ fontWeight: '700', border: 'none', padding: '5px'  }}>Customer Name</td>
            </tr>
          </thead>
          <tbody>{ accountRows }</tbody>
        </Table>
        <div><Button color='primary' outline style={{ marginTop: '30px' }} onClick={()=> this.setState({modalOpen: true})}> Add Manually </Button>
        </div>
        <Button color='success' outline style={{ marginTop: '30px' }} onClick={this.onClickFinishSetup}> Finish Setup </Button>
        
        <Modal isOpen={this.state.modalOpen}>
          <ModalHeader>Enter adwords account ID:</ModalHeader>
          <ModalBody>
            <Input type="text" onChange={(e)=> this.setState({accountId: e.target.value})}/>
            <div className="d-flex justify-content-around">
              <Button color='success' outline style={{ marginTop: '30px' }} onClick={this.addManualAccount}> Submit </Button>
              <Button color='danger' outline style={{ marginTop: '30px' }} onClick={()=>this.setState({modalOpen: false})}> Cancel </Button>
            </div>
          </ModalBody>
        </Modal>
      </CardBody>
    );
  }
  addManualAccount = () => {
    let accounts =[...this.state.manualAccounts]
    if(this.state.accountId != "") {
      accounts.push(
        {
          customer_id: this.state.accountId
        }
      )
    }
    this.setState({
      modalOpen: false,
      manualAccounts: accounts
    })
  }

  isCustomerAccountSelected() {
    return  this.props.currentProjectSettings && this.props.currentProjectSettings.int_adwords_customer_account_id && !this.state.addNewAccount;
  }

  renderSettingInfo() {
    let style = { fontWeight: 700, color: '#BBB', fontSize: '20px', textAlign: 'center', 
      paddingTop: '60px', paddingBottom: '60px' }
      
    let isCustomerAccountChosen = this.props.currentProjectSettings.int_adwords_customer_account_id && 
      this.props.currentProjectSettings.int_adwords_customer_account_id != "" && !this.state.addNewAccount;
    
    // get all adwords account when no account is chosen and not account list not loaded.
    if (this.isIntAdwordsEnabled() && !isCustomerAccountChosen && !this.state.customerAccountsLoaded) {
      this.props.fetchAdwordsCustomerAccounts({ "project_id": this.props.currentProjectId })
        .then(() => { this.setState({ customerAccountsLoaded: true }); })
    }

    if (this.isCustomerAccountSelected()) {
      return <CardBody style={{ padding: '2rem 3rem' }}>
        <div>
          <div style={{ paddingBottom: '10px', fontSize: '15px', fontWeight: '500', color: '#444'}}>
            Adwords sync account details
          </div>
          <span style={{ fontWeight: '700' }}>Account Id:</span> { this.props.currentProjectSettings.int_adwords_customer_account_id }
        </div>
      </CardBody>
    }

    if (this.state.customerAccountsLoaded) return this.renderAccountsList();

    return <CardBody style={style}> Adwords sync is disabled </CardBody>
  }

  onClickEnableAdwords = () => {
    this.props.enableAdwordsIntegration(this.props.currentProjectId)
      .then((r) => {
        if (r.status == 304) {
          window.location = this.getRedirectURL();
          return
        }
      });
  }
  
  render() {
    if (!this.isLoaded()) return <Loading />;

    return (
      <div className='fapp-content fapp-content-margin'>
        <Card className='fapp-bordered-card'>
          <div>
            <CardHeader  style={{ marginBottom: '0', padding: '15px 20px 25px' }}>
              <strong>Adwords</strong>
              <div style={{float: 'right'}}>
                <Button hidden={this.isIntAdwordsEnabled()} color='primary' style={{ marginTop: '-3px' }} 
                  outline onClick={this.onClickEnableAdwords}> 
                  <img src={googleSvg} style={{ marginRight: '6px', marginBottom: '3px', width: '15px' }}></img>
                  Enable with Google
                </Button>
                <Button hidden={!this.isIntAdwordsEnabled()} color='primary' style={{ marginTop: '-3px' }} 
                  outline onClick={()=> this.setState({addNewAccount: true})}> 
                  + Add More
                </Button>
              </div>
            </CardHeader>
            {this.renderSettingInfo()}
          </div>
        </Card>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(Adwords);