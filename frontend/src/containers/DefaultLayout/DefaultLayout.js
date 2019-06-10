import React, { Component } from 'react';
import { connect } from "react-redux";
import { bindActionCreators } from 'redux';
import { Redirect, Route, Switch } from 'react-router-dom';
import { Container } from 'reactstrap';
import {
  AppHeader,
  AppSidebar,
  AppSidebarFooter,
  AppSidebarMinimizer,
  AppSidebarNav,
} from '@coreui/react';

// sidebar nav config
import navigation from '../../_nav';
// routes config
import routes from '../../routes';
import DefaultHeader from './DefaultHeader';
import {
  fetchProjects,
  fetchProjectEvents, 
  fetchProjectSettings,
  fetchProjectModels
} from "../../actions/projectsActions";
import Loading from '../../loading';
import factorsicon from '../../assets/img/brand/factors-icon.svg';
import { fetchAgentInfo, fetchAgentBillingAccount } from '../../actions/agentActions';
import { hotjar } from 'react-hotjar';
import { isProduction } from '../../util';

// inits factorsai sdk for app.
import factorsai from '../../common/factorsaiObj';

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
    currentProjectId: store.projects.currentProjectId,
    projects: store.projects.projects,
    isAgentLoggedIn: store.agents.isLoggedIn,
    agent: store.agents.agent,
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchProjects,
    fetchProjectEvents, 
    fetchProjectSettings,
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
      }
    }
  }

  componentWillMount() {
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

      if(isProduction()) {
        hotjar.initialize(1259925, 6);
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
    return this.props.isAgentLoggedIn
  }

  getAgentName = () => {
    return (this.props.agent 
      && this.props.agent.first_name) ? this.props.agent.first_name : '';
  }

  renderProjectsDropdown(){    
    const selectableProjects = Array.from(
      Object.values(this.props.projects), 
      // selectable_projects object structure.
      project => ({ "label": project.name, "value": project.id }) 
    )

    if (selectableProjects.length == 0 ){
      return <DefaultHeader 
        refresh={this.refresh}
        changeViewToAccountSettings= {this.changeViewToAccountSettings}
        changeViewToUserProfile = {this.changeViewToUserProfile}
        getProfileName={this.getAgentName} 
        agentEmail={this.props.agent.email} 
      /> 
    }

    return <DefaultHeader
      refresh={this.refresh}
      changeViewToAccountSettings= {this.changeViewToAccountSettings}
      changeViewToUserProfile = {this.changeViewToUserProfile}
      selectableProjects={selectableProjects}
      selectedProject={{
        label: this.props.projects[this.props.currentProjectId].name,
        value: this.props.currentProjectId 
      }}
      getProfileName={this.getAgentName}
      agentEmail={this.props.agent.email}
    />
  }

  render() {

    if (!this.isAgentLoggedIn()){
      return <Redirect to='/login' />
    }

    if (!this.isLoaded()) return <Loading />;

    if (this.state.projects.loaded && this.state.projects.error) 
      return <div>Failed loading your project.</div>;

    return (
      <div className="app">
        <div className="app-body fapp-body">
          <AppSidebar minimized className="fapp-sidebar" fixed display="lg">
            <img style={{marginTop: '12px', marginBottom: '20px'}} src={factorsicon} />
            <AppSidebarNav navConfig={navigation} {...this.props} />
          </AppSidebar>
          <main className="main fapp-main">
          <AppHeader className="fapp-header" fixed>
            {this.renderProjectsDropdown()} 
          </AppHeader>
            <Container className='fapp-right-pane' fluid>
              <Switch>
                {routes.map((route, idx) => {
                    return route.component ? (<Route key={idx} path={route.path} exact={route.exact} name={route.name} 
                      render={props => (<route.component {...props} />)} />) : (null);
                  },
                )}
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
