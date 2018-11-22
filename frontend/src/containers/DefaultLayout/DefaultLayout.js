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
    projects : store.projects.projects,
    currentProject : store.projects.currentProject,
  };
})

class DefaultLayout extends Component {
  componentWillMount() {
    this.props.dispatch(fetchProjects())
  }

  /**
   * Loads all dependencies; 
   * */
  fetchProjectDependencies  = (project) => {
    // Independent methods. Not chained.
    this.props.dispatch(fetchCurrentProjectSettings(project));
    this.props.dispatch(fetchCurrentProjectEvents(project));
  }

  handleChange = (selectedProject) => {
    this.fetchProjectDependencies(selectedProject);
  }

  render() {
    const mappedProjects = Array.from(this.props.projects,
       project => ({"label": project.name, "value": project.id}))
    if (!this.props.currentProject && mappedProjects.length > 0) {
      // Default select first project.    
      this.fetchProjectDependencies(mappedProjects[0]);
    }

    return (
      <div className="app">
        <AppHeader fixed>
          <DefaultHeader />
        </AppHeader>
        <div className="app-body">
          <AppSidebar fixed display="lg">
            <AppSidebarHeader />
            <AppSidebarForm>
              <Select
                options={mappedProjects}
                value={this.props.currentProject}
                onChange={this.handleChange}
                styles={projectSelectStyles}
                placeholder={"Select Project ..."}
                blurInputOnSelect={true}
                />
            </AppSidebarForm>
            <AppSidebarNav navConfig={navigation} {...this.props} />
            <AppSidebarFooter />
            <AppSidebarMinimizer />
          </AppSidebar>
          <main className="main">
            <AppBreadcrumb appRoutes={routes}/>
            <Container fluid>
              <Switch>
                {routes.map((route, idx) => {
                    return route.component ? (<Route key={idx} path={route.path} exact={route.exact} name={route.name} render={props => (
                        <route.component {...props} />
                      )} />)
                      : (null);
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
