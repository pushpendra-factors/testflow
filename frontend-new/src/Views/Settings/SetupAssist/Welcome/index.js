import React, { useState, useEffect } from 'react';
import {
    Row, Col, Modal, Button, notification
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { getHubspotContact, fetchDemoProject, setActiveProject } from 'Reducers/global';
import styles from './index.module.scss';
import Video from './Video';
import { meetLink } from '../../../../utils/hubspot';
import FaHeader from '../../../../components/FaHeader';

const Welcome = ({currentAgent, activeProject, getHubspotContact, fetchDemoProject, setActiveProject, projects}) => {
    const [showModal,setShowModal] = useState(false);
    const [ownerID, setownerID] = useState();
    const history = useHistory();

    const handleRoute = () => {
        history.push('/project-setup');
    }

    useEffect(() => {
        let email = currentAgent?.email;
        getHubspotContact(email).then((res) => {
            setownerID(res?.data?.hubspot_owner_id)
        }).catch((err) => {
            console.log(err?.data?.error)
        });
    }, []);

    const switchProject = () => {
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
    };

    return (
      <>
        <div className={"m-0"}>
          <Row justify={"center"} className={"mt-24"}>
            <Col span={12}>
              <Text type={"title"} level={2} weight={"bold"} align={'center'} extraClass={"m-0 mt-16"}>
                Hey {currentAgent?.first_name ? currentAgent?.first_name : 'there'}, welcome to Factors
              </Text>
              <Text type={"title"} level={6} align={'center'} weight={"regular"} extraClass={"m-0 mt-2"} color={"grey"}>
                What are you looking to do next?
              </Text>
            </Col>
          </Row>
          <Row justify={"center"} className={"mt-8"}>
            <Col span={15}>
              <Row className={"justify-between"}>
                <div className={`${styles.first}`} onClick={handleRoute}>
                  <Row>
                    <Col className={`${styles.img}`}>
                      <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/computer.svg' />
                    </Col>
                    <Col justify={'center'} span={24} className={'mt-24'}>
                      <Text type={"title"} level={5} align={'center'} weight={"bold"} extraClass={"m-0"}>
                        Start implementing
                      </Text>
                      <Text
                        type={"title"}
                        level={6}
                        weight={"regular"}
                        align={'center'}
                        color={"grey"}
                      >
                        Approximated time ~15 min
                      </Text>
                    </Col>
                  </Row>
                </div>
                <div className={`${styles.first}`} onClick={() => switchProject}>
                  <Row>
                    <Col className={`${styles.img}`}>
                      <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/file.svg' />
                    </Col>
                    <Col justify={'center'} span={24} className={'mt-24'}>
                      <Text type={"title"} align={'center'} level={5} weight={"bold"} extraClass={"m-0"}>
                        Explore demo
                      </Text>
                      <Text
                        type={"title"}
                        level={6}
                        align={'center'}
                        weight={"regular"}
                        extraClass={"m-0"}
                        color={"grey"}
                      >
                        Jump into a sample project
                      </Text>
                    </Col>
                  </Row>
                </div>
                <div className={`${styles.first}`} onClick={() => {
                  window.open(meetLink(ownerID), '_blank');
                }}>
                  <Row>
                    <Col className={`${styles.img}`}>
                      <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/callrep.svg' />
                    </Col>
                    <Col justify={'center'} span={24} className={'mt-24'}>
                      <Text
                        type={"title"}
                        level={5}
                        align={'center'}
                        weight={"bold"}
                        extraClass={"m-0"}
                      >
                        Need help?
                      </Text>
                      <Text
                        type={"title"}
                        level={7}
                        align={'center'}
                        extraClass={"m-0"}
                        style={{ opacity: 0.7 }}
                      >
                        Get help for setup and more
                      </Text>
                    </Col>
                  </Row>
                </div>
              </Row>
            </Col>
          </Row>
          <Row justify='center' className={'mt-12'}>
            <Col span={7}>
              <Text type={"title"} level={6} weight={"regular"} extraClass={"m-0 inline"} color={"grey-2"}>
                Want a quick video tour first?
              </Text>
              <Button
                type={"text"}
                size={"large"}
                className={"inline ml-1 mb-4"}
                onClick={() => setShowModal(true)}
              >
                <SVG name={"PlayButton"} size={25} />
                Play Video
              </Button>
            </Col>
          </Row>
        </div>
        {/* video modal */}
        <Video showModal={showModal} setShowModal={setShowModal} />
      </>
    );
}
           

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentAgent: state.agent.agent_details,
    projects: state.global.projects
});

export default connect(mapStateToProps, { getHubspotContact, fetchDemoProject, setActiveProject })(Welcome);
