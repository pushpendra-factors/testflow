import React, {
    useState,
    useEffect
} from 'react';
import { Text, SVG } from 'factorsComponents';
import {
    Row,
    Col,
    Menu,
    Dropdown,
    Button,
    Table,
    notification,
    Tabs,
    Badge,
    Switch,
    Modal,
    Space,
    Input,
    Tag,
    Collapse,
    Select,
    Form
} from 'antd';

const WorkflowTrigger = ({
    onChangeSegmentType,
    segmentType,
    selectedSegment,
    onChangeSegment,
    segmentOptions,
    queryList,
    activeGrpBtn

}) => {
    return (<>
        <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}> When do you want to trigger this workflow? </Text>
        <Text type={'title'} level={8} weight={'thin'} color={'grey'} extraClass={'mt-1 mb-2'}>Account has performed event</Text>

        <Row>
            <Col span={22}>
                <Select
                    showSearch
                    style={{ minWidth: 350 }}
                    className='fa-select'
                    placeholder='Select segment type'
                    optionFilterProp='children'
                    onChange={onChangeSegmentType}
                    filterOption={(input, option) =>
                        option.props.children
                            .toLowerCase()
                            .indexOf(input.toLowerCase()) >= 0
                    }
                    value={segmentType}
                >
                    {activeGrpBtn === 'users' ? (
                        <Option value='action_event'>Performs an event</Option>
                    ) : (
                        <>
                            <Option value='action_event'>Performs an event</Option>
                            <Option value='action_segment_entry'>
                                Enter the segment
                            </Option>
                            <Option value='action_segment_exit'>
                                Exit the segment
                            </Option>
                        </>
                    )}
                </Select>
            </Col>
        </Row>


        {segmentType !== 'action_event' ? (
            <>
                <Row className='mt-4'>
                    <Col span={18}>
                        <Text type={'title'} level={7} extraClass={'m-0'}>
                            Segment name
                        </Text>
                    </Col>
                </Row>
                <Row className='mt-2'>
                    <Col span={18}>
                        <Select
                            showSearch
                            style={{ minWidth: 350 }}
                            className='fa-select'
                            placeholder='Select or search segment'
                            labelInValue
                            value={selectedSegment}
                            onChange={onChangeSegment}
                            filterOption={(input, option) => {
                                return (
                                    option?.value
                                        ? getSegmentNameFromId(option?.value).toLowerCase()
                                        : ''
                                ).includes(input.toLowerCase());
                            }}
                            options={segmentOptions}
                        ></Select>
                    </Col>
                </Row>
            </>
        ) : (
            <>
                <Row className='mt-4'>
                    <Col span={18}>
                        <Text type={'title'} level={7} extraClass={'m-0'}>
                            Event details
                        </Text>
                    </Col>
                </Row>
                <Row className={'mt-2'}>
                    <Col span={22}>
                        <div className=''>
                            <Form.Item name='event_name' className={'m-0'}>
                                {queryList()}
                            </Form.Item>
                        </div>
                    </Col>
                </Row>
            </>
        )}
    </>)
}

export default WorkflowTrigger