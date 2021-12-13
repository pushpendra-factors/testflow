import React, { useState, useEffect } from 'react';
import {
  Row, Col, Form, Input, Button, Modal, Radio, Space, Select
} from 'antd'; 
import { PlusOutlined } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents';
const { Option, OptGroup } = Select;


function AddEditValue ({visible, handleCancel, submitValues}) {
    const [modalForm] = Form.useForm();
    const [comboOp, setComboOper] = useState('AND');

    const onFinishValues = (values) => {
        submitValues(values);
        modalForm.resetFields();
    }

    const onChangeValue = () => {

    }

    const onSelectCombinationOperator = (val) => {
        setComboOper(val.target.value);
    }

    return (
        <>
            <Modal
            title={null}
            visible={visible}
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
                    <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Add Value</Text>
                    <Row className={'mt-4'}>
                        <Col span={24}>
                            <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Value</Text>
                            <Form.Item
                            name="content_group_value"
                            rules={[{ required: true, message: 'Please enter a value' }]}
                            >
                            <Input size="large" className={'fa-input w-full'} />
                            </Form.Item>
                        </Col>
                    </Row>
                    <Row className={'mt-4'}>
                        <Col span={12}>
                            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Rules</Text>
                        </Col>
                        <Col span={12}>
                            <div className={'flex justify-end items-baseline'}>
                            <Text type={'title'} level={7} extraClass={'mr-2'}>Operator:</Text>
                                <Radio.Group onChange={onSelectCombinationOperator} value={comboOp}>
                                    <Radio value={'AND'}>And</Radio>
                                    <Radio value={'OR'}>Or</Radio>
                                </Radio.Group>
                            </div>
                        </Col>
                    </Row>

                    <Row className={'mt-4'}>
                        <Col span={16}>
                        <Form.List name="rule">
                            {(fields, { add, remove }) => (
                            <>
                                {fields.map(({ key, name, fieldKey, ...restField }) => (
                                <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                                    <Form.Item
                                    {...restField}
                                    initialValue={fieldKey===0?'URL':comboOp}
                                    name={[name, 'lop']}
                                    fieldKey={[fieldKey, 'lop']}
                                    >
                                        <Select showArrow={false} open={false} bordered={false}>
                                            <Option value={fieldKey===0?'URL':comboOp}>{fieldKey===0?'Page':comboOp}</Option>
                                        </Select>
                                    </Form.Item>
                                    <div className={'w-24 fa-select'}>
                                        <Form.Item
                                        {...restField}
                                        initialValue={'contains'}
                                        name={[name, 'op']}
                                        fieldKey={[fieldKey, 'op']}
                                        rules={[{ required: true, message: 'Select any' }]}
                                        >
                                        
                                            <Select showArrow={false}>
                                                <Option value="contains">Contains</Option>
                                                <Option value="startsWith">Starts With</Option>
                                                <Option value="endsWith">Ends With</Option>
                                            </Select>
                                        </Form.Item>
                                    </div>
                                    <Form.Item
                                    {...restField}
                                    name={[name, 'va']}
                                    fieldKey={[fieldKey, 'va']}
                                    rules={[{ required: true, message: 'Missing value' }]}
                                    >
                                    <Input placeholder="value" className={'fa-input'}/>
                                    </Form.Item>
                                <Button type={'text'} onClick={() => remove(name)}><SVG name={'Delete'} size={18} color='gray' /></Button>
                                </Space>
                                ))}
                                <Form.Item>
                                    <div className={'w-24'}>
                                    <Button size={'middle'} onClick={() => add()} block icon={<PlusOutlined />}>
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
                            <div className="flex justify-end">
                            <Button size={'large'} onClick={handleCancel}>Cancel</Button>
                            <Button size={'large'}  className={'ml-2'} type={'primary'} htmlType="submit">Save</Button>
                            </div>
                        </Col>
                    </Row>

                </Form>
            </Modal>
        </>
    );
}

export default AddEditValue;