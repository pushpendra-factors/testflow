import React, { useMemo, useState } from 'react';
import { Row, Col, Button, Avatar, Table, Radio, Menu, Dropdown } from 'antd';
import { SVG, Text } from '../../factorsComponents';
import FaTimeline from '../../FaTimeline';

function ContactDetails({ onCancel, userDetails }) {
  const userDetail = useMemo(() => {
    return userDetails;
  }, [userDetails]);

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
                      color: '#f56a00',
                      backgroundColor: '#fde3cf',
                      fontSize: '42px',
                      textTransform: 'uppercase',
                      fontWeight: '400',
                    }}
                  >
                    U
                  </Avatar>
                </Col>
              </Row>
              <Row className='py-2'>
                <Col>
                  {userDetail.name ? (
                    <Text
                      type={'title'}
                      level={6}
                      extraClass={'m-0'}
                      weight={'bold'}
                    >
                      {userDetail.name}
                    </Text>
                  ) : (
                    'Unidentified User'
                  )}
                  {userDetail.role && userDetail.company ? (
                    <Text
                      type={'title'}
                      level={7}
                      extraClass={'m-0'}
                      color={'grey'}
                    >
                      {`${userDetail?.role}, ${userDetail?.company}`}
                    </Text>
                  ) : (
                    <Text
                      type={'title'}
                      level={7}
                      extraClass={'m-0'}
                      color={'grey'}
                    >
                      {`${userDetail?.user_id}`}
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
                    {userDetail?.email || '-'}
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
                    {userDetail?.country || '-'}
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
                    {userDetail?.web_sessions_count || '-'}
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
                    {userDetail?.number_of_page_views || '-'}
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
                    {userDetail?.time_spent_on_site + ' secs' || '-'}
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
                  {userDetail?.groups?.map((group) => {
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
                    <div className='mr-2'>
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
                    activities={userDetail?.user_activities}
                    granularity={granularity}
                    collapse={collapse}
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
