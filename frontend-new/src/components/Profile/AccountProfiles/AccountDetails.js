import React, { useState, useEffect } from 'react';
import { Row, Col, Button, Dropdown, Menu, Popover, Tabs, Input } from 'antd';
import { bindActionCreators } from 'redux';
import { connect } from 'react-redux';
import { Text, SVG } from '../../factorsComponents';
import AccountTimeline from './AccountTimeline';
import { getHost, granularityOptions } from '../utils';
import {
  udpateProjectSettings,
  fetchProjectSettings
} from '../../../reducers/global';
import { getProfileAccountDetails } from '../../../reducers/timelines/middleware';
import { getActivitiesWithEnableKeyConfig } from '../../../reducers/timelines/utils';
import SearchCheckList from '../../SearchCheckList';

function AccountDetails({
  onCancel,
  accountDetails,
  activeProject,
  currentProjectSettings,
  fetchProjectSettings,
  udpateProjectSettings
}) {
  const [granularity, setGranularity] = useState('Daily');
  const [collapseAll, setCollapseAll] = useState(true);
  const [activities, setActivities] = useState([]);

  useEffect(() => {
    fetchProjectSettings(activeProject.id);
  }, [activeProject]);

  useEffect(() => {
    const listActivities = getActivitiesWithEnableKeyConfig(
      accountDetails.data?.account_events,
      currentProjectSettings.timelines_config?.disabled_events
    );
    setActivities(listActivities);
  }, [currentProjectSettings, accountDetails]);

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
                setGranularity('Daily');
                setActivities([]);
                setCollapseAll(true);
              }}
            />
            <Text type="title" level={4} weight="bold" extraClass="m-0">
              Account Details
            </Text>
          </Col>
          <Col>
            <Button
              size="large"
              type="text"
              onClick={() => {
                onCancel();
                setGranularity('Daily');
                setActivities([]);
                setCollapseAll(true);
              }}
              icon={<SVG name="times" />}
            />
          </Col>
        </Row>
      </div>

      <div className="mt-16">
        <Row span={24} gutter={[24, 24]}>
          <Col span={5} style={{ borderRight: '1px solid #e7e9ed' }}>
            <div className="ml-12 my-12">
              <Row className="">
                <Col>
                  <img
                    src={`https://logo.clearbit.com/${getHost(
                      accountDetails?.data?.host
                    )}`}
                    onError={(e) => {
                      if (
                        e.target.src !==
                        'https://s3.amazonaws.com/www.factors.ai/assets/img/buildings.svg'
                      ) {
                        e.target.src =
                          'https://s3.amazonaws.com/www.factors.ai/assets/img/buildings.svg';
                      }
                    }}
                    alt=""
                    height={72}
                    width={72}
                  />
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={6} extraClass="m-0" weight="bold">
                    {accountDetails?.data?.name}
                  </Text>
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={7} extraClass="m-0" color="grey">
                    Industry
                  </Text>

                  <Text type="title" level={7} extraClass="m-0">
                    {accountDetails?.data?.industry}
                  </Text>
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={7} extraClass="m-0" color="grey">
                    Country
                  </Text>
                  <Text type="title" level={7} extraClass="m-0">
                    {accountDetails?.data?.country}
                  </Text>
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={7} extraClass="m-0" color="grey">
                    Employee Size
                  </Text>
                  <Text type="title" level={7} extraClass="m-0">
                    {accountDetails?.data?.number_of_employees}
                  </Text>
                </Col>
              </Row>
              <Row className="py-2">
                <Col>
                  <Text type="title" level={7} extraClass="m-0" color="grey">
                    Number of Users
                  </Text>
                  <Text type="title" level={7} extraClass="m-0">
                    {parseInt(accountDetails?.data?.number_of_users) > 25
                      ? '25+'
                      : accountDetails?.data?.number_of_users}
                  </Text>
                </Col>
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
                        onClick={() => setCollapseAll(false)}
                      >
                        <SVG name="line_height" size={22} />
                      </Button>
                      <Button
                        className="fa-dd--custom-btn"
                        type="text"
                        onClick={() => setCollapseAll(true)}
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
                  <AccountTimeline
                    timelineEvents={
                      activities?.filter(
                        (activity) => activity.enabled === true
                      ) || []
                    }
                    timelineUsers={accountDetails.data?.account_users || []}
                    collapseAll={collapseAll}
                    setCollapseAll={setCollapseAll}
                    granularity={granularity}
                    loading={accountDetails?.isLoading}
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
  accountDetails: state.timelines.accountDetails
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getProfileAccountDetails,
      fetchProjectSettings,
      udpateProjectSettings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(AccountDetails);
