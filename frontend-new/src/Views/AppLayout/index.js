import React from 'react';
import { Row, Col, Layout } from 'antd';
import Sidebar from '../../components/Sidebar';
import CoreQuery from '../CoreQuery';
import HeaderComp from './Header';
import ProjectSettings from '../Settings/ProjectSettings';
import { HashRouter, Route, Switch } from 'react-router-dom';

function AppLayout() {
    const { Header, Content } = Layout;

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
    )
}

export default AppLayout;