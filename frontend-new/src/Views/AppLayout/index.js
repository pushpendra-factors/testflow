import React, { useEffect, useState } from 'react';
import { Layout, Spin } from 'antd';
import Sidebar from '../../components/Sidebar';
import CoreQuery from '../CoreQuery';
import Dashboard from '../Dashboard';
import ProjectSettings from '../Settings/ProjectSettings';
import componentsLib from '../../Views/componentsLib';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  HashRouter, Route, Switch, useHistory
} from 'react-router-dom';
import { fetchProjects } from '../../reducers/global';

function AppLayout({ fetchProjects, isAgentLoggedIn }) {
  const [dataLoading, setDataLoading] = useState(true);
  const { Content } = Layout;
  const history = useHistory();

  useEffect(() => {
    fetchProjects().then(() => {
      setDataLoading(false);
    });
  }, [fetchProjects]);

  if (!isAgentLoggedIn) {
    history.push('/login');
    return null;
  }

  return (
    <>
    {dataLoading ? <Spin size={'large'} className={'fa-page-loader'} />
      : <Layout>
          <Sidebar />
          <Layout className="fa-content-container">
            <Content className="bg-white min-h-screen">
              <HashRouter>
                <Switch>
                  <Route path="/components/" name="componentsLib" component={componentsLib} />
                  <Route path="/settings/" component={ProjectSettings} />
                  <Route path="/core-analytics" name="Home" component={CoreQuery} />
                  <Route path="/" name="Home" component={Dashboard} />
                </Switch>
              </HashRouter>
            </Content>
          </Layout>
      </Layout>
    }
    </>
  );
}

const mapStateToProps = (state) => ({
  projects: state.global.projects,
  isAgentLoggedIn: state.agent.isLoggedIn
});

const mapDispatchToProps = dispatch => bindActionCreators({
  fetchProjects
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(AppLayout);
