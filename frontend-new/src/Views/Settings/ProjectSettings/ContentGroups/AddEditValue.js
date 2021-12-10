import React, { useState, useEffect } from 'react';
import {
  Row, Col, Form, Input, Button, Modal, Radio, Space, Select
} from 'antd'; 
import { PlusOutlined } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents';


function AddEditValue (props) {
    const [radioValue, setRadioValue] = useState('AND');

    const onFinishValues = (values) => {
        console.log(values);

    }

    const onChangeValue = () => {

    }

    const onRadioChange = (e) => {
        setRadioValue(e.target.value);
    }

    const handleCancel = () => {
        props.setshowAddValueModal(false);
    }

    return (
        <>
            <Modal
            title={null}
            visible={props.visible}
            closable={false}
            footer={null}
            className={'fa-modal--regular'}
            >
                <Form
                // form={form}
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
                            name="value"
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
                            <Form.Item
                            name="combOperator"
                            initialValue={'AND'}
                            rules={[{ required: true, message: 'Select one value' }]}
                            >
                                <Radio.Group onChange={onRadioChange} value={radioValue}>
                                    <Radio value={'AND'}>And</Radio>
                                    <Radio value={'OR'}>Or</Radio>
                                </Radio.Group>
                            </Form.Item>
                            </div>
                        </Col>
                    </Row>

                    <Row className={'mt-4'}>
                        <Col span={16}>
                        <Form.List name="rules">
                            {(fields, { add, remove }) => (
                            <>
                                {fields.map(({ key, name, fieldKey, ...restField }) => (
                                <Space key={key} style={{ display: 'flex', marginBottom: 8 }} align="baseline">
                                    <Form.Item
                                    {...restField}
                                    initialValue={fieldKey===0?'page':radioValue}
                                    name={[name, 'operator']}
                                    fieldKey={[fieldKey, 'operator']}
                                    >
                                        <Select showArrow={false} open={false} bordered={false}>
                                            <Option value={fieldKey===0?'page':radioValue}>{fieldKey===0?'Page':radioValue}</Option>
                                        </Select>
                                    </Form.Item>
                                    <div className={'w-24 fa-select'}>
                                        <Form.Item
                                        {...restField}
                                        initialValue={'contains'}
                                        name={[name, 'filter']}
                                        fieldKey={[fieldKey, 'filter']}
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
                                    name={[name, 'value']}
                                    fieldKey={[fieldKey, 'value']}
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