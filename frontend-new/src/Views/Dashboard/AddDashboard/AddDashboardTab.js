import React from 'react';
import { Text, SVG } from '../../../components/factorsComponents';
import { Row, Col, Input, Button } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
// const { Option } = Select;

function AddDashboardTab({
  title,
  setTitle,
  description,
  setDescription,
  dashboardType,
  setDashboardType,
  editDashboard,
  showDeleteModal,
  inputComponentRef
}) {
  return (
    <>
      <Row className={'pt-4'} gutter={[24, 24]}>
        <Col span={12}>
          <Text type={'title'} level={7} extraClass={'m-0'}>
            Title
          </Text>
          <Input
            onChange={(e) => setTitle(e.target.value)}
            value={title}
            className={'fa-input'}
            size={'large'}
            placeholder='Dashboard Title'
            ref={inputComponentRef}
          />
        </Col>
        <Col span={12}>
          <Text type={'title'} level={7} extraClass={'m-0'}>
            Description (Optional)
          </Text>
          <Input
            onChange={(e) => setDescription(e.target.value)}
            value={description}
            className={'fa-input'}
            size={'large'}
            placeholder='Description (Optional)'
          />
        </Col>
      </Row>
      <Row className={'pt-2'} gutter={[24, 4]}>
        <Col span={24}>
          <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>
            Who can access this dashboard ?
          </Text>
        </Col>
        <Col span={24}>
          <Row gutter={[24, 4]}>
            <Col span={12}>
              <div
                onClick={() => setDashboardType.bind(this, 'pr')}
                className={`${
                  dashboardType === 'pr'
                    ? 'fa-dasboard-privacy--card selected'
                    : 'fa-dasboard-privacy--card'
                } border-radius--medium p-4`}
              >
                <div className={'flex justify-between items-start'}>
                  <div>
                    <SVG
                      name={'lock'}
                      color={'grey'}
                      extraClass={'mr-2 mt-1'}
                    />
                  </div>
                  <div>
                    <Text
                      type={'title'}
                      level={5}
                      weight={'bold'}
                      extraClass={'m-0'}
                    >
                      Private
                    </Text>
                    <Text
                      type={'title'}
                      level={7}
                      color={'grey'}
                      extraClass={'m-0'}
                    >
                      Only you have access to the contents of Private
                      Dashboards.
                    </Text>
                  </div>
                </div>
              </div>
            </Col>
            <Col span={12}>
              <div
                onClick={() => setDashboardType.bind(this, 'pv')}
                className={`${
                  dashboardType === 'pv'
                    ? 'fa-dasboard-privacy--card selected'
                    : 'fa-dasboard-privacy--card'
                } border-radius--medium p-4`}
              >
                <div className={'flex justify-between items-start'}>
                  <div>
                    <SVG
                      name={'globe'}
                      color={'grey'}
                      extraClass={'mr-2 mt-1'}
                    />
                  </div>
                  <div>
                    <Text
                      type={'title'}
                      level={5}
                      weight={'bold'}
                      extraClass={'m-0'}
                    >
                      Public
                    </Text>
                    <Text
                      type={'title'}
                      level={7}
                      color={'grey'}
                      extraClass={'m-0'}
                    >
                      Everyone in your organization has access to this
                      dashboard.
                    </Text>
                  </div>
                </div>
              </div>
            </Col>
          </Row>
        </Col>
      </Row>
      {editDashboard ? (
        <div className='pt-2'>
          <Button
            onClick={showDeleteModal.bind(this, true)}
            style={{ display: 'flex', alignItems: 'center', padding: 0 }}
            type='text'
            icon={<DeleteOutlined />}
          >
            Delete Dashboard?
          </Button>
        </div>
      ) : null}
      {/* <Row className={'pt-4'} gutter={[24, 12]}>
            <Col span={24}>
                    <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Data display</Text>
                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>By default, render widgets with data from this period of time.</Text>
            </Col>
            <Col span={12}>
                <Select className={'fa-select w-full'} size={'large'} defaultValue="Date Range">
                    <Option value="jack">1 Month</Option>
                    <Option value="lucy2">2 Months</Option>
                    <Option value="lucy3">6 Months</Option>
                    <Option value="lucy4">1 Year</Option>
                    <Option value="lucy5">1+ Year</Option>
                </Select>
            </Col>
        </Row> */}
    </>
  );
}

export default AddDashboardTab;
