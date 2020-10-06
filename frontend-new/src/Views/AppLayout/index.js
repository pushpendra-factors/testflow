import React, { useEffect } from 'react';
import { Layout } from 'antd';
import Sidebar from '../../components/Sidebar';
import CoreQuery from '../CoreQuery';
import ProjectSettings from '../Settings/ProjectSettings';
import componentsLib from '../../Views/componentsLib';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  HashRouter, Route, Switch, useHistory
} from 'react-router-dom';
import { fetchProjects, setActiveProject } from '../../reducers/global';

function AppLayout({ fetchProjects, isAgentLoggedIn }) {
  // const [dataLoading, setDataLoading] = useState(true);
  const { Content } = Layout;
  const history = useHistory();

  if (!isAgentLoggedIn) {
    history.push('/login');
    return null;
  }

  useEffect(() => {
    fetchProjects();
    // setTimeout(() => {
    //   setDataLoading(false);
    // }, 500);
  }, [fetchProjects]);

  return (
  <Layout>
      <Sidebar />
      <Layout className="fa-content-container">
        <Content className="bg-white min-h-screen">
          <HashRouter>
            <Switch>
              <Route path="/components/" name="componentsLib" component={componentsLib} />
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
  active_project: state.global.active_project,
  isAgentLoggedIn: state.agent.isLoggedIn
});

const mapDispatchToProps = dispatch => bindActionCreators({
  fetchProjects,
  setActiveProject
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(AppLayout);
