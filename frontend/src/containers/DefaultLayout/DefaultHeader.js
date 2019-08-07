import React, { Component } from 'react';
import { DropdownItem, DropdownMenu, DropdownToggle, Input, Button, Form, Nav,
  Modal, ModalHeader, ModalBody, ModalFooter, Badge, TabContent, TabPane, NavItem, 
  NavLink, Card, CardTitle, CardText, Row, Col } from 'reactstrap';
import PropTypes from 'prop-types';
import classnames from 'classnames';
import { AppHeaderDropdown, AppSidebarToggler } from '@coreui/react';
import { AppSidebarForm } from '@coreui/react';
import Select from 'react-select';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import Avatar from 'react-avatar';

import { 
  changeProject, 
  createProject, 
  fetchProjectEvents, 
  fetchProjectSettings,
} from '../../actions/projectsActions';
import { signout } from '../../actions/agentActions';
import factorsai from '../../common/factorsaiObj';

import JsSdk from '../../views/settings/JsSdk';
import AndroidSdk from '../../views/settings/AndroidSdk';
import Segment from '../../views/settings/Segment';
import NoContent from '../../common/NoContent';

const propTypes = {
  children: PropTypes.node,
};
const defaultProps = {};

const EVENT_POLL_LIMIT = 20;
const EVENT_POLL_INTERVAL = 30000; // 30sec

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
    eventNames: store.projects.currentProjectEventNames,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    changeProject, 
    createProject, 
    signout,
    fetchProjectEvents,
    fetchProjectSettings,
  }, dispatch);
}

class DefaultHeader extends Component {
  constructor(props) {
    super(props);

    this.state = {
      createProject: {
        allow: true,
        projectName : '',
      },
      
      showAddProjectModal: false,
      addProjectMessage: null,
      eventNamePollStarted: false,
      changeToLastSeen: true,
      activeTab: '1',
      loggedOut: false,
    }
  }

  projectHasEvents() {
    return this.props.eventNames && this.props.eventNames.length > 0
  }

  triggerSetupProjectIfRequired = () => {
    this.props.fetchProjectEvents(this.props.selectedProject.value)
      .then((r) => {
        if (r.status == 404) {
          this.setState({ showAddProjectModal: true });

          let pollCount = 0;
          // start first event poll, if not started.
          if (!this.state.eventNamePollStarted) {
            this.setState({ eventNamePollStarted: true });
            setInterval(() => {
              if (!this.projectHasEvents() && !this.state.loggedOut && pollCount < EVENT_POLL_LIMIT) {
                this.props.fetchProjectEvents(this.props.selectedProject.value);
                pollCount++;
              }
            }, EVENT_POLL_INTERVAL)
          }
          
          this.props.showSetupProjectNotification(true);
        } else {
          this.props.showSetupProjectNotification(false);
        }

        if (this.projectHasEvents()) {
          factorsai.addUserProperties({"activationState": "seenProjectWithEvents"})
        }
      })
      .catch(console.debug);
  }

  componentWillMount() {
    this.triggerSetupProjectIfRequired();
  }

  componentDidUpdate(prevProps) {
    if (this.state.changeToLastSeen) {
      // change to project last seen by user if any.
      let lsProjectId = this.getLastSeenProject();
      if (lsProjectId && this.props.selectedProject && 
        this.props.selectedProject.value != lsProjectId && 
        this.hasProject(lsProjectId)) {
        
        this.props.changeProject(lsProjectId);
      }
    }
    
    // on initial project selected and when changed.
    if ((!prevProps.selectedProject && this.props.selectedProject) || 
      prevProps.selectedProject.value != this.props.selectedProject.value)
        this.triggerSetupProjectIfRequired();
  }

  handleProjectChange = (selectedProject) => {
    let projectId = selectedProject.value;

    // change the project and update the
    // project's settings on the store.
    this.props.changeProject(projectId);
    this.props.fetchProjectSettings(projectId);

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
    this.setState({ createProject: { projectName: name, allow: true } });
  }

  changeViewToAccountSettings = () => {
    this.props.changeViewToAccountSettings();
  }

  changeViewToUserProfile = () => {
    this.props.changeViewToUserProfile();
  }
  
  handleCreateProject = (e) => {
    e.preventDefault();

    let projectName = this.state.createProject.projectName;
    if(projectName == "") {
      this.showAddProjectMessage({success: false, message: 'Your project name cannot be empty'});
      return;
    }

    // disable create action.
    this.setState({ createProject: { allow: false } });

    this.props.createProject(projectName)
      .then((r) => {
        this.handleProjectChange({ label: r.data.name, value: r.data.id });
        this.showAddProjectMessage({ 
          success: true,
          message: 'Successfully created your project.'
        });
      })
      .catch((r) => this.showAddProjectMessage({ 
        success: false, 
        message: 'Failed to create your project. Please try again.'
      }));
  }

  handleLogout = () => {
    this.setState({ loggedOut: true });
    this.props.signout();
    factorsai.track('logout', { email: this.props.currentAgent.email });
  }

  toggleAddProjectModal = () => {
    this.props.closeSetupProjectModal();
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
    if (!this.props.billingAccount) return;

    let noOfNotifications = 0;
    let dropDownItems = [];
    if (this.props.accountPlan.code != "free" && !this.props.billingAccount.organization_name){
      dropDownItems.push(
        <DropdownItem key={1} onClick={this.changeViewToAccountSettings}>
        <span className="text-muted">Complete Billing Info</span>
        </DropdownItem>
      );
      noOfNotifications++;
    } else {
      dropDownItems.push(
        <DropdownItem disabled key={0}>
          <span className="text-muted">No messages here.</span>
        </DropdownItem>
      );
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

  getAddProjectMessageStyle() {
    if (this.state.addProjectMessage == null) return null;
    return {
      paddingLeft: '10px',
      color: this.state.addProjectMessage.success ? '#4dbd74' : '#d64541' 
    };
  }


  toggleTab(tab) {
    if (this.state.activeTab !== tab) {
      this.setState({
        activeTab: tab
      });
    }
  }

  renderSetupProjectModal() {
    return (
      <Modal isOpen={this.state.showAddProjectModal || this.props.showSetupProjectModal} toggle={this.toggleAddProjectModal} style={{ marginTop: "3rem", minWidth: "52rem" }}>
        <ModalHeader toggle={this.toggleAddProjectModal}>Setup your project (3 Steps)</ModalHeader>
        <ModalBody style={{ height: "40rem", padding: "10px 40px", overflow: "scroll" }}>
          <h5 style={{ margin: "15px 0", fontWeight: "500", fontSize: "15px", color: "#444" }}>
            <span className="fapp-rounded-tag">1</span> Select or create a project
          </h5>
          <Row>
            <Col md={5} className="fapp-select light" style={{ padding: "20px", paddingLeft: "30px" }}>
              { this.renderProjectSelector() }
            </Col>
            <Col md={1} style={{ padding: "23px", textAlign: "center" }}>
              <span style={{ textAlign: "center", fontSize: "18px", fontWeight: "600", color:" #777" }}> OR </span>
            </Col>
            <Col md={6} style={{ padding: "20px", paddingRight: "30px" }}>
              <Input style={{ display: "inline-block", width: "70%", marginRight: "10px", height: "40px", border: "1px solid #bbb" }} 
                type="text" placeholder="Your Project Name" onChange={this.handleProjectNameFormChange} />
              <Button outline color="primary" onClick={this.handleCreateProject} disabled={!this.state.createProject.allow}>Create</Button>
              <span style={this.getAddProjectMessageStyle()} hidden={this.state.addProjectMessage == null}>
                { this.getAddProjectMessage() }
              </span>
            </Col>
          </Row>

          <hr style={{ margin: "30px -40px" }} />
          <h5 style={{ margin: "25px 0", fontWeight: "500", fontSize: "15px", color: "#444" }}>
            <span className="fapp-rounded-tag">2</span> Integrate SDK
          </h5>
          <Nav tabs>
            <NavItem>
              <NavLink className={classnames({ active: this.state.activeTab === "1" })} onClick={() => this.toggleTab("1")}>
                Javascript
              </NavLink>
            </NavItem>
            <NavItem>
              <NavLink className={classnames({ active: this.state.activeTab === "2" })} onClick={() => this.toggleTab("2")}>
                Android
              </NavLink>
            </NavItem>
            <NavItem>
              <NavLink className={classnames({ active: this.state.activeTab === "3" })} onClick={() => this.toggleTab("3")}>
                Segment
              </NavLink>
            </NavItem>
          </Nav>
          <TabContent activeTab={this.state.activeTab}>
            <TabPane style={{ padding: "30px", paddingBottom: "0" }} tabId="1"><JsSdk cardOnly /></TabPane>
            <TabPane style={{ padding: "30px", paddingBottom: "0" }} tabId="2"><AndroidSdk cardOnly /></TabPane> 
            <TabPane style={{ padding: "30px", paddingBottom: "0" }} tabId="3"><Segment cardOnly /></TabPane>
          </TabContent>

          <hr style={{ margin: "30px -40px" }} />
          <h5 style={{ margin: "25px 0", fontWeight: "500", fontSize: "15px", color: "#444" }}>
            <span className="fapp-rounded-tag">3</span> Send us an event
          </h5>
          <div style={{ textAlign: "center", padding: "30px", paddingBottom: "65px" }}>
            <h4 className="fapp-gray">
              { this.projectHasEvents() ? "We've received your first event successfully." : "We're listening for your first event..." }
            </h4>
          </div>

          <div style={{ textAlign: "center", marginBottom: "20px"}}>
            { "Having some trouble? Let us fix it for you. " }
            <button className="fapp-small-button" 
            onClick={() => { factorsai.track("clicked_setup_talk_to_us"); window.fcWidget.open(); }}>
              Talk to us
            </button>
          </div>
        </ModalBody>
      </Modal>
    );
  }

  renderProjectSelector() {
    return (
      <Select
        options={this.props.selectableProjects} 
        value={this.props.selectedProject} 
        onChange={this.handleProjectChange} 
        styles={projectSelectStyles} 
        placeholder={"Select Project ..."}
        blurInputOnSelect={true}
      />
    );
  }

  render() {
    // eslint-disable-next-line
    const { children, ...attributes } = this.props;

    let selectProjectDropDown = !!this.props.selectableProjects ? this.renderProjectSelector() : null;
          
    return (
      <React.Fragment>
        <AppSidebarToggler className="d-lg-none" display="md" mobile />
        {/* <AppSidebarToggler className="d-md-down-none fapp-navbar-toggler" display="lg" /> */}
        <AppSidebarForm className="fapp-select light fapp-header-dropdown" style={{width: '40%'}}>
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
        { this.renderSetupProjectModal() }
      </React.Fragment>
    );
  }
}

DefaultHeader.propTypes = propTypes;
DefaultHeader.defaultProps = defaultProps;

export default connect(mapStateToProps, mapDispatchToProps)(DefaultHeader);
