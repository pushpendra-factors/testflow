import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';
import {
  Tooltip,
  ConfigProvider,
  Empty,
  Row,
  Col,
  Form,
  Button,
  Input,
  Select,
  Table,
  notification,
  Dropdown,
  Menu
} from 'antd';
import { MoreOutlined } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents';
import { SmartPropertyClass, PropertyRule, FilterClass } from './utils';

import {
  fetchSmartPropertiesConfig,
  addSmartProperty,
  updateSmartProperty
} from 'Reducers/settings/middleware';
import PropetyValueModal from '../PropetyValueModal';

import { reverseOperatorMap } from '../utils';
import ConfirmationModal from '../../../../../components/ConfirmationModal';
import useAutoFocus from 'hooks/useAutoFocus';
const { Option } = Select;

function SmartProperties({
  activeProject,
  setShowSmartProperty,
  smartProperty,
  fetchSmartPropertiesConfig,
  config,
  addSmartProperty,
  updateSmartProperty
}) {
  const [form] = Form.useForm();

  const [formState, setFormState] = useState('add');

  const [isModalVisible, setShowModalVisible] = useState(false);

  const [selectedRule, setSelectedRule] = useState(null);

  const [valueSources, setValueSources] = useState([]);

  const [smartPropState, setSmartPropState] = useState({});

  const [rulesState, setRulesState] = useState([]);

  const [rulesData, setRulesData] = useState([]);

  const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
  const [deleteApiCalled, setDeleteApiCalled] = useState(false);
  const inputComponentRef = useAutoFocus(formState !== 'view');

  const renderFilterViewButtons = (filters) => {
    return filters.map((obj, i) => {
      return (
        <div className={`flex justify-start ${i > 0 && 'mt-4'}`}>
          <Button
            icon={
              obj && obj.name ? (
                <SVG name={obj.name} size={16} color={'grey'} />
              ) : null
            }
            className={`fa-button--truncate`}
            style={{ pointerEvents: 'none' }}
          >
            {' '}
            {obj?.property}
          </Button>

          <Button
            className={`fa-button--truncate ml-4`}
            style={{ pointerEvents: 'none' }}
          >
            {' '}
            {reverseOperatorMap[obj?.condition]}
          </Button>

          <Button
            className={`fa-button--truncate ml-4`}
            style={{ pointerEvents: 'none' }}
          >
            {' '}
            {obj?.value}
          </Button>

          {i + 1 === filters.length ? null : (
            <Text type={'title'} level={7} extraClass={'ml-2'}>
              {obj.logical_operator}
            </Text>
          )}
        </div>
      );
    });
  };

  const columns = [
    {
      title: 'Value',
      dataIndex: 'value',
      key: 'value',
      render: (text) => <span className={'capitalize'}>{text}</span>
    },
    {
      title: 'Source',
      dataIndex: 'source',
      key: 'source',
      render: (text) => <span className={'capitalize'}>{text}</span>
    },
    {
      title: 'Rule',
      dataIndex: 'rule',
      key: 'rule',
      render: (filters) => renderFilterViewButtons(filters)
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      render: (obj) => (
        <div className={`flex justify-end`}>
          <Dropdown overlay={() => menu(obj)} trigger={['click']}>
            <Button size={'large'} type='text' icon={<MoreOutlined />} />
          </Dropdown>
        </div>
      )
    }
  ];

  const menu = (obj) => {
    return (
      <Menu>
        <Menu.Item
          disabled={rulesData.length == 1 ? true : false}
          key='0'
          onClick={() => showDeleteWidgetModal(obj)}
        >
          <Tooltip
            placement='top'
            title="Can't remove because the property should have atleast one value."
            trigger={rulesData.length == 1 ? ['hover'] : []}
          >
            <a>Remove</a>
          </Tooltip>
        </Menu.Item>
        <Menu.Item key='0' onClick={() => editProp(obj)}>
          <a>Edit</a>
        </Menu.Item>
      </Menu>
    );
  };

  const editProp = (obj) => {
    setSelectedRule(obj);
    setShowModalVisible(true);
  };

  const confirmRemove = (obj) => {
    const rulesToUpdate = [
      ...smartPropState.rules.filter(
        (rule) => JSON.stringify(rule) !== JSON.stringify(obj)
      )
    ];

    if (formState !== 'add') {
      const smrtProp = Object.assign({}, smartPropState);
      smrtProp.rules = rulesToUpdate;
      updateForm(smrtProp);
    }
  };

  const confirmDelete = useCallback(async () => {
    try {
      setDeleteApiCalled(true);
      confirmRemove(deleteWidgetModal);
      setDeleteApiCalled(false);
      showDeleteWidgetModal(false);
    } catch (err) {
      console.log(err);
      console.log(err.response);
    }
  }, [deleteWidgetModal]);

  useEffect(() => {
    if (
      activeProject?.id &&
      (!config?.name || config?.name !== smartPropState.type_alias)
    ) {
      fetchSmartPropertiesConfig(
        activeProject.id,
        smartPropState?.type_alias ? smartPropState.type_alias : 'campaign'
      );
    }
  }, [activeProject, smartPropState]);

  useEffect(() => {
    const columData = [];
    rulesState.forEach((rl) => {
      columData.push({
        value: rl.value,
        source: rl.source,
        rule: rl.filters,
        actions: rl
      });
    });
    setRulesData(columData);
  }, [rulesState]);

  useEffect(() => {
    if (smartProperty) {
      setSmartPropState(smartProperty);
      setFormState('view');
      setRulesState(smartProperty.rules);
    }
  }, [smartProperty]);

  useEffect(() => {
    if (config?.sources) {
      setValueSources(config.sources.map((sr) => sr.name));
    }
  }, [config]);

  const createForm = (smrtProp) => {
    addSmartProperty(activeProject.id, smrtProp).then(
      (res) => {
        smrtProp.id = res.data.id;
        setSmartPropState({ ...smrtProp });
        setFormState('view');
        setShowModalVisible(false);
        notification.success({
          message: 'Success',
          description: 'Custom Dimension rules created successfully ',
          duration: 5
        });
      },
      (err) => {
        notification.error({
          message: 'Error',
          description: err.data.error,
          duration: 5
        });
      }
    );
  };

  const updateForm = (smrtProp) => {
    updateSmartProperty(activeProject.id, smrtProp).then(
      (res) => {
        smrtProp.id = res.data.id;
        setSmartPropState({ ...smrtProp });
        setRulesState(smrtProp.rules);
        setFormState('view');
        setShowModalVisible(false);
        notification.success({
          message: 'Success',
          description: 'Custom Dimension rules updated successfully ',
          duration: 5
        });
      },
      (err) => {
        notification.error({
          message: 'Error',
          description: err.data.error,
          duration: 5
        });
      }
    );
  };

  const onFinish = (data) => {
    if (data) {
      // Save with data
      // Close modal
      const smrtProp = new SmartPropertyClass(
        smartPropState.id ? smartPropState.id : '',
        data.name,
        data.description,
        data.type,
        rulesState
      );
      if (formState !== 'add') {
        updateForm(smrtProp);
      } else {
        delete smrtProp.id;
        createForm(smrtProp);
      }
    }
  };

  const handleCancel = () => {
    setShowModalVisible(false);
    setSelectedRule(null);
  };

  const handleValuesSubmit = (data, oldRule) => {
    if (data) {
      const valueFilters = [
        ...data.filters.map((fl) => {
          const nF = new FilterClass();
          nF.setFilter(fl, data.combOperator);
          return nF.getFilter();
        })
      ];
      const rule = new PropertyRule(data.value, data.source, valueFilters);
      const rulesToUpdate = [
        ...rulesState.filter(
          (rl) => JSON.stringify(rl) !== JSON.stringify(oldRule)
        )
      ];
      rulesToUpdate.push(rule);
      setRulesState(rulesToUpdate);
      setShowModalVisible(false);
      setSelectedRule(null);
      if (formState === 'view') {
        const smrtProp = new SmartPropertyClass(
          smartPropState.id,
          smartPropState.name,
          smartPropState.description,
          smartPropState.type_alias,
          rulesToUpdate
        );
        updateForm(smrtProp);
      }
    }
  };

  const renderPropertyForm = () => {
    return (
      <>
        <Row className={'mt-8'}>
          <Col span={18}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Type
            </Text>
            <Form.Item name='type'>
              <Select
                className={'fa-select w-full'}
                size={'large'}
                onChange={(val) => {
                  if (val !== config.name) {
                    fetchSmartPropertiesConfig(activeProject.id, val);
                  }
                }}
              >
                <Option value='campaign'>Campaign</Option>
                <Option value='ad_group'>Ad Group</Option>
              </Select>
            </Form.Item>
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
                { required: true, message: 'Please input display name.' }
              ]}
            >
              <Input
                value={smartPropState.name}
                size='large'
                className={'fa-input w-full'}
                placeholder='Display Name'
                ref={inputComponentRef}
              />
            </Form.Item>
          </Col>
        </Row>

        <Row className={'mt-8'}>
          <Col span={18}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Description{' '}
            </Text>
            <Form.Item
              name='description'
              rules={[{ required: true, message: 'Please enter description.' }]}
            >
              <Input
                value={smartPropState.description}
                size='large'
                className={'fa-input w-full'}
                placeholder='Description'
              />
            </Form.Item>
          </Col>
        </Row>
      </>
    );
  };

  const renderPropertyDetail = () => {
    return (
      <>
        <Row className={'mt-8'}>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Type
            </Text>
            <Text
              type={'title'}
              level={6}
              extraClass={'m-0 capitalize'}
              weight={'bold'}
            >
              {smartPropState.type_alias.replace('_', ' ')}
            </Text>
          </Col>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Display Name
            </Text>
            <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>
              {smartPropState.name}
            </Text>
          </Col>
        </Row>

        <Row className={'mt-8'}>
          <Col span={18}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Description{' '}
            </Text>
            <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>
              {smartPropState.description}
            </Text>
          </Col>
        </Row>
      </>
    );
  };

  const renderSmartPropForm = () => {
    return (
      <Row>
        <Col span={24}>
          <div>
            <Form
              form={form}
              onFinish={onFinish}
              className={'w-full'}
              loading={true}
              initialValues={{
                type: smartProperty?.type_alias
                  ? smartProperty.type_alias
                  : 'campaign',
                description: smartProperty?.description
                  ? smartProperty.description
                  : '',
                name: smartProperty?.name ? smartProperty.name : ''
              }}
            >
              <Row>
                <Col span={12}>
                  <Text
                    type={'title'}
                    level={3}
                    weight={'bold'}
                    extraClass={'m-0'}
                  >
                    {formState === 'add'
                      ? 'New Custom Dimension'
                      : 'Custom Dimension Details'}
                  </Text>
                </Col>
                <Col span={12}>
                  <div className={'flex justify-end'}>
                    <Button
                      size={'large'}
                      disabled={false}
                      onClick={() => setShowSmartProperty(false)}
                    >
                      Cancel
                    </Button>
                    {formState === 'view' ? (
                      <Button
                        size={'large'}
                        disabled={false}
                        className={'ml-2'}
                        onClick={() => setFormState('edit')}
                      >
                        Edit
                      </Button>
                    ) : null}
                    {formState !== 'view' ? (
                      <Tooltip
                        placement='top'
                        title='Please add at least one value for this property'
                        trigger={rulesData.length < 1 ? ['hover'] : []}
                      >
                        <Button
                          size={'large'}
                          disabled={rulesData.length < 1 ? true : false}
                          className={'ml-2'}
                          type={'primary'}
                          htmlType='submit'
                        >
                          Save
                        </Button>
                      </Tooltip>
                    ) : null}
                  </div>
                </Col>
              </Row>

              {formState !== 'view'
                ? renderPropertyForm()
                : renderPropertyDetail()}
            </Form>
          </div>
        </Col>
      </Row>
    );
  };

  const renderValuesTable = () => {
    return (
      <>
        <Row className={`mt-8`}>
          <Col span={12}>
            <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>
              Values
            </Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button size={'medium'} onClick={() => setShowModalVisible(true)}>
                <SVG name={'plus'} extraClass={'mr-2'} size={16} />
                New Value
              </Button>
            </div>
          </Col>
        </Row>
        <Row>
          <Col span={24}>
            <ConfigProvider
              renderEmpty={() => (
                <Empty
                  image={
                    'https://s3.amazonaws.com/www.factors.ai/assets/img/product/NoData.png'
                  }
                  imageStyle={{
                    display: 'flex',
                    flexDirection: 'vertical',
                    justifyContent: 'center'
                  }}
                  description={
                    <span>
                      No values added yet. Please add few values before saving
                      the property.
                    </span>
                  }
                >
                  <Button
                    size={'medium'}
                    onClick={() => setShowModalVisible(true)}
                  >
                    <SVG name={'plus'} extraClass={'mr-2'} size={16} />
                    Add New Value
                  </Button>
                </Empty>
              )}
            >
              <Table
                className='fa-table--basic mt-4'
                columns={columns}
                dataSource={rulesData}
                pagination={false}
              />
            </ConfigProvider>
          </Col>
        </Row>
        {renderValuesModal()}
      </>
    );
  };

  const renderValuesModal = () => {
    if (!isModalVisible) return null;
    return (
      <PropetyValueModal
        config={config}
        rule={selectedRule}
        sources={valueSources}
        handleCancel={handleCancel}
        submitValues={handleValuesSubmit}
      ></PropetyValueModal>
    );
  };

  return (
    <>
      {renderSmartPropForm()}
      {renderValuesTable()}
      <ConfirmationModal
        visible={deleteWidgetModal ? true : false}
        confirmationText='Do you really want to remove this value?'
        onOk={confirmDelete}
        onCancel={showDeleteWidgetModal.bind(this, false)}
        title='Remove Value'
        okText='Confirm'
        cancelText='Cancel'
        confirmLoading={deleteApiCalled}
      />
    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  config: state.settings.propertyConfig
});

export default connect(mapStateToProps, {
  fetchSmartPropertiesConfig,
  addSmartProperty,
  updateSmartProperty
})(SmartProperties);
