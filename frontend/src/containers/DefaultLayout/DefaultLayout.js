import React, { Component } from 'react';
import { connect } from "react-redux";
import { bindActionCreators } from 'redux';
import { Redirect, Route, Switch } from 'react-router-dom';
import { Container } from 'reactstrap';
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
    projects: store.projects.projects
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

  fetchProjectDependencies  = (projectId) => {
    // Todo(Dinesh): Remove dependency reload from here. dispatch changeProject action to
    // re-render corresponding component which will call fetch on mount.
    this.props.fetchProjectSettings(projectId);
    this.props.fetchProjectEvents(projectId);
    this.props.fetchProjectModels(projectId);
  }

  render() {
    // Todo(Dinesh): Define a generic loading screen.
    if (!this.state.projects.loaded) return <div>Loading...</div>;

    if (this.state.projects.loaded && this.state.projects.error) 
      return <div>Failed loading your project.</div>;

    const selectableProjects = Array.from(
      Object.values(this.props.projects), 
      project => ({ "label": project.name, "value": project.id }) // selectable_projects object structure.
    )

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
