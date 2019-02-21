import React, { Component } from 'react';
import { connect } from "react-redux";
import { bindActionCreators } from 'redux';
import { Redirect, Route, Switch } from 'react-router-dom';
import { Container, Button } from 'reactstrap';
import {
  AppFooter,
  AppHeader,
  AppSidebar,
  AppSidebarFooter,
  AppSidebarHeader,
  AppSidebarMinimizer,
  AppSidebarNav,
} from '@coreui/react';

// sidebar nav config
import navigation from '../../_nav';
// routes config
import routes from '../../routes';
import DefaultFooter from './DefaultFooter';
import DefaultHeader from './DefaultHeader';
import {
  fetchProjects,
  fetchProjectEvents, 
  fetchProjectSettings,
  fetchProjectModels
} from "../../actions/projectsActions";
import Loading from '../../loading';


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
    isAgentLoggedIn: store.agents.isLoggedIn
  }
}

const mapDispatchToProps = dispatch => {
  return bindActionCreators({ 
    fetchProjects,
    fetchProjectEvents, 
    fetchProjectSettings,
    fetchProjectModels
  }, dispatch);
}

class DefaultLayout extends Component {

  constructor(props) {
    super(props);
    this.state = {
      projects: {
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
  }

  refresh = () => {
    this.props.history.push('/refresh');
  }

  isLoaded() {
    return this.state.projects.loaded;
  }

  isAgentLoggedIn(){
    return this.props.isAgentLoggedIn
  }

  renderProjectsDropdown(){    
    const selectableProjects = Array.from(
      Object.values(this.props.projects), 
      project => ({ "label": project.name, "value": project.id }) // selectable_projects object structure.
    )
    if (selectableProjects.length == 0 ){
      return <DefaultHeader refresh={this.refresh} /> 
    }
    return <DefaultHeader refresh={this.refresh} selectableProjects={selectableProjects}
      selectedProject={{ label: this.props.projects[this.props.currentProjectId].name,
      value: this.props.currentProjectId }} />
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
        <AppHeader className="fapp-header" fixed>
          {this.renderProjectsDropdown()} 
        </AppHeader>
        <div className="app-body">
          <AppSidebar className="fapp-sidebar" fixed display="lg">
            <AppSidebarHeader />
            <AppSidebarNav className="fapp-sidebar-nav" navConfig={navigation} {...this.props} />
            <AppSidebarFooter />
            <AppSidebarMinimizer />
          </AppSidebar>
          <main className="main fapp-main">
            <Container fluid>
              <Switch>
                {routes.map((route, idx) => {
                    return route.component ? (<Route key={idx} path={route.path} exact={route.exact} name={route.name} 
                      render={props => (<route.component {...props} />)} />) : (null);
                  },
                )}
                <Redirect from="/" to="/factor" />
              </Switch>
            </Container>
          </main>
        </div>
        <AppFooter hidden>
          <DefaultFooter />
        </AppFooter>
      </div>
    );
  }
}

export default connect(mapStateToProps, mapDispatchToProps)(DefaultLayout);
