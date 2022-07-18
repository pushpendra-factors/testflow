import React, { useEffect, useState } from 'react';
import {
  Row,
  Col,
  Button,
  Avatar,
  Menu,
  Dropdown,
  Popover,
  Checkbox,
} from 'antd';
import { SVG, Text } from '../../factorsComponents';
import FaTimeline from '../../FaTimeline';
import { formatDurationIntoString } from '../../../utils/dataFormatter';

function ContactDetails({ onCancel, userDetails }) {
  const [activities, setActivities] = useState([]);

  useEffect(() => {
    let allActivitiesEnabled = [];
    if (userDetails.data.user_activities) {
      allActivitiesEnabled = userDetails.data.user_activities.map(
        (activity) => {
          let isEnabled = true;
          if (
            activity.display_name.includes('Contact Updated') ||
            activity.display_name.includes('Campaign Member Updated')
          )
            isEnabled = false;
          return {
            ...activity,
            enabled: isEnabled,
          };
        }
      );
    }
    setActivities(allActivitiesEnabled);
  }, [userDetails]);

  const [granularity, setGranularity] = useState('Hourly');
  const [collapse, setCollapse] = useState(true);
  const options = ['Timestamp', 'Hourly', 'Daily', 'Weekly', 'Monthly'];
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
      <div className='fa-filter-popupcard'>
        <div className='fa-header'>
          <Text
            type='title'
            level={7}
            weight='bold'
            color='grey'
            extraClass='px-2 pt-2'
          >
            Filter Activities
          </Text>
          <div className={'fa-divider'} />
        </div>

        {activities.length ? (
          activities
            .filter(
              (value, index, self) =>
                index ===
                self.findIndex((t) => t.event_name === value.event_name)
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
                  <Text mini extraClass='mb-0 truncate' type='paragraph'>
                    {option.display_name || option.event_name}
                  </Text>
                </div>
              );
            })
        ) : (
          <div className='text-center p-2 italic'>No Activity</div>
        )}
      </div>
    );
  };

  return (
    <>
      <div
        className={'fa-modal--header px-8'}
        style={{ borderBottom: '1px solid #e7e9ed' }}
      >
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
                setActivities([]);
                setCollapse(true);
                setGranularity('Hourly');
                onCancel();
              }}
              icon={<SVG name='times'></SVG>}
            ></Button>
          </Col>
        </Row>
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
                      display: 'flex',
                      alignItems: 'center',
                    }}
                  >
                    <SVG name='user' size={40} />
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
                    {!userDetails?.data?.is_anonymous
                      ? userDetails?.data?.name || '-'
                      : 'Unidentified User'}
                  </Text>
                  {
                    <Text
                      type={'title'}
                      level={7}
                      extraClass={'m-0'}
                      color={'grey'}
                    >
                      {userDetails?.data?.company || userDetails?.data?.user_id}
                    </Text>
                  }
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
                    {userDetails?.data?.email || '-'}
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
                    {userDetails?.data?.country || '-'}
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
                    {parseInt(userDetails?.data?.web_sessions_count) || '-'}
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
                    {parseInt(userDetails?.data?.number_of_page_views) || '-'}
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
                    {formatDurationIntoString(
                      userDetails?.data?.time_spent_on_site
                    )}
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
                  {userDetails?.data?.group_infos?.map((group) => {
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
                    <div className='flex justify-between'>
                      <Button
                        className='fa-dd--custom-btn'
                        type='text'
                        onClick={() => setCollapse(false)}
                      >
                        <SVG name='line_height' size={22} />
                      </Button>
                      <Button
                        className='fa-dd--custom-btn'
                        type='text'
                        onClick={() => setCollapse(true)}
                      >
                        <SVG name='grip_lines' size={22} />
                      </Button>
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
                          <SVG name={'activity_filter'} />
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

export default ContactDetails;
