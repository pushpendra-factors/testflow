import React, { useState, useEffect } from 'react';
import {
    Row, Col, Modal, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { fetchDemoProject, setActiveProject } from 'Reducers/global';
import styles from './index.module.scss';

const ExpalinBeforeIntegration = ({currentAgent, setActiveProject, fetchDemoProject, projects}) => {
    const history = useHistory();
    const [loading, setLoading] = useState(false);

    const handleRoute = () => {
        history.push('/project-setup');
    }

    const switchProject = () => {
        setLoading(true);
        fetchDemoProject().then((res) => {
            let id = res.data[0];
            let selectedProject = projects.filter(project => project.id === id);
            selectedProject = selectedProject[0];
            localStorage.setItem('activeProject', selectedProject?.id);
            setActiveProject(selectedProject);
            history.push('/');
            notification.success({
                message: 'Project Changed!',
                description: `You are currently viewing data from ${selectedProject.name}`
            });
        });
        setLoading(false);
    };

    return (
      <>
        <div className={"m-0"}>
            <Row justify={"center"} className={'mt-24'}>
                <Col span={15}>
                    <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/explain-img.svg' />
                </Col>
            </Row>
            <Row justify={"center"} className={"mt-8"}>
                <Col span={15}>
                    <Text type={"title"} level={2} weight={"bold"} extraClass={"m-0"}>
                        Explain
                    </Text>
                    <Text type={"title"} level={6} weight={"regular"} extraClass={"m-0 w-11/12"} color={"grey"}>
                        Find yourself asking how you can improve conversions between any two milestones of a user journey? The Explain module helps you investigate the impact of various user segments and their behaviors on your marketing efforts.
                    </Text>
                    <Text type={"title"} level={6} weight={"regular"} extraClass={"m-0 w-11/12 mt-2"} color={"grey"}>
                        To access this, please complete your setup by connecting to your data sources.
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
                                            extraClass={"m-0 w-9/12"}
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

export default connect(mapStateToProps, { fetchDemoProject, setActiveProject })(ExpalinBeforeIntegration);
