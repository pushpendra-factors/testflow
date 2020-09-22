import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton, Input, Select
} from 'antd';
import { Text } from 'factorsComponents';
import { UserOutlined } from '@ant-design/icons';

const { Option } = Select;

function EditBasicSettings(props) {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    setTimeout(() => {
      setDataLoading(false);
    }, 200);
  });

  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Basic Details</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button disabled={dataLoading} onClick={() => props.setEditMode(false)}>Cancel</Button>
              <Button type="primary" disabled={dataLoading} className={'ml-2'} onClick={() => props.setEditMode(false)}>Save</Button>
            </div>
          </Col>
        </Row>
        <Row className={'mt-2'}>
          <Col>
            { dataLoading ? <Skeleton.Avatar active={true} size={104} shape={'square'} />
              : <Avatar size={104} shape={'square'} icon={<UserOutlined />} />
            }
            <Text type={'paragraph'} mini extraClass={'m-0 mt-1'} color={'grey'} >A logo helps personalise your Project</Text>
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col span={24}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Project Name</Text>
            <Input className={'fa-input'} size={'large'} placeholder="Project Name" />
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col span={24}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Project URL</Text>
            <Input className={'fa-input'} size={'large'} placeholder="Project URL" />
          </Col>
        </Row>

        <Row className={'mt-6'}>
          <Col span={24}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Date Format</Text>
            <Select className={'fa-select w-full'} size={'large'} defaultValue="DD-MM-YYYY">
                <Option value="DD-MM">DD-MM-YYYY</Option>
                <Option value="MM-DD">MM-DD-YYYY</Option>
            </Select>
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col span={24}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Time Format</Text>
            <Select className={'fa-select w-full'} size={'large'} defaultValue="DD-MM-YYYY">
                <Option value="12 Hours">12 Hours</Option>
                <Option value="24 Hours">24 Hours</Option>
            </Select>
          </Col>
        </Row>

        <Row className={'mt-6'}>
          <Col span={24}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Time Zone</Text>
            <Select className={'fa-select w-full'} size={'large'} defaultValue="DD-MM-YYYY">
                <Option value="India and Sri Lanka">IST -- UTC +5:30 India and Sri Lanka</Option>
                <Option value="India">IST -- UTC +3:30 India</Option>
            </Select>
          </Col>
        </Row>
      </div>

    </>

  );
}

export default EditBasicSettings;
