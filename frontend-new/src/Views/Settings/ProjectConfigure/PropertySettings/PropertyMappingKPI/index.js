import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import {
  Row,
  Col,
  Form,
  Button,
  Input,
  message
} from 'antd';
import { Text } from 'factorsComponents';
import QueryBlock from './QueryBlock';
import { getPropertiesDetails } from './utils';

import {
  fetchPropertyMappings,
  addPropertyMapping
} from 'Reducers/settings/middleware';



const validateRegex = /^[a-zA-Z0-9_ ]{1,}$/;


const PropertyMappingForm = ({
  KPI_config,
  setShowForm,
  setTabNo,
  activeProject,
  fetchPropertyMappings,
  addPropertyMapping
}) => {
  const [queries, setQueries] = useState([]);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const createEventMapping = (values) => {
    setLoading(true);
    let payload = {
      "display_name": values?.name,
      "properties": queries ? getPropertiesDetails(queries) : []
    }
    addPropertyMapping(activeProject?.id, payload).then(() => {
      fetchPropertyMappings(activeProject?.id);
      setLoading(false);
      setTabNo(3);
      setShowForm(false)
      message.success('Property Map added!');
    }).catch((err) => {
      message.error(err?.data?.error);
      console.log('Property Map creation failed-->', err);
      setLoading(false);
    });
  }

  const renderPropertyForm = () => {
    return (
      <>
        <Form
          form={form}
          name="login"
          onFinish={createEventMapping}

        >


          <Row>
            <Col span={12}>
              <Text
                type={'title'}
                level={3}
                weight={'bold'}
                extraClass={'m-0'}
              >
                Add new mapping
              </Text>
            </Col>
            <Col span={12}>
              <div className={'flex justify-end'}>
                <Button
                  size={'large'}
                  disabled={false}
                  onClick={() => setShowForm(false)}
                >
                  Cancel
                </Button>

                <Button
                  size={'large'}
                  disabled={false}
                  className={'ml-2'}
                  type={'primary'}
                  // onClick={() => createEventMapping()}
                  htmlType='submit'
                  loading={loading}
                >
                  Save
                </Button>
              </div>
            </Col>
          </Row>


          <Row className={'mt-8'}>
            <Col span={18}>
              <Text type={'title'} level={7} extraClass={'m-0'}>
                Display Name
              </Text>
              <Form.Item
                name='name'
                rules={[
                  { 
                    // required: true,  
                    message: 'Please input Display name (Only letters, numbers, and underscores are allowed)',
                    pattern: validateRegex
                   }
                ]}
              >
                <Input
                  // value={smartPropState.name}
                  size='large'
                  className={'fa-input w-full'}
                  placeholder='Display Name'
                />
              </Form.Item>
            </Col>
          </Row>
          {/* 
          <Row className={'mt-8'}>
            <Col span={18}>
              <Text type={'title'} level={7} extraClass={'m-0'}>
                ID Name{' '}
              </Text>
              <Form.Item
                name='description'
                rules={[{ required: true, message: 'Please enter description.' }]}
              >
                <Input
                  // value={smartPropState.description}
                  size='large'
                  className={'fa-input w-full'}
                  placeholder='ID Name'
                />
              </Form.Item>
            </Col>
          </Row> */}

        </Form>

      </>
    );
  };


  const handleEventChange = (...props) => {
    // console.log('handleEventChange', props)
    queryChange(...props)
  };

  const queryChange = (newEvent, index, changeType = 'add') => {
    const queryupdated = [...queries];
    if (queryupdated[index]) {
      if (changeType === 'add') {
        if (
          JSON.stringify(queryupdated[index]) !== JSON.stringify(newEvent)
        ) {
          // deleteGroupByForEvent(newEvent, index);
        }
        queryupdated[index] = newEvent;
      } else if (changeType === 'filters_updated') {
        // dont remove group by if filter is changed
        queryupdated[index] = newEvent;
      } else {
        // deleteGroupByForEvent(newEvent, index);
        queryupdated.splice(index, 1);
      }
    } else {
      queryupdated.push(newEvent);
    }
    setQueries(queryupdated)
    // setProfileQueries(queryupdated);
  }


  const queryList = () => {
    const blockList = [];
    let event = {};

    queries.forEach((event, index) => {
      blockList.push(
        <div key={index}
        >
          <QueryBlock
            index={index + 1}
            queryType={'KPI'}
            event={event}
            queries={queries}
            eventChange={handleEventChange}
          // setSelectedMainCategory={setSelectedMainCategory}
          // KPIConfigProps={KPIConfigProps}
          />
        </div>
      );
    });

    if (queries.length < 10) {
      blockList.push(
        <div key={'init'}
        //  className={styles.composer_body__query_block}
        >
          <QueryBlock
            queryType={'KPI'}
            index={queries.length + 1}
            queries={queries}
            eventChange={handleEventChange}
          // groupBy={queryOptions.groupBy}
          // selectedMainCategory={selectedMainCategory} 
          // KPIConfigProps={KPIConfigProps}
          />
        </div>
      );
    }

    return blockList;
  };

  return (
    <>
      {renderPropertyForm()}

      <Row>
        <Col span={24}>
          <div className='flex flex-col mt-8'>
            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Properties to Map</Text>
            <div className='flex items-center flex-wrap mr-10'>
              {queryList()}
            </div>
          </div>
        </Col>
      </Row>
    </>
  )
}



const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  KPI_config: state.kpi?.config,
});

export default connect(mapStateToProps, { fetchPropertyMappings, addPropertyMapping })(PropertyMappingForm)