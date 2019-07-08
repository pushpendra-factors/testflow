import React, { Component } from 'react';
import { DropdownItem, DropdownMenu, DropdownToggle, Input, Button, Form, Nav,
  Modal, ModalHeader, ModalBody, ModalFooter, Badge } from 'reactstrap';
import PropTypes from 'prop-types';
import { AppHeaderDropdown, AppSidebarToggler } from '@coreui/react';
import { AppSidebarForm } from '@coreui/react';
import Select from 'react-select';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import Avatar from 'react-avatar';

import { changeProject, createProject } from '../../actions/projectsActions';
import { signout } from '../../actions/agentActions';
import factorsai from '../../common/factorsaiObj';

const propTypes = {
  children: PropTypes.node,
};

const defaultProps = {};

const projectSelectStyles = {
  option: (base, state) => ({
    ...base,
    color: '#5c6873',
    background: '#fff',
  }),
  singleValue: base => ({
    ...base,
    background: '#fff',
    color: '#5c6873',
  }),
  valueContainer: base => ({
    ...base,
    background: '#fff',
    color: '#5c6873',
  }),
  container: base => ({
    ...base,
    background: '#fff',
    border: 'none',
  }),
  indicatorSeparator: () => ({
    display: 'none',
  }),
}

const mapStateToProps = store => {
  return {
    billingAccount: store.agents.billing.billingAccount,
    accountPlan: store.agents.billing.plan,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    changeProject, 
    createProject, 
    signout,
  }, dispatch);
}

class DefaultHeader extends Component {
  constructor(props){
    super(props);
    this.state = {
      createProject: {
        showForm: false,
        projectName : ""
      },
      
      showAddProjectModal: false,
      addProjectMessage: null,
    }
  }

  componentDidUpdate() {
    // change to project last seen by user if any.
    let lsProjectId = this.getLastSeenProject();
    if (lsProjectId && this.props.selectedProject && 
      this.props.selectedProject.value != lsProjectId && 
      this.hasProject(lsProjectId))
        this.props.changeProject(lsProjectId);
  }

  handleProjectChange = (selectedProject) => {
    let projectId = selectedProject.value;
    this.props.changeProject(projectId);
    this.setLastSeenProject(projectId);

    this.props.refresh();
  }

  getLastSeenProjectKey() {
    // _project_ls:<agent_id>
    return this.props.currentAgent ? '_project_ls:'+this.props.currentAgent.uuid : '';
  }
  
  setLastSeenProject(projectId) {
    let projectKey = this.getLastSeenProjectKey();
    if (projectKey == '') return;
    localStorage.setItem(projectKey, projectId);
  }

  getLastSeenProject() {
    let projectKey = this.getLastSeenProjectKey();
    if (projectKey == '') return null;
    return localStorage.getItem(projectKey);
  }

  hasProject(projectId) {
    for(let i=0; i<this.props.selectableProjects.length; i++) {
      let sProject = this.props.selectableProjects[i];
      if (sProject.value == projectId) return true;
    }

    return false;
  }

  handleProjectNameFormChange = (e) => {
    this.setState({ addProjectMessage: null });

    let name = e.target.value.trim();
    if(name == "") console.error("project name cannot be empty");
    this.setState({ createProject: { projectName: name } });
  }

  changeViewToAccountSettings = () => {
    this.props.changeViewToAccountSettings();
  }

  changeViewToUserProfile = () => {
    this.props.changeViewToUserProfile();
  }

  toggleCreateProjectForm = () => {
    this.setState(prevState => {
      let _state = { ...prevState };
      _state.createProject.showForm = !prevState.createProject.showForm;
      return _state;
    })
  }
  
  handleCreateProject = (e) => {
    e.preventDefault();

    let projectName = this.state.createProject.projectName;
    if(projectName == "") {
      this.showAddProjectMessage({success: false, message: 'Your project name cannot be empty'});
      return;
    }

    this.props.createProject(projectName)
      .then((r) => this.setState({ showAddProjectModal: false }))
      .catch((r) => this.showAddProjectMessage({success: false, message: 'Failed to create your project. Please try again.'}))
  }

  handleLogout = () => {
    this.props.signout();
    factorsai.track('logout', { email: this.props.currentAgent.email });
  }

  toggleAddProjectModal = () => {
    this.setState((pState) => {
      let state = { showAddProjectModal: !pState.showAddProjectModal }
      if (!state.showAddProjectModal) {
        // reset message on close.
        state.addProjectMessage = null;
      }
      return state
    });
  }

  renderNotifications = () => {

    if (!this.props.billingAccount){
      return
    }
    let noOfNotifications = 0;
    let dropDownItems = [];
    if (this.props.accountPlan.code != "free" && !this.props.billingAccount.organization_name){
      dropDownItems.push(<DropdownItem key={1} onClick={this.changeViewToAccountSettings}><span className="text-muted">Complete Billing Info</span></DropdownItem>);
      noOfNotifications++;
    }else{
      dropDownItems.push(<DropdownItem disabled key={0}><span className="text-muted">No messages here.</span></DropdownItem>);
    }

    return (
      <AppHeaderDropdown direction="down">
        <DropdownToggle nav>	
          <i className="icon-bell fapp-bell"></i>	
          { noOfNotifications > 0 && <Badge pill color="danger">{noOfNotifications}</Badge> }	
        </DropdownToggle>	
        <DropdownMenu right style={{ right: 'auto' }}>
          {
            dropDownItems.map((item) => {
              return item;
            })
          }
        </DropdownMenu>
      </AppHeaderDropdown>
    )
  }

  showAddProjectMessage(msg) {
    this.setState({addProjectMessage: msg});
  }

  getAddProjectMessage() {
    if (this.state.addProjectMessage == null) return '';
    return this.state.addProjectMessage.message;
  }

  render() {
    // eslint-disable-next-line
    const { children, ...attributes } = this.props;

    let selectProjectDropDown = null;
    if (!!this.props.selectableProjects)
      selectProjectDropDown = <Select 
        options={this.props.selectableProjects} 
        value={this.props.selectedProject} 
        onChange={this.handleProjectChange} 
        styles={projectSelectStyles} 
        placeholder={"Select Project ..."} 
        blurInputOnSelect={true}
      />;

    return (
      <React.Fragment>
        <AppSidebarToggler className="d-lg-none" display="md" mobile />
        {/* <AppSidebarToggler className="d-md-down-none fapp-navbar-toggler" display="lg" /> */}
        <AppSidebarForm className="fapp-select fapp-header-dropdown" style={{width: '40%'}}>
          <div style={{display: 'inline-block', width: '60%', marginRight: '5px'}}> { selectProjectDropDown } </div>
          <Button outline color="primary" onClick={this.toggleAddProjectModal} style={{fontSize: '20px', padding: '0 10px', height: '38px'}}>+</Button>
        </AppSidebarForm>
        <Nav className="ml-auto fapp-header-right" navbar>
            { this.renderNotifications() }
          <AppHeaderDropdown direction="down">  
            <DropdownToggle nav>
              <Avatar name={this.props.getProfileName()}  maxInitials={1} round={true} color='#3a539b' textSizeRatio={2} size='35' style={{fontWeight: '700', marginTop: '5px'}} />
            </DropdownToggle>
            <DropdownMenu right style={{ right: 'auto' }}>
              <DropdownItem onClick = {() => {this.changeViewToUserProfile();}}><i className="fa fa-user"></i> Profile</DropdownItem>
              <DropdownItem onClick={()=>{this.changeViewToAccountSettings();}} ><i className="fa fa-wrench"></i>Account Settings</DropdownItem>
              <DropdownItem onClick={this.handleLogout}><i className="fa fa-lock"></i> Logout</DropdownItem>
            </DropdownMenu>
          </AppHeaderDropdown>
        </Nav>
        <Modal isOpen={this.state.showAddProjectModal} toggle={this.toggleAddProjectModal} style={{marginTop: '10rem'}}>
          <ModalHeader toggle={this.toggleAddProjectModal}>New Project</ModalHeader>
          <ModalBody style={{padding: '25px 35px'}}>
            <div style={{textAlign: 'center', marginBottom: '15px'}}>
              <span style={{display: 'inline-block'}} className='fapp-error' hidden={this.state.addProjectMessage == null}>{ this.getAddProjectMessage() }</span>
            </div>
            <Form onSubmit={this.handleCreateProject} >
              <span className='fapp-label'>Project Name </span>         
              <Input className='fapp-input' type="text" placeholder="Your Project Name" onChange={this.handleProjectNameFormChange} />
            </Form>
          </ModalBody>
          <ModalFooter style={{borderTop: 'none', paddingBottom: '30px', paddingRight: '35px'}}>
            <Button outline color="success" onClick={this.handleCreateProject}>Create</Button>
            <Button outline color='danger' onClick={this.toggleAddProjectModal}>Cancel</Button>
          </ModalFooter>
        </Modal>
      </React.Fragment>
    );
  }
}

DefaultHeader.propTypes = propTypes;
DefaultHeader.defaultProps = defaultProps;

export default connect(mapStateToProps, mapDispatchToProps)(DefaultHeader);