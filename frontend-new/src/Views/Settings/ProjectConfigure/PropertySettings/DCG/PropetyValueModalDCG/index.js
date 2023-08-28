import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import { Text } from 'factorsComponents';
import { Modal, Form, Row, Col, Select, Input, Button, message } from 'antd';
import { getEventPropertiesV2 } from 'Reducers/coreQuery/middleware';
import { udpateProjectDetails } from 'Reducers/global';
import defaultRules from '../defaultRules';
import _ from 'lodash';
import useAutoFocus from 'hooks/useAutoFocus';
import { operatorMap } from 'Utils/operatorMapping';
import GlobalFilter from 'Components/GlobalFilter';

function PropertyValueModal({
  eventPropertiesV2,
  activeProject,
  getEventPropertiesV2,
  isModalVisible,
  setShowModalVisible,
  setShowDCGForm,
  udpateProjectDetails,
  setTabNo,
  editProperty,
  setEditProperty
}) {
  const [form] = Form.useForm();
  const [globalFilters, setGlobalFilters] = useState([]);
  const [filterProps, setFilterProperties] = useState({});
  const [loading, setLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const inputComponentRef = useAutoFocus(isModalVisible);

  useEffect(() => {
    if (!eventPropertiesV2['$session']) {
      getEventPropertiesV2(activeProject.id, '$session');
    }
    if (eventPropertiesV2) {
      const props = {};
      props['event'] = eventPropertiesV2['$session'];
      setFilterProperties(props);
    }
  }, [eventPropertiesV2]);

  // console.log('eventPropertiesV2',eventPropertiesV2);
  const onReset = () => {
    // seterrorInfo(null);
    // setVisible(false);
    // handleCancel();
    form.resetFields();
  };

  const onFinishValues = (data) => {
    seterrorInfo('');
    if (!_.isEmpty(globalFilters)) {
      setLoading(true);
      let dataSet = {
        channel: data.value,
        conditions: getGlobalFilters(globalFilters)
      };

      let ruleSet = null;
      if (activeProject?.channel_group_rules) {
        ruleSet = activeProject?.channel_group_rules;
      } else {
        ruleSet = defaultRules;
      }

      // if (_.isEmpty(activeProject?.channel_group_rules)) {
      //   ruleSet = defaultRules;
      // }

      let FinalDataSet = [];
      if (editProperty) {
        let currentArr = ruleSet;
        currentArr[editProperty?.index] = dataSet;
        FinalDataSet = [...currentArr];
      } else {
        FinalDataSet = [...ruleSet, dataSet];
      }
      udpateProjectDetails(activeProject.id, {
        channel_group_rules: FinalDataSet
      })
        .then(() => {
          message.success('Channel Group added!');
          // setVisible(false);
          onReset();
          setLoading(false);
          handleCancel();
        })
        .catch((err) => {
          console.log('err->', err);
          // seterrorInfo(err.data.error);
          setLoading(false);
        });
    } else {
      setLoading(false);
      seterrorInfo('Please add condition(s)');
    }
  };
  const onChangeValue = () => {
    seterrorInfo('');
  };

  const handleCancel = () => {
    onReset();
    setTabNo(2);
    setShowDCGForm(false);
    setShowModalVisible(false);
    setEditProperty(null);
    setLoading(false);
    setGlobalFilters([]);
    seterrorInfo(null);
  };

  const getGlobalFilters = (globalFilters = []) => {
    const filterProps = [];
    globalFilters.forEach((fil) => {
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            logical_operator: !index ? 'AND' : 'OR',
            condition: operatorMap[fil.operator],
            property: fil.props[1],
            // ty: fil.props[1],
            value: val
          });
        });
      } else {
        filterProps.push({
          logical_operator: 'AND',
          condition: operatorMap[fil.operator],
          property: fil.props[1],
          // ty: fil.props[1],
          value: fil.values
        });
      }
    });

    return filterProps;
  };

  useEffect(() => {
    if (editProperty) {
      setGlobalFilters(editProperty?.conditions);
    }
    return () => {
      onReset();
      setGlobalFilters([]);
    };
  }, [editProperty]);

  return (
    <Modal
      title='Add channel group'
      visible={isModalVisible}
      onCancel={() => handleCancel()}
      footer={null}
      width={750}
    >
      <Form
        form={form}
        onFinish={onFinishValues}
        className={'w-full'}
        onChange={onChangeValue}
        loading={false}
        initialValues={{
          ['value']: editProperty ? editProperty?.channel : ''
        }}
      >
        <Row className={'mt-8'}>
          <Col span={24}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Channel
            </Text>
            <Form.Item
              name='value'
              rules={[
                { required: true, message: 'Please enter a channel name' }
              ]}
            >
              <Input
                disabled={false}
                size='large'
                className={'fa-input w-full'}
                placeholder='Channel'
                ref={inputComponentRef}
              />
            </Form.Item>
          </Col>
        </Row>

        <Row className={'mt-8'}>
          <Col span={24}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Condition(s)
            </Text>
            <div style={{ width: '100%' }}>
              <GlobalFilter
                event={{ label: '$session' }}
                filters={globalFilters}
                setGlobalFilters={setGlobalFilters}
              />
            </div>
          </Col>
        </Row>

        {errorInfo && (
          <Row>
            <Col span={24}>
              <div className={'flex items-center mt-4'}>
                <Text type={'title'} color={'red'} size={'7'} className={'m-0'}>
                  {errorInfo}
                </Text>
              </div>
            </Col>
          </Row>
        )}

        <Row className={'mt-8'}>
          <Col span={24}>
            <div className='flex justify-end'>
              <Button size={'large'} disabled={loading} onClick={handleCancel}>
                Cancel
              </Button>
              <Button
                size={'large'}
                disabled={loading}
                loading={loading}
                className={'ml-2'}
                type={'primary'}
                htmlType='submit'
              >
                Save
              </Button>
            </div>
          </Col>
        </Row>
      </Form>
    </Modal>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2
});

export default connect(mapStateToProps, {
  getEventPropertiesV2,
  udpateProjectDetails
})(PropertyValueModal);
