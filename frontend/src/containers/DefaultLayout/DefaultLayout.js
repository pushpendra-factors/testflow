import React, { Component } from 'react';
import { connect } from "react-redux";
import { bindActionCreators } from 'redux';
import { Redirect, Route, Switch } from 'react-router-dom';
import { Container, Button } from 'reactstrap';
import {
  AppHeader,
  AppSidebar,
  AppSidebarNav,
} from '@coreui/react';

import InternalRoute from '../../routes.internal';

// sidebar nav config
import {sideBarItems,  internalSideBarItems } from '../../_nav';
// routes config
import {routes, internalRoutes}  from '../../routes';
import DefaultHeader from './DefaultHeader';
import {
  fetchProjects,
  fetchProjectModels
} from "../../actions/projectsActions";
import Loading from '../../loading';
import factorsicon from '../../assets/img/brand/factors-icon.svg';
import { fetchAgentInfo, fetchAgentBillingAccount } from '../../actions/agentActions';
import { hotjar } from 'react-hotjar';
import { isProduction, isFromFactorsDomain } from '../../util';

// inits factorsai sdk for app.
import factorsai from '../../common/factorsaiObj';

const isHotJarExcludedEmail = (email) => {
  const excludedEmailDomains = [ "factors.ai" ];
  return email && excludedEmailDomains.indexOf(email.split("@")[1]) > -1;
}

const mapStateToProps = store => {
  return {
    currentProjectId: store.projects.currentProjectId,
    projects: store.projects.projects,
    isAgentLoggedIn: store.agents.isLoggedIn,
    agent: store.agents.agent,
    eventNames: store.projects.currentProjectEventNames,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchProjects,
    fetchProjectModels,
    fetchAgentInfo,
    fetchAgentBillingAccount
  }, dispatch);
}

class DefaultLayout extends Component {
  constructor(props) {
    super(props);
    this.state = {
      projects: {
        loaded: false,
        error: null
      },
      agent: {
        loaded: false,
        error: null
      },
      showSetupProjectNotification: false,
      showSetupProjectModal: false,
    }
  }

  componentWillMount() {
    if (window.fcWidget) {
      window.fcWidget.init({
        token: "3208785c-3624-47c7-be9a-4f60aa0e60f9",
        host: "https://wchat.freshchat.com"
      });
    }

    this.props.fetchProjects()
      .then((action) => {
        this.setState({ 
          projects: {
            loaded: true
          }
        });
      })
      .catch((response) => {
        this.setState({ 
          projects: { 
            loaded: true, 
            error: response.payload
          } 
        });
      });

    this.props.fetchAgentInfo()
      .then((r) => {
        this.setState({
          agent: {
            loaded: true
          }
        });

        factorsai.identify(r.data.email);
        factorsai.addUserProperties({
          "email": r.data.email,
          "firstName": r.data.first_name,
          "lastName": r.data.last_name,
        });

        if (window.fcWidget) {
          window.fcWidget.setExternalId(r.data.email);
          window.fcWidget.user.setEmail(r.data.email);
          window.fcWidget.user.setFirstName(r.data.first_name);
        }
      })
      .catch((r) => {
        this.setState({ 
          agent: { 
            loaded: true,
            error: 'Failed to get agent information'
          } 
        });
      });

    this.props.fetchAgentBillingAccount();    
  }

  componentDidUpdate(prevProps, prevState) {
    if (prevProps.agent && this.props.agent && prevProps.agent.email != this.props.agent.email) {
      if (isProduction() && !isHotJarExcludedEmail(this.props.agent.email))
        hotjar.initialize(1259925, 6);
    }

    if (window.fcWidget && this.props.currentProjectId && this.props.projects) {
      window.fcWidget.user.setProperties({
        "Project Id": this.props.currentProjectId,
        "Project Name": this.props.projects[this.props.currentProjectId].name,
      });

      factorsai.addUserProperties({
        projectId: this.props.currentProjectId,
      });
    }
  }

  refresh = () => {
    this.props.history.push('/refresh');
  }

  changeViewToAccountSettings = () => {
    this.props.history.push('/account_settings');
  }

  changeViewToUserProfile = () => {
    this.props.history.push('/profile');
  }

  isLoaded() {
    return this.state.projects.loaded && this.state.agent.loaded;
  }

