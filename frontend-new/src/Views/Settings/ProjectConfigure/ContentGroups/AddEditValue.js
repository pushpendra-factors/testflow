import React, { useState } from 'react';
import {
  Row,
  Col,
  Form,
  Input,
  Button,
  Modal,
  Radio,
  Space,
  Select,
  notification
} from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents';
import useAutoFocus from 'hooks/useAutoFocus';
const { Option } = Select;

function AddEditValue({ selectedRule, handleCancel, submitValues }) {
  const [modalForm] = Form.useForm();
  const [comboOp, setComboOper] = useState('AND');
  const inputComponentRef = useAutoFocus();

  const onFinishValues = (values) => {
    if (values.content_group_value && values.rule) {
      values.rule.forEach((val) => {
        val.lop = comboOp;
      });
      submitValues(values, selectedRule);
    } else {
      notification.error({
        message: 'Error',
        description: 'Please add atleast one rule',
        duration: 5
      });
    }
  };

  const onChangeValue = () => {};

  const onSelectCombinationOperator = (val) => {
    setComboOper(val.target.value);
  };

  return (
    <>
      <Modal
        title={null}
        visible={true}
        closable={false}
        footer={null}
        className={'fa-modal--regular'}
      >
        <Form
          form={modalForm}
          onFinish={onFinishValues}
          className={'w-full'}
          onChange={onChangeValue}
          loading={false}
        >
          <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>
            Add Value
          </Text>
          <Row className={'mt-4'}>
            <Col span={24}>
              <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>
                Value
              </Text>
              <Form.Item
                initialValue={
                  selectedRule?.content_group_value
                    ? selectedRule.content_group_value
                    : ''
                }
                name='content_group_value'
                rules={[{ required: true, message: 'Please enter a value' }]}
              >
                <Input
                  size='large'
                  className={'fa-input w-full'}
                  ref={inputComponentRef}
                />
              </Form.Item>
            </Col>
          </Row>
          <Row className={'mt-4'}>
            <Col span={12}>
              <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>
                Rules
              </Text>
            </Col>
            <Col span={12}>
              <div className={'flex justify-end items-baseline'}>
                <Text type={'title'} level={7} extraClass={'mr-2'}>
                  Operator:
                </Text>
                <Radio.Group
                  onChange={onSelectCombinationOperator}
                  value={comboOp}
                >
                  <Radio value={'AND'}>And</Radio>
                  <Radio value={'OR'}>Or</Radio>
                </Radio.Group>
              </div>
            </Col>
          </Row>

          <Row className={'mt-4'}>
            <Col span={16}>
              <Form.List name='rule'>
                {(fields, { add, remove }) => (
                  <>
                    {fields.map(({ key, name, fieldKey, ...restField }) => (
                      <Space
                        key={key}
                        style={{ display: 'flex', marginBottom: 8 }}
                        align='baseline'
                      >
                        <div className={'w-16'}>
                          <Form.Item
                            {...restField}
                            initialValue={
                              selectedRule?.rule &&
                              selectedRule.rule[fieldKey].lop
                                ? selectedRule.rule[fieldKey].lop
                                : comboOp
                            }
                            name={[name, 'lop']}
                            fieldKey={[fieldKey, 'lop']}
                          >
                            {/* <Select showArrow={false} open={false} bordered={false}>
                                                <Option value={comboOp}>{fieldKey===0?'Page':comboOp}</Option>
                                            </Select> */}
                            <Text
                              type={'title'}
                              level={7}
                              extraClass={'m-0 ml-2'}
                            >
                              {fieldKey === 0 ? 'Page' : comboOp}
                            </Text>
                          </Form.Item>
                        </div>
                        <div className={'fa-select'}>
                          <Form.Item
                            {...restField}
                            initialValue={
                              selectedRule?.rule &&
                              selectedRule.rule[fieldKey].op
                                ? selectedRule.rule[fieldKey].op
                                : 'contains'
                            }
                            name={[name, 'op']}
                            fieldKey={[fieldKey, 'op']}
                            rules={[{ required: true, message: 'Select any' }]}
                          >
                            <Select
                              showArrow={false}
                              style={{ width: '110px' }}
                            >
                              <Option value='equals'>Equals</Option>
                              <Option value='notEqual'>Not Equals</Option>
                              <Option value='contains'>Contains</Option>
                              <Option value='notContains'>Not Contains</Option>
                              <Option value='startsWith'>Starts With</Option>
                              <Option value='endsWith'>Ends With</Option>
                            </Select>
                          </Form.Item>
                        </div>
                        <div className={'w-24'}>
                          <Form.Item
                            {...restField}
                            initialValue={
                              selectedRule?.rule &&
                              selectedRule.rule[fieldKey].va
                                ? selectedRule.rule[fieldKey].va
                                : ''
                            }
                            name={[name, 'va']}
                            fieldKey={[fieldKey, 'va']}
                            rules={[
                              { required: true, message: 'Missing value' }
                            ]}
                          >
                            <Input placeholder='value' className={'fa-input'} />
                          </Form.Item>
                        </div>
                        <Button type={'text'} onClick={() => remove(name)}>
                          <SVG name={'Delete'} size={18} color='gray' />
                        </Button>
                      </Space>
                    ))}
                    <Form.Item>
                      <div className={'w-24 mt-2'}>
                        <Button
                          size={'middle'}
                          onClick={() => add()}
                          block
                          icon={<PlusOutlined />}
                        >
                          Add rule
                        </Button>
                      </div>
                    </Form.Item>
                  </>
                )}
              </Form.List>
            </Col>
          </Row>

          <Row className={'mt-8'}>
            <Col span={24}>
              <div className='flex justify-end'>
                <Button size={'large'} onClick={handleCancel}>
                  Cancel
                </Button>
                <Button
                  size={'large'}
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
    </>
  );
}

export default AddEditValue;
