import React, { useEffect } from 'react';
import { Layout } from 'antd';
import Sidebar from '../../components/Sidebar';
import CoreQuery from '../CoreQuery';
import ProjectSettings from '../Settings/ProjectSettings';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { HashRouter, Route, Switch } from 'react-router-dom';
import { fetchProjects, setActiveProject } from '../../reducers/global';

function AppLayout({ projects, fetchProjects }) {
  const { Content } = Layout;

  useEffect(() => {
    fetchProjects();
  }, []);

  console.log(projects);

  return (
    <Layout>
      <Sidebar />
      <Layout className="fa-content-container">
        <Content className="bg-white min-h-screen">
          <HashRouter>
            <Switch>
              <Route path="/settings/" component={ProjectSettings} />
              <Route path="/" name="Home" component={CoreQuery} />
            </Switch>
          </HashRouter>
        </Content>
      </Layout>
    </Layout>
  );
}

const mapStateToProps = (state) => ({
  projects: state.global.projects,
  active_project: state.global.active_project
});

const mapDispatchToProps = dispatch => bindActionCreators({
  fetchProjects,
  setActiveProject
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(AppLayout);
