import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import { Text, SVG } from 'factorsComponents';

import FAFilterSelect from '../../../../../components/FaFilterSelect';
import PropertyFilter from '../PropertiesFilter';

import {Modal, Form, Row, Col, Select, Input, Button, Radio} from 'antd';

const { Option, OptGroup } = Select;

function PropertyValueModal ({config, type = 'campaign', sources, rule, handleCancel, submitValues}) {

    const [showNewFilter, setShowNewFilter] = useState(false);
    const [modalForm] = Form.useForm();
    const [filters, setFilters] = useState([]);
    const [propertyOpts, setPropertyOpts] = useState([]);
    const [selectedSource, setSelectedSource] = useState('all');
    const [comboOp, setComboOper] = useState();

    useEffect(() => {
        if(config) {
            const src = config.sources? config.sources.filter((sr) => sr.name === selectedSource) : [];
            const propertyOptions = [];
            if(src.length && src[0].objects_and_properties?.length) {
                src[0].objects_and_properties.forEach((grp) => {
                    propertyOptions.push({
                        label: grp.name,
                        icon: grp.name,
                        values: grp.properties.map((pr) => [pr.name, pr.type])
                    })
                })
            }
            setPropertyOpts(propertyOptions);
        }
    }, [config])

    useEffect(() => {
        if(rule && rule.filters) {
            const filterConverted = [...rule.filters.map((fil) => {return {
                prop: {type: fil.name, name: fil.property, category: "categorical"},
                operator: fil.condition,
                values: fil.value
            }})];
            setFilters(filterConverted)
        }
    }, [rule])

    const onFinishValues = (data) => {
        const modalResult = {...data, filters: [...filters]}
        submitValues(modalResult, rule);
    }

    const onSelectCombinationOperator = (val) => {
        setComboOper(val.target.value);
    }

    const insertFilter = (index, filter) => {
        if(index < 0) {
            const filtersToUpdate = [...filters];
            filtersToUpdate.push(filter);
            setFilters(filtersToUpdate);
        } else {
            const filtersToUpdate = [...filters];
            filtersToUpdate[index] = filter;
            setFilters(filtersToUpdate);
        }

        setShowNewFilter(false);
    }

    const onChangeValue = () => {}

    const renderFilters = () => {
        const filterElements = [];
        filters.forEach((fil, i) => {
            filterElements.push(
                <div>
                    <PropertyFilter filter={fil} propOpts={propertyOpts} insertFilter={(filt) => insertFilter(i, filt)}></PropertyFilter>

                    {i>0? <span className={`ml-2`}>{comboOp}</span> : null}
                </div>
            )
        })
        return filterElements;
    }

    const renderSelectFilter = () => {
        const filterSelect = !showNewFilter ? 
            (<Button className={`mt-4`} size={'medium'} onClick={() => setShowNewFilter(true)}><SVG name={'plus'} extraClass={'mr-2'} size={16} />New Filter</Button>)
            :
            (<PropertyFilter propOpts={propertyOpts} insertFilter={(filt) => insertFilter(-1, filt)}></PropertyFilter>)
        return filterSelect;
    }

    const renderCombinationOperator = () => {
        return (
            <Row className={'mt-8'}>
                <Col span={24} >
                    <div className={`flex justify-end items-baseline`}>
                        <Text type={'title'} level={7} extraClass={'mr-2'}>Combination Operator</Text>
                        <Form.Item
                            name="combOperator"
                                        rules={[{ required: true, message: 'Select one value' }]}
                                >
                                <Radio.Group onChange={e=>onSelectCombinationOperator(e)}>
                                                        <Radio value={'AND'}>And</Radio>
                                                        <Radio value={'OR'}>Or</Radio> 
                                                    </Radio.Group> 
                        </Form.Item>
                    </div>
                    
                    
                </Col>
            </Row>
        )
    }

    const renderSourceOptions = () => {
        if(!sources?.length) return null;
        return sources.map((source) => {
            return <Option value={source} className={`capitalize`}> {source} </Option>
        })
    }

    return (
        <Modal title="Add/Edit new value" visible={true} onCancel={handleCancel}
            footer={null}
        >
            <Form
                    form={modalForm}
                    onFinish={onFinishValues}
                    className={'w-full'}
                    onChange={onChangeValue}
                    loading={false}
                    initialValues={{
                            value: rule?.value? rule.value : "",
                            source: rule?.source? rule.source : 'all' ,
                            combOperator: rule?.filters && rule.filters[0].logical_operator? rule.filters[0].logical_operator : 'AND'
                        }
                    }
                    >

                    <Row className={'mt-8'}>
                            <Col span={24}>
                            <Text type={'title'} level={7} extraClass={'m-0'}>Value</Text>
                            <Form.Item
                                    name="value"
                                    rules={[{ required: true, message: 'Please input value.' }]}
                            >
                            <Input disabled={false} size="large" className={'fa-input w-full'} placeholder="Value" />
                                    </Form.Item>
                            </Col> 
                    </Row>

                    <Row className={'mt-8'}>
                            <Col span={24}>
                            <Text type={'title'} level={7} extraClass={'m-0'}>Source</Text>
                            <Form.Item
                                    name="source"
                                    rules={[{ required: true, message: 'Please input value.' }]}
                            >
                                <Select className={'fa-select w-full'} defaultValue="all" size={'large'}>
                                    {renderSourceOptions()}
                                </Select>
                            </Form.Item>
                            </Col> 
                    </Row>


                    <Row className={'mt-8'}>
                            <Col span={24}>
                                <Text type={'title'} level={7} extraClass={'m-0'}>Filters</Text>
                                {renderFilters()}
                                {renderSelectFilter()}
                                {renderCombinationOperator()}
                            </Col> 
                    </Row>

                    <Row className={'mt-8'}>
                        <Col span={24}>
                            <div className="flex justify-end">
                                <Button size={'large'} disabled={false} onClick={handleCancel}>Cancel</Button>
                            <Button size={'large'} disabled={false}  className={'ml-2'} type={'primary'}  htmlType="submit">Save</Button>
                        </div>
                        </Col>
                    </Row>
            </Form>
        </Modal>
    )

}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
});


export default connect(mapStateToProps, {})(PropertyValueModal);
