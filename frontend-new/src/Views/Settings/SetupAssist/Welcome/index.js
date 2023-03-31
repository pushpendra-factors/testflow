import React, { useState, useEffect } from 'react';
import { Row, Col, Modal, Button, notification } from 'antd';
import { Text, SVG } from 'factorsComponents';
import { connect, useDispatch } from 'react-redux';
import { Link, useHistory } from 'react-router-dom';
import {
  getHubspotContact,
  fetchDemoProject,
  setActiveProject
} from 'Reducers/global';
import styles from './index.module.scss';
import Video from './Video';
import { meetLink } from '../../../../utils/hubspot';
import FaHeader from '../../../../components/FaHeader';
import OnBoard from './OnboardFlow';
import { TOGGLE_WEBSITE_VISITOR_IDENTIFICATION_MODAL } from 'Reducers/types';

const EachWelcomeCard = ({ onClick, title, description, type }) => {
  return (
    <div
      className={`${styles.first} ml-2 mr-2 ${styles['eachWelcomeCard']}`}
      onClick={onClick}
    >
      <Row
        style={{
          display: 'flex',
          justifyContent: 'center',
          flexWrap: 'initial',
          flexDirection: 'column'
        }}
      >
        <Col className={styles['img']}>
          {type === 1 ? (
            <SVG name='onboardtarget' />
          ) : (
            <SVG name='onboardsearch' />
          )}
        </Col>
        <Col justify={'center'} span={24} className={'mt-8'}>
          <Text
            type={'title'}
            level={5}
            align={'center'}
            weight={'bold'}
            extraClass={'m-0'}
          >
            {title}
          </Text>
          <Text
            type={'title'}
            level={7}
            weight={'regular'}
            align={'center'}
            color={'grey'}
            style={{ padding: '10px 0 0 0' }}
          >
            {description}
          </Text>
        </Col>
      </Row>
    </div>
  );
};
const Welcome = ({
  currentAgent,
  activeProject,
  getHubspotContact,
  fetchDemoProject,
  setActiveProject,
  projects
}) => {
  const [showModal, setShowModal] = useState(false);
  const [ownerID, setownerID] = useState();
  const dispatch = useDispatch();
  const history = useHistory();

  const handleRoute = () => {
    history.push('/project-setup');
  };
  const handleRoute1 = () => {
    // dispatch({ type: TOGGLE_WEBSITE_VISITOR_IDENTIFICATION_MODAL });
    history.push('/welcome/visitoridentification/1');
  };

  useEffect(() => {
    let email = currentAgent?.email;
    getHubspotContact(email)
      .then((res) => {
        setownerID(res?.data?.hubspot_owner_id);
      })
      .catch((err) => {
        console.log(err?.data?.error);
      });
  }, []);

  const switchProject = () => {
    fetchDemoProject().then((res) => {
      let id = res.data[0];
      let selectedProject = projects.filter((project) => project?.id === id);
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
      <div className={'m-0'}>
        <Row justify={'center'} className={'mt-24'}>
          <Col span={12}>
            <Text
              type={'title'}
              level={2}
              weight={'bold'}
              align={'center'}
              extraClass={'m-0 mt-16'}
            >
              What do you want to get started on first?
            </Text>
            <Text
              type={'title'}
              level={6}
              align={'center'}
              weight={'regular'}
              extraClass={'m-0 mt-2'}
              color={'grey'}
            >
              Don't worry, you can set up the rest at any time
            </Text>
          </Col>
        </Row>
        <Row justify={'center'} className={'mt-8'}>
          <Col span={15}>
            <Row className={'justify-center'}>
              <EachWelcomeCard
                title='Website visitor identification'
                description='Identify anonymous users and track high intent accounts'
                type={2}
                onClick={handleRoute1}
              />
              <EachWelcomeCard
                title='Analytics and Attribution'
                description='Make data-driven decisions and optimize marketing strategies'
                type={1}
                onClick={handleRoute}
              />

              {/* <div className={`${styles.first}`} onClick={() => {
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
                </div> */}
            </Row>
          </Col>
        </Row>
        {/* <OnBoard /> */}
        {/* <Row justify='center' className={'mt-12'}>
          <Col span={7}>
            <Text
              type={'title'}
              level={6}
              weight={'regular'}
              extraClass={'m-0 inline'}
              color={'grey-2'}
              style={{ userSelect: 'none' }}
            >
              Want a quick video tour first?
            </Text>
            <Button
              type={'text'}
              size={'large'}
              className={'inline ml-1 mb-4 ' + styles['playvideobtn']}
              onClick={() => setShowModal(true)}
            >
              <SVG name={'PlayButton'} size={25} />
              Play Video
            </Button>
          </Col>
        </Row> */}
      </div>
      {/* video modal */}
      {/* <Video showModal={showModal} setShowModal={setShowModal} /> */}
    </>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentAgent: state.agent.agent_details,
  projects: state.global.projects
});

export default connect(mapStateToProps, {
  getHubspotContact,
  fetchDemoProject,
  setActiveProject
})(Welcome);
