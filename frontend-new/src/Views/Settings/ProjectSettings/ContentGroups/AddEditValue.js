import React, { useState, useEffect } from 'react';
import {
  Row, Col, Form, Input, Button, Modal, Radio
} from 'antd'; 
import { Text, SVG } from 'factorsComponents';


function AddEditValue (props) {
    const [radioValue, setRadioValue] = useState('and');

    const onFinishValues = () => {

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
                        <Col span={14}>
                            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Rules</Text>
                        </Col>
                        <Col span={10} className={'flex justify-end'}>
                            <Text
                                type={'paragraph'}
                                mini
                                weight={'thin'}
                                color={'grey-2'}
                                extraClass={'m-0 mr-2 inline'}
                            >
                                Operator: 
                            </Text>
                            <Radio.Group onChange={onRadioChange} value={radioValue}>
                                <Radio value={'and'}>And</Radio>
                                <Radio value={'or'}>Or</Radio>
                            </Radio.Group>
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