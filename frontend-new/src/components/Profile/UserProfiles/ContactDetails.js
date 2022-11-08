import React, { useEffect, useState } from 'react';
import {
  Row,
  Col,
  Button,
  Avatar,
  Menu,
  Dropdown,
  Popover,
  Tabs
} from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { SVG, Text } from '../../factorsComponents';
import FaTimeline from '../../FaTimeline';
import { formatDurationIntoString } from '../../../utils/dataFormatter';
import { granularityOptions } from '../utils';
import {
  udpateProjectSettings,
  fetchProjectSettings
} from '../../../reducers/global';
import { getProfileUserDetails } from '../../../reducers/timelines/middleware';
import { getActivitiesWithEnableKeyConfig } from '../../../reducers/timelines/utils';
import SearchCheckList from '../../SearchCheckList';

function ContactDetails({
  onCancel,
  userDetails,
  activeProject,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings
}) {
  const [activities, setActivities] = useState([]);
  const [granularity, setGranularity] = useState('Daily');
  const [collapse, setCollapse] = useState(true);

  useEffect(() => {
    fetchProjectSettings(activeProject.id);
  }, [activeProject]);

  useEffect(() => {
    const listActivities = getActivitiesWithEnableKeyConfig(
      userDetails?.data?.user_activities,
      currentProjectSettings.timelines_config?.disabled_events
    );
    setActivities(listActivities);
  }, [currentProjectSettings, userDetails]);

  const granularityMenu = (
    <Menu>
      {granularityOptions.map((option) => (
        <Menu.Item key={option} onClick={(key) => setGranularity(key.key)}>
          <div className="flex items-center">
            <span className="mr-3">{option}</span>
          </div>
        </Menu.Item>
      ))}
    </Menu>
  );

  const handleChange = (option) => {
    const timelinesConfig = { ...currentProjectSettings.timelines_config };
    if (!timelinesConfig.disabled_events) {
      timelinesConfig.disabled_events = [];
    }
    if (option.enabled) {
      timelinesConfig.disabled_events.push(option.display_name);
    } else if (!option.enabled) {
      timelinesConfig.disabled_events.splice(
        timelinesConfig.disabled_events.indexOf(option.display_name),
        1
      );
    }
    udpateProjectSettings(activeProject.id, {
      timelines_config: { ...timelinesConfig }
    });
  };

  const controlsPopover = () => (
    <Tabs defaultActiveKey="events" size="small">
      <Tabs.TabPane
        tab={<span className="fa-activity-filter--tabname">Events</span>}
        key="events"
      >
        <SearchCheckList
          placeholder="Search Events"
          mapArray={activities}
          titleKey="display_name"
          checkedKey="enabled"
          onChange={handleChange}
        />
      </Tabs.TabPane>
    </Tabs>
  );

  return (
    <>
      <div
        className="fa-modal--header px-8"
        style={{ borderBottom: '1px solid #e7e9ed' }}
      >
        <Row justify="space-between" className="my-3 m-0">
          <Col className="flex items-center">
            <Button
              style={{ padding: '0' }}
              type="text"
              icon={<SVG name="brand" size={36} />}
              size="large"
              onClick={() => {
                onCancel();
                setCollapse(true);
                setGranularity('Daily');
              }}
            />
            <Text type="title" level={4} weight="bold" extraClass="m-0">
              Contact Details
            </Text>
          </Col>
          <Col>
            <Button
              size="large"
              type="text"
              onClick={() => {
                onCancel();
                setCollapse(true);
                setGranularity('Daily');
              }}
              icon={<SVG name="times" />}
            />
          </Col>
        </Row>
      </div>

      <div className="my-16">
        <Row span={24} gutter={[24, 24]}>
          <Col span={5} style={{ borderRight: '1px solid #e7e9ed' }}>
            <div className="ml-12 my-12">
              <Row className="">
                <Col>
                  <Avatar
                    size={72}
                    style={{
                      color: '#3E516C',
                      backgroundColor: '#F1F1F1',
                      display: 'flex',
                      alignItems: 'center'
                    }}
                  >
                    <SVG name="user" size={40} />
                  </Avatar>
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={6} extraClass="m-0" weight="bold">
                    {userDetails.data?.title}
                  </Text>
                  <Text type="title" level={7} extraClass="m-0" color="grey">
                    {userDetails.data?.subtitle}
                  </Text>
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={7} extraClass="m-0" color="grey">
                    Email
                  </Text>

                  <Text type="title" level={7} extraClass="m-0">
                    {userDetails?.data?.email || '-'}
                  </Text>
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={7} extraClass="m-0" color="grey">
                    Country
                  </Text>
                  <Text type="title" level={7} extraClass="m-0">
                    {userDetails?.data?.country || '-'}
                  </Text>
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={7} extraClass="m-0" color="grey">
                    Number of Web Sessions
                  </Text>
                  <Text type="title" level={7} extraClass="m-0">
                    {parseInt(userDetails.data?.web_session_count) || '-'}
                  </Text>
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={7} extraClass="m-0" color="grey">
                    Number of Page Views
                  </Text>
                  <Text type="title" level={7} extraClass="m-0">
                    {parseInt(userDetails.data?.number_of_page_views) || '-'}
                  </Text>
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={7} extraClass="m-0" color="grey">
                    Time Spent on Site
                  </Text>
                  <Text type="title" level={7} extraClass="m-0">
                    {formatDurationIntoString(
                      userDetails.data?.time_spent_on_site
                    )}
                  </Text>
                </Col>
              </Row>
              <Row
                className="mt-3 pt-3"
                style={{ borderTop: '1px dashed #e7e9ed' }}
              >
                <Col>
                  <Text
                    type="title"
                    level={7}
                    extraClass="m-0 my-2"
                    color="grey"
                  >
                    Associated Groups:
                  </Text>
                  {userDetails?.data?.group_infos?.map((group) => (
                    <Text type="title" level={7} extraClass="m-0 mb-2">
                      {group.group_name}
                    </Text>
                  )) || '-'}
                </Col>
              </Row>
              <Row className="mt-6">
                <Col className="flex justify-start items-center" />
              </Row>
            </div>
          </Col>
          <Col span={18}>
            <Row gutter={[24, 24]} justify="left">
              <Col span={24} className="mx-8 my-12">
                <Col className="flex items-center justify-between mb-4">
                  <div>
                    <Text type="title" level={3} weight="bold">
                      Timeline
                    </Text>
                  </div>
                  <div className="flex justify-between">
                    <div className="flex justify-between">
                      <Button
                        className="fa-dd--custom-btn"
                        type="text"
                        onClick={() => setCollapse(false)}
                      >
                        <SVG name="line_height" size={22} />
                      </Button>
                      <Button
                        className="fa-dd--custom-btn"
                        type="text"
                        onClick={() => setCollapse(true)}
                      >
                        <SVG name="grip_lines" size={22} />
                      </Button>
                    </div>
                    <div>
                      <Popover
                        overlayClassName="fa-activity--filter"
                        placement="bottomLeft"
                        trigger="hover"
                        content={controlsPopover}
                      >
                        <Button
                          size="large"
                          className="fa-btn--custom mx-2 relative"
                          type="text"
                        >
                          <SVG name="activity_filter" />
                        </Button>
                      </Popover>
                    </div>
                    <div>
                      <Dropdown
                        overlay={granularityMenu}
                        placement="bottomRight"
                      >
                        <Button className="ant-dropdown-link flex items-center">
                          {granularity}
                          <SVG name="caretDown" size={16} extraClass="ml-1" />
                        </Button>
                      </Dropdown>
                    </div>
                  </div>
                </Col>
                <Col span={24}>
                  <FaTimeline
                    activities={activities?.filter(
                      (activity) => activity.enabled === true
                    )}
                    loading={userDetails.isLoading}
                    granularity={granularity}
                    collapse={collapse}
                    setCollapse={setCollapse}
                  />
                </Col>
              </Col>
            </Row>
          </Col>
        </Row>
      </div>
    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  userDetails: state.timelines.contactDetails
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileUserDetails,
      fetchProjectSettings,
      udpateProjectSettings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(ContactDetails);
