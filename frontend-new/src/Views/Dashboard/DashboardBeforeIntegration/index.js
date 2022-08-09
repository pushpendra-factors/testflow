import React, { useState, useEffect } from 'react';
import {
    Row, Col, Modal, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { fetchDemoProject, setActiveProject } from 'Reducers/global';
import styles from './index.module.scss';

const DashboardBeforeIntegration = ({currentAgent, setActiveProject, fetchDemoProject, projects}) => {
    const history = useHistory();
    const [loading, setLoading] = useState(false);

    const handleRoute = () => {
        history.push('/project-setup');
    }

    const switchProject = () => {
        setLoading(true);
        fetchDemoProject().then((res) => {
            let id = res.data[0];
            let selectedProject = projects.filter(project => project?.id === id);
            selectedProject = selectedProject[0];
            localStorage.setItem('activeProject', selectedProject?.id);
            setActiveProject(selectedProject);
            history.push('/');
            notification.success({
                message: 'Project Changed!',
                description: `You are currently viewing data from demo project`
            });
        });
        setLoading(false);
    };

    return (
      <>
        <div className={"m-0"}>
            <Row justify={"center"} className={'mt-24'}>
                <Col span={15}>
                    <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/dashboard-img.svg' />
                </Col>
            </Row>
            <Row justify={"center"} className={"mt-4"}>
                <Col span={15}>
                    <Text type={"title"} level={2} weight={"bold"} extraClass={"m-0"}>
                        Dashboard
                    </Text>
                    <Text type={"title"} level={6} weight={"regular"} extraClass={"m-0 w-11/12"} color={"grey"}>
                        All your important metrics at a glance. The dashboard is where you save your analyses for quick and easy viewing. Create multiple dashboards for different needs, and toggle through them as you wish. Making the right decisions just became easier. <a href='https://help.factors.ai/en/articles/6294988-dashboards' target={'_blank'}>Learn more</a>
                    </Text>
                    {/* <Text type={"title"} level={6} weight={"regular"} extraClass={"m-0"} color={"grey"}>
                        Use this information to scale channels that bring you success.
                    </Text> */}
                    <Text
                        type={"title"}
                        level={6}
                        weight={"regular"}
                        extraClass={"m-0 mt-2"}
                        color={"grey"}
                    >
                        Finish connecting to your data sources - it's time to build your first dashboard!
                    </Text>
                </Col>
            </Row>
            <Row justify={"center"} className={"mt-6"}>
                <Col span={15}>
                    <Row className={"justify-between"}>
                        <div className={`${styles.first}`}>
                        <Row>
                            <Col span={18}>
                                <div className={'ml-6 mt-4 mb-6'}>
                                    <Text type={"title"} level={5} weight={"bold"} extraClass={"m-0"}>
                                        Complete Project Setup
                                    </Text>
                                    <Text
                                        type={"title"}
                                        level={6}
                                        weight={"regular"}
                                        extraClass={"m-0"}
                                        color={"grey"}
                                    >
                                        Approximated time ~6 minutes
                                    </Text>
                                </div>
                                <Button
                                    type={"primary"}
                                    size={"middle"}
                                    className={"ml-6 mb-4"}
                                    onClick={handleRoute}
                                >
                                    Finish Setup
                                </Button>
                            </Col>
                            <Col className={`${styles.img}`}>
                                <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/computer.svg' />
                            </Col>
                        </Row>
                        </div>
                        <div className={`${styles.first}`}>
                            <Row>
                                <Col span={20}>
                                    <div className={'ml-6 mt-4 mb-2'}>
                                        <Text type={"title"} level={5} weight={"bold"} extraClass={"m-0"}>
                                            Explore the Demo Project 
                                        </Text>
                                        <Text
                                            type={"title"}
                                            level={6}
                                            weight={"regular"}
                                            extraClass={"w-9/12"}
                                            color={"grey"}
                                        >
                                            A sample playground with sample datasets
                                        </Text>
                                    </div>
                                    <Button
                                        type={"default"}
                                        size={"middle"}
                                        className={"ml-6 mb-4"}
                                        loading={loading}
                                        onClick={switchProject}
                                    >
                                        {/* <SVG name={"PlayButton"} size={25} /> */}
                                        View Demo Project
                                    </Button>
                                </Col>
                                <Col className={`${styles.img}`}>
                                <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/file.svg' />
                                </Col>
                            </Row>
                        </div>
                    </Row>
                </Col>
            </Row>
        </div>
      </>
    );
}
           

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentAgent: state.agent.agent_details,
    projects: state.global.projects
});

export default connect(mapStateToProps, { fetchDemoProject, setActiveProject })(DashboardBeforeIntegration);
