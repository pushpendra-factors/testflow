import React, { useState, useEffect } from 'react';
import {
  Row,
  Col,
  Button,
  Avatar,
  Dropdown,
  Menu,
  Popover,
  Checkbox,
} from 'antd';
import { Text, SVG } from '../../factorsComponents';
import AccountTimeline from './AccountTimeline';
import { granularityOptions } from '../utils';

function AccountDetails({ onCancel, accountDetails }) {
  const [granularity, setGranularity] = useState('Daily');
  const [collapseAll, setCollapseAll] = useState(true);
  const [activities, setActivities] = useState([]);

  useEffect(() => {
    setActivities(accountDetails.data?.account_events);
  }, [accountDetails]);

  const handleChange = (option) => {
    setActivities((currActivities) => {
      const newState = currActivities.map((activity) => {
        if (activity.display_name === option.display_name) {
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

        {activities?.length ? (
          activities
            .filter(
              (value, index, self) =>
                index ===
                self.findIndex((t) => t.display_name === value.display_name)
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
                    {option.display_name}
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

  const granularityMenu = (
    <Menu>
      {granularityOptions.map((option) => {
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

  return (
    <>
      <div
        className={'fa-modal--header px-8'}
        style={{ borderBottom: '1px solid #e7e9ed' }}
      >
        <Row justify={'space-between'} className={'my-3 m-0'}>
          <Col className='flex items-center'>
            <Button
              style={{ padding: '0' }}
              type='text'
              icon={<SVG name={'brand'} size={36} />}
              size={'large'}
              onClick={() => {
                onCancel();
                setGranularity('Daily');
                setActivities([]);
                setCollapseAll(true);
              }}
            />
            <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'}>
              Account Details
            </Text>
          </Col>
          <Col>
            <Button
              size={'large'}
              type='text'
              onClick={() => {
                onCancel();
                setGranularity('Daily');
                setActivities([]);
                setCollapseAll(true);
              }}
              icon={<SVG name='times'></SVG>}
            ></Button>
          </Col>
        </Row>
      </div>

      <div className='mt-16'>
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
                    {accountDetails?.data?.name}
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
                    Industry
                  </Text>

                  <Text type={'title'} level={7} extraClass={'m-0'}>
                    {accountDetails?.data?.industry}
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
                    {accountDetails?.data?.country}
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
                    Employee Size
                  </Text>
                  <Text type={'title'} level={7} extraClass={'m-0'}>
                    {accountDetails?.data?.number_of_employees}
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
                    Number of Users
                  </Text>
                  <Text type={'title'} level={7} extraClass={'m-0'}>
                    {parseInt(accountDetails?.data?.number_of_users) > 25
                      ? '25+'
                      : accountDetails?.data?.number_of_users}
                  </Text>
                </Col>
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
                        onClick={() => setCollapseAll(false)}
                      >
                        <SVG name='line_height' size={22} />
                      </Button>
                      <Button
                        className='fa-dd--custom-btn'
                        type='text'
                        onClick={() => setCollapseAll(true)}
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
                      <Dropdown
                        overlay={granularityMenu}
                        placement='bottomRight'
                      >
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
export default AccountDetails;
