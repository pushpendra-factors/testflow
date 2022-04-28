import React, { useState, useEffect } from 'react';
import {
    Row, Col, Modal, Button
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect } from 'react-redux';
import { useHistory } from 'react-router-dom';
import { getHubspotContact } from 'Reducers/global';
import styles from './index.module.scss';
import Video from './Video';
import { meetLink } from '../../../../utils/hubspot';
import FaHeader from '../../../../components/FaHeader';

const Welcome = ({currentAgent, activeProject, getHubspotContact}) => {
    const [showModal,setShowModal] = useState(false);
    const [ownerID, setownerID] = useState();
    const history = useHistory();

    const handleRoute = () => {
        history.push('/project-setup');
    }

    useEffect(() => {
        let email = currentAgent.email;
        getHubspotContact(email).then((res) => {
            setownerID(res.data.hubspot_owner_id)
        }).catch((err) => {
            console.log(err.data.error)
        });
    }, []);

    return (
      <>
        <div className={"m-0"}>
          <Row justify={"center"} className={"mt-24"}>
            <Col span={16}>
              <Text type={"title"} level={2} weight={"bold"} extraClass={"m-0"}>
                Hey! Welcome {currentAgent.first_name}
              </Text>
              <Text type={"title"} level={6} weight={"regular"} extraClass={"m-0"} color={"grey"}>
                With Factors, you will always know which marketing activites drive conversions.
              </Text>
              <Text type={"title"} level={6} weight={"regular"} extraClass={"m-0"} color={"grey"}>
                Use this information to scale channels that bring you success.
              </Text>
              <Text
                type={"title"}
                level={6}
                weight={"regular"}
                extraClass={"m-0 mt-2"}
                color={"grey"}
              >
                The first step to get up and running with Factors is to get data into your project:
              </Text>
            </Col>
          </Row>
          <Row justify={"center"} className={"mt-6"}>
            <Col span={16}>
              <Row className={"justify-between"}>
                <div className={`${styles.first}`}>
                  <Row>
                    <Col span={18}>
                      <Text type={"title"} level={5} weight={"bold"} extraClass={"m-0 ml-6 mt-4"}>
                        Connect to data sources
                      </Text>
                      <Text
                        type={"title"}
                        level={6}
                        weight={"regular"}
                        extraClass={"m-0 ml-6 mb-6"}
                        color={"grey"}
                      >
                        Approximated time ~15 minutes
                      </Text>
                      <Button
                        type={"primary"}
                        size={"large"}
                        className={"ml-6 mb-4"}
                        onClick={handleRoute}
                      >
                        Get started
                      </Button>
                    </Col>
                    <Col className={`${styles.img}`}>
                      <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/cuate.svg' />
                    </Col>
                  </Row>
                </div>
                <div className={`${styles.first}`}>
                  <Row>
                    <Col span={20}>
                      <Text type={"title"} level={5} weight={"bold"} extraClass={"m-0 ml-6 mt-4"}>
                        See how Factors works
                      </Text>
                      <Text
                        type={"title"}
                        level={6}
                        weight={"regular"}
                        extraClass={"m-0 ml-6 mb-2"}
                        color={"grey"}
                      >
                        Quickly learn about features, configuration, and more
                      </Text>
                      <Button
                        type={"text"}
                        size={"large"}
                        className={"ml-2 mb-4"}
                        onClick={() => setShowModal(true)}
                      >
                        <SVG name={"PlayButton"} size={25} />
                        Play video
                      </Button>
                    </Col>
                    <Col className={`${styles.img}`}>
                      <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/rafiki.svg' />
                    </Col>
                  </Row>
                </div>
              </Row>
            </Col>
          </Row>
          <Row justify={"center"}>
            <Col span={16} className={`${styles.second} mt-4`}>
              <Row>
                <Col span={14}>
                  <Text
                    type={"title"}
                    level={5}
                    weight={"bold"}
                    color={"white"}
                    extraClass={"m-0 ml-6 mt-4"}
                  >
                    Need help with setup?
                  </Text>
                  <Text
                    type={"title"}
                    level={7}
                    color={"white"}
                    extraClass={"m-0 ml-6 mb-4"}
                    style={{ opacity: 0.7 }}
                  >
                    To get the most out of the trial period with Factors, get a personalized demo
                    with our in-house product experts
                  </Text>
                  <a href={meetLink(ownerID)} target='_blank'>
                    <Button
                      type={"text"}
                      style={{ background: "bottom", color: "white" }}
                      className={"ml-4 mb-4"}
                    >
                      Let’s Chat
                      <SVG name={"Arrowright"} size={16} extraClass={"ml-1"} color={"white"} />
                    </Button>
                  </a>
                </Col>
                <Col className={`${styles.callimg}`}>
                  <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/1to1.svg' />
                </Col>
              </Row>
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
});

export default connect(mapStateToProps, { getHubspotContact })(Welcome);