  isAgentLoggedIn(){
    return this.props.isAgentLoggedIn;
  }

  getAgentName = () => {
    return (this.props.agent 
      && this.props.agent.first_name) ? this.props.agent.first_name : '';
  }

  renderProjectsDropdown() {    
    const selectableProjects = Array.from(
      Object.values(this.props.projects), 
      // selectable_projects object structure.
      project => ({ "label": project.name, "value": project.id }) 
    )

    if (selectableProjects.length == 0 ){
      return <DefaultHeader 
        refresh={this.refresh}
        changeViewToAccountSettings={this.changeViewToAccountSettings}
        changeViewToUserProfile={this.changeViewToUserProfile}
        getProfileName={this.getAgentName} 
        currentAgent={this.props.agent}
        showSetupProjectNotification={this.showSetupProjectNotification}
        showSetupProjectModal={this.state.showSetupProjectModal}
        closeSetupProjectModal={this.closeSetupProjectModal}
      />
    }

    return <DefaultHeader
      refresh={this.refresh}
      changeViewToAccountSettings={this.changeViewToAccountSettings}
      changeViewToUserProfile={this.changeViewToUserProfile}
      selectableProjects={selectableProjects}
      selectedProject={{
        label: this.props.projects[this.props.currentProjectId].name,
        value: this.props.currentProjectId 
      }}
      getProfileName={this.getAgentName}
      currentAgent={this.props.agent}
      showSetupProjectNotification={this.showSetupProjectNotification}
      showSetupProjectModal={this.state.showSetupProjectModal}
      closeSetupProjectModal={this.closeSetupProjectModal}
    />
  }

  showInternalSideBarItems(){
    return isFromFactorsDomain(this.props.agent.email);
  }

  renderSetupProjectNotification() {
    if (!this.state.showSetupProjectNotification) return null;
  
    return (
      <div style={{ width: '105%', textAlign: 'center', background: '#6610f2', margin: '0 -30px', fontSize: '13px', fontWeight: '700', background: '#6f42c1', color: '#FFF' }}>
        Please complete setting up your project to use FactorsAI.
        <Button color='warning' style={{ margin: '5px', marginLeft: '10px', padding: '1px 7px', fontSize: '12px' }} 
          onClick={() => { this.toggleSetupProjectModal(); factorsai.track('clicked_complete_project_setup') }}>Setup Project</Button>
      </div>
    );
  }

  showSetupProjectNotification = (value) => {
    this.setState({ showSetupProjectNotification: value });
  }

  toggleSetupProjectModal = () => {
    this.setState({ showSetupProjectModal: !this.state.showSetupProjectModal });
  }

  closeSetupProjectModal = () => {
    this.setState({ showSetupProjectModal: false });
  }

  render() {
    if (!this.isAgentLoggedIn()) return <Redirect to='/login' />;
    if (!this.isLoaded()) return <Loading />;

    if (this.state.projects.loaded && this.state.projects.error) 
      return <div>Failed loading your project.</div>;

    let sideBarItemsToDisplay = {items: sideBarItems};

    if(this.showInternalSideBarItems()){
      sideBarItemsToDisplay.items = [...sideBarItems , ...internalSideBarItems]
    }

    return (
      <div className="app">
        <div className="app-body fapp-body">
          <AppSidebar minimized className="fapp-sidebar" fixed display="lg">
            <img style={{marginTop: '12px', marginBottom: '20px'}} src={factorsicon} />
            <AppSidebarNav navConfig={sideBarItemsToDisplay} {...this.props} />
          </AppSidebar>
          <main className="main fapp-main">
            <AppHeader className="fapp-header" fixed>
              { this.renderProjectsDropdown() }
            </AppHeader>
            <Container className='fapp-right-pane' fluid>
              { this.renderSetupProjectNotification() }
              <Switch>
                {
                  routes.map((route, idx) => {
                    return route.component ? (<Route key={idx} path={route.path} exact={route.exact} name={route.name} 
                      render={props => (<route.component {...props} />)} />) : (null);
                  })
                }
                {
                  internalRoutes.map((route, idx) => {
                    return (<InternalRoute key={idx} path={route.path} exact={route.exact} name={route.name} component={route.component}/>)
                  })
                }
                <Redirect from="/" to="/dashboard" />
              </Switch>
            </Container>
          </main>
        </div>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(DefaultLayout);
