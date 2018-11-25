import React, { Component } from 'react';
import { connect } from "react-redux"

import { Redirect, Route, Switch } from 'react-router-dom';
import { Container } from 'reactstrap';

import {
  AppAside,
  AppBreadcrumb,
  AppFooter,
  AppHeader,
  AppSidebar,
  AppSidebarFooter,
  AppSidebarForm,
  AppSidebarHeader,
  AppSidebarMinimizer,
  AppSidebarNav,
} from '@coreui/react';
import Select from 'react-select';
// sidebar nav config
import navigation from '../../_nav';
// routes config
import routes from '../../routes';
import DefaultAside from './DefaultAside';
import DefaultFooter from './DefaultFooter';
import DefaultHeader from './DefaultHeader';
import { 
  fetchProjects, 
  fetchCurrentProjectEvents, 
  fetchCurrentProjectSettings 
} from "../../actions/projectsActions";


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

@connect((store) => {
  return {
    currentProjectId : store.projects.currentProjectId,
    projects : store.projects.projects
  };
})

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
    this.props.dispatch(fetchProjects())
      .then((response) => {
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

  /**
   * Loads all dependencies; 
   * */
  fetchProjectDependencies  = (projectId) => {
    // Independent methods. Not chained.
    this.props.dispatch(fetchCurrentProjectSettings(projectId));
    this.props.dispatch(fetchCurrentProjectEvents(projectId));
  }

  render() {
    // Todo(Dinesh): Define a generic loading..
    if (!this.state.projects.loaded) return <div>Loading...</div>;

    // Todo(Dinesh): Handle project fetch failure on frontend.
    if (this.state.projects.loaded && this.state.projects.error) return <div>Failed loading your project.</div>;

    const selectableProjects = Array.from(
      Object.values(this.props.projects), 
      project => ({ "label": project.name, "value": project.id }) // selectable_projects object structure.
    )

    this.fetchProjectDependencies(this.props.currentProjectId);

    return (
      <div className="app">
        <AppHeader className="fapp-header" fixed>
          <DefaultHeader 
            fetchProjectDependencies={this.fetchProjectDependencies} 
            selectableProjects={selectableProjects}
            selectedProject={{ label: this.props.projects[this.props.currentProjectId].name, value: this.props.currentProjectId }} 
          />
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
                <Redirect from="/" to="/dashboard" />
              </Switch>
            </Container>
          </main>
          <AppAside fixed hidden>
            <DefaultAside />
          </AppAside>
        </div>
        <AppFooter>
          <DefaultFooter />
        </AppFooter>
      </div>
    );
  }
}

export default DefaultLayout;
