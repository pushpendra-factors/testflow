import React from 'react'; 
import { Row, Col, Layout } from 'antd';
import Sidebar from '../../components/Sidebar';
import CoreQuery from '../CoreQuery'; 
import HeaderComp from './Header'; 
import { HashRouter, Route, Switch } from 'react-router-dom';

function AppLayout() {
    const { Header, Content } = Layout;

    return (
        <Layout>
            <Sidebar />
            <Layout className="fa-content-container">
                <HeaderComp /> 
                <Content className="px-8 bg-white min-h-screen" style={{ marginTop: 64 }}>    
                    <HashRouter>
                            <Switch>
                                <Route path="/" name="Home" component={CoreQuery} />  
                            </Switch> 
                    </HashRouter>  
                </Content>
            </Layout>
        </Layout>
    )
}

export default AppLayout;