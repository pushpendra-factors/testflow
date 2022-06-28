import React, { useState } from 'react';
import {
  Row,
  Col,
  Button,
  Avatar,
  Radio,
  Menu,
  Dropdown,
  Popover,
  Checkbox,
} from 'antd';
import { SVG, Text } from '../../factorsComponents';
import FaTimeline from '../../FaTimeline';

function ContactDetails({ onCancel, userDetails, timelineLoading }) {
  const allActivitiesEnabled = userDetails?.user_activities?.map((activity) => {
    return {
      ...activity,
      enabled: true,
    };
  });
  const [activities, setActivities] = useState(allActivitiesEnabled);
  const [granularity, setGranularity] = useState('Hourly');
  const [collapse, setCollapse] = useState(true);
  const options = ['Default', 'Hourly', 'Daily', 'Weekly', 'Monthly'];
  const menu = (
    <Menu>
      {options.map((option) => {
        return (
          <Menu.Item key={option} onClick={(key) => setGranularity(key.key)}>
            <div className={'flex items-center'}>
              <span className='mr-3'>{option}</span>
            </div>
          </Menu.Item>
        );
      })}
    </Menu>
  );

  const handleChange = (option) => {
    setActivities((currActivities) => {
      const newState = currActivities.map((activity) => {
        if (activity.event_name === option.event_name) {
          return {
            ...activity,
            enabled: !activity.enabled,
          };
        }
        return activity;
      });
      return newState;
    });
  };

  const controlsPopover = () => {
    return (
      <div className='fa-popupcard'>
        <div className='fa-search-bar'>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey'
            extraClass='px-2 pt-2'
          >
            Filter Activities
          </Text>
          <div className={'fa-popupcard-divider'} />
        </div>

        {activities
          ?.filter(
            (value, index, self) =>
              index === self.findIndex((t) => t.event_name === value.event_name)
          )
          .map((option) => {
            return (
              <div
                key={option.event_name}
                className='flex justify-start items-center px-4 py-2'
              >
                <div className='mr-2'>
                  <Checkbox
                    checked={option.enabled}
                    onChange={handleChange.bind(this, option)}
                  />
                </div>
                <Text mini extraClass='mb-0' type='paragraph'>
                  {option.display_name || option.event_name}
                </Text>
              </div>
            );
          }) || <div className='text-center p-2 italic'>No Activity</div>}
      </div>
    );
  };

  return (
    <>
      <div
        className={'fa-modal--header'}
        style={{ borderBottom: '1px solid #e7e9ed' }}
      >
        <div className={'fa-container'}>
          <Row justify={'space-between'} className={'my-3 m-0'}>
            <Col className='flex items-center'>
              <SVG name={'brand'} size={36} />
              <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'}>
                Contact Details
              </Text>
            </Col>
            <Col>
              <Button
                size={'large'}
                type='text'
                onClick={() => {
                  onCancel();
                }}
                icon={<SVG name='times'></SVG>}
              ></Button>
            </Col>
          </Row>
        </div>
      </div>

      <div className='my-16'>
        <Row span={24} gutter={[24, 24]}>
          <Col span={5} style={{ borderRight: '1px solid #e7e9ed' }}>
            <div className={'ml-12 my-12'}>
              <Row className={''}>
                <Col>
                  <Avatar
                    size={72}
                    style={{
                      color: '#3E516C',
                      backgroundColor: '#F1F1F1',
                      fontSize: '42px',
                      textTransform: 'uppercase',
                      fontWeight: '400',
                    }}
                  >
                    {userDetails?.name[0] || 'U'}
                  </Avatar>
                </Col>
              </Row>
              <Row className='py-2'>
                <Col>
                  <Text
                    type={'title'}
                    level={6}
                    extraClass={'m-0'}
                    weight={'bold'}
                  >
                    {!userDetails?.is_anonymous
                      ? userDetails?.name || '-'
                      : 'Unidentified User'}
                  </Text>
                  {userDetails?.role && userDetails?.company ? (
                    <Text
                      type={'title'}
                      level={7}
                      extraClass={'m-0'}
                      color={'grey'}
                    >
                      {`${userDetails?.role || '-'}, ${
                        userDetails?.company || '-'
                      }`}
                    </Text>
                  ) : (
                    <Text
                      type={'title'}
                      level={7}
                      extraClass={'m-0'}
                      color={'grey'}
                    >
                      {`${userDetails?.user_id || '-'}`}
                    </Text>
                  )}
                </Col>
              </Row>
              <Row className={'py-2'}>
                <Col>
                  <Text
                    type={'title'}
                    level={7}
                    extraClass={'m-0'}
                    color={'grey'}
                  >
                    Email
                  </Text>

                  <Text type={'title'} level={7} extraClass={'m-0'}>
                    {userDetails?.email || '-'}
                  </Text>
                </Col>
              </Row>
              <Row className={'py-2'}>
                <Col>
                  <Text
                    type={'title'}
                    level={7}
                    extraClass={'m-0'}
                    color={'grey'}
                  >
                    Country
                  </Text>
                  <Text type={'title'} level={7} extraClass={'m-0'}>
                    {userDetails?.country || '-'}
                  </Text>
                </Col>
              </Row>
              <Row className={'py-2'}>
                <Col>
                  <Text
                    type={'title'}
                    level={7}
                    extraClass={'m-0'}
                    color={'grey'}
                  >
                    Number of Web Sessions
                  </Text>
                  <Text type={'title'} level={7} extraClass={'m-0'}>
                    {userDetails?.web_sessions_count || '-'}
                  </Text>
                </Col>
              </Row>
              <Row className={'py-2'}>
                <Col>
                  <Text
                    type={'title'}
                    level={7}
                    extraClass={'m-0'}
                    color={'grey'}
                  >
                    Number of Page Views
                  </Text>
                  <Text type={'title'} level={7} extraClass={'m-0'}>
                    {userDetails?.number_of_page_views || '-'}
                  </Text>
                </Col>
              </Row>
              <Row className={'py-2'}>
                <Col>
                  <Text
                    type={'title'}
                    level={7}
                    extraClass={'m-0'}
                    color={'grey'}
                  >
                    Time Spent on Site
                  </Text>
                  <Text type={'title'} level={7} extraClass={'m-0'}>
                    {userDetails?.time_spent_on_site || '-' + ' secs'}
                  </Text>
                </Col>
              </Row>
              <Row
                className={'mt-3 pt-3'}
                style={{ borderTop: '1px dashed #e7e9ed' }}
              >
                <Col>
                  <Text
                    type={'title'}
                    level={7}
                    extraClass={'m-0 my-2'}
                    color={'grey'}
                  >
                    Associated Groups:
                  </Text>
                  {userDetails?.groups?.map((group) => {
                    return (
                      <Text type={'title'} level={7} extraClass={'m-0 mb-2'}>
                        {group.group_name}
                      </Text>
                    );
                  }) || '-'}
                </Col>
              </Row>
              <Row className={'mt-6'}>
                <Col className={'flex justify-start items-center'}></Col>
              </Row>
            </div>
          </Col>
          <Col span={18}>
            <Row gutter={[24, 24]} justify='left'>
              <Col span={24} className='mx-8 my-12'>
                <Col className='flex items-center justify-between mb-4'>
                  <div>
                    <Text type={'title'} level={3} weight={'bold'}>
                      Timeline
                    </Text>
                  </div>
                  <div className='flex justify-between'>
                    <div>
                      <Radio.Group
                        onChange={(e) => setCollapse(e.target.value)}
                        defaultValue={true}
                      >
                        <Radio.Button
                          value={false}
                          className={'fa-btn--custom'}
                        >
                          <SVG name='line_height' size={22} />
                        </Radio.Button>
                        <Radio.Button value={true} className={'fa-btn--custom'}>
                          <SVG name='grip_lines' size={22} />
                        </Radio.Button>
                      </Radio.Group>
                    </div>
                    <div>
                      <Popover
                        overlayClassName='fa-activity--filter'
                        placement='bottomLeft'
                        trigger='hover'
                        content={controlsPopover}
                      >
                        <Button
                          size='large'
                          className='fa-btn--custom mx-2 relative'
                          type='text'
                        >
                          <SVG name={'controls'} />
                        </Button>
                      </Popover>
                    </div>
                    <div>
                      <Dropdown overlay={menu} placement='bottomRight'>
                        <Button
                          className={`ant-dropdown-link flex items-center`}
                        >
                          {granularity}
                          {<SVG name='caretDown' size={16} extraClass='ml-1' />}
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
                    granularity={granularity}
                    collapse={collapse}
                    loading={timelineLoading}
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

export default ContactDetails;
