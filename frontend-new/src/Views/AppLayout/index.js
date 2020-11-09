import React, { useEffect, useState, useCallback } from 'react';
import { Layout, Spin } from 'antd';
import Sidebar from '../../components/Sidebar';
import CoreQuery from '../CoreQuery';
import Dashboard from '../Dashboard';
import Factors from '../Factors';
import FactorsInsights from '../Factors/FactorsInsights';
import ProjectSettings from '../Settings/ProjectSettings';
import componentsLib from '../../Views/componentsLib';
import { connect, useSelector, useDispatch } from 'react-redux';
import { bindActionCreators } from 'redux';
import {
  HashRouter, Route, Switch, useHistory
} from 'react-router-dom';
import { fetchProjects } from 'Reducers/agentActions';
import { fetchQueries } from '../../reducers/coreQuery/services';
import { fetchDashboards } from '../../reducers/dashboard/services';

function AppLayout({ fetchProjects }) {
  const [dataLoading, setDataLoading] = useState(true);
  const { Content } = Layout;
  const history = useHistory();
  const agentState = useSelector(state => state.agent);
  const isAgentLoggedIn = agentState.isLoggedIn;
  const { active_project } = useSelector(state => state.global);
  const dispatch = useDispatch();

  const asyncCallOnLoad = useCallback(async () => {
    try {
      await fetchProjects();
      setDataLoading(false);
    } catch (err) {
      console.log(err);
    }
  }, [fetchProjects]);

  useEffect(() => {
    asyncCallOnLoad();
  }, [asyncCallOnLoad]);

  useEffect(() => {
    if (active_project.id) {
      fetchDashboards(dispatch, active_project.id);
      fetchQueries(dispatch, active_project.id);
    }
  }, [dispatch, active_project.id]);

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
                  <Route path="/factors/insights" name="Factors" component={FactorsInsights} />
                  <Route path="/factors" name="Factors" component={Factors} />
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

const mapDispatchToProps = dispatch => bindActionCreators({
  fetchProjects,
  fetchDashboards
}, dispatch);

export default connect(null, mapDispatchToProps)(AppLayout);
