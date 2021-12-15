
import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from 'factorsComponents';
import {
    getEventProperties,
} from 'Reducers/coreQuery/middleware';
import { fetchProjects, udpateProjectDetails } from "Reducers/global";
import {
    Row, Col, Button, Tabs, Table, Dropdown, Menu, notification, Tooltip
} from 'antd';
import TouchpointView from './TouchPointView';
import MarketingInteractions from '../MarketingInteractions';
import FAFilterSelect from '../../../../components/FaFilterSelect';

import {
    reverseOperatorMap, reverseDateOperatorMap
} from '../../../../Views/CoreQuery/utils';


const { TabPane } = Tabs;

const Touchpoints = ({ activeProject, currentProjectSettings, getEventProperties, fetchProjects, udpateProjectDetails }) => {

    const [tabNo, setTabNo] = useState("1");

    const [touchPointsData, setTouchPointsData] = useState([]);

    const [touchPointState, setTouchPointState] = useState('list');

    const columns = [
        {
            title: 'Hubspot Object',
            dataIndex: 'filters',
            key: 'filters',
            render: (obj) => { return renderObjects(obj) }
        },
        {
            title: 'Property Mapping',
            dataIndex: 'properties_map',
            key: 'properties_map',
            render: (obj) => { return renderPropertyMap(obj) }
        },

    ];

    function callback(key) {
        setTabNo(key);
    }

    useEffect(() => {
        const touchpointObjs = activeProject['hubspot_touch_points'] && activeProject['hubspot_touch_points']['hs_touch_point_rules'] ? [...activeProject['hubspot_touch_points']['hs_touch_point_rules']] : [];
        setTouchPointsData(touchpointObjs);

        getEventProperties(activeProject.id, '$hubspot_contact_updated')
    }, [activeProject])

    useEffect(() => {
        console.log(currentProjectSettings);
    }, [currentProjectSettings])

    const renderObjects = (obj) => {
        const filters = [];
        obj?.forEach((filterObj, ind) => {
            if (filterObj.lop === 'AND') {
                filters.push({
                    operator:
                        filterObj.ty === 'datetime'
                            ? reverseDateOperatorMap[filterObj.op]
                            : reverseOperatorMap[filterObj.op],
                    props: [filterObj.pr, filterObj.ty ? filterObj.ty : 'categorical', filterObj.en ? filterObj.en : 'event'],
                    values: [filterObj.va],
                });
            } else {
                filters[filters.length - 1].values.push(filterObj.va);
            }
        });
        return filters.map((filt) => (<div className={`mt-2 max-w-3xl`}>
            <FAFilterSelect filter={filt} disabled={true} applyFilter={() => { }}></FAFilterSelect>
        </div>));
    }

    const renderPropertyMap = (obj) => {
        return (
            <Col>
                {obj['$type'] && obj['$type']['va'] && <Row>
                    <Col span={10} >
                        <Row className={'relative justify-between'}>
                            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Type</Text>
                            <SVG name={`ChevronRight`} />
                        </Row>
                    </Col>

                    <Col>
                        <Text level={7} type={'title'} extraClass={'ml-4'} weight={'thin'}>{obj['$type']['va']}</Text>
                    </Col>
                </Row>}

                {obj['$source'] && obj['$source']['va'] && <Row>
                    <Col span={10} >
                        <Row className={'relative justify-between'}>
                            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Source</Text>
                            <SVG name={`ChevronRight`} />
                        </Row>
                    </Col>

                    <Col>
                        <Text level={7} type={'title'} extraClass={'ml-4'} weight={'thin'}>{obj['$source']['va']}</Text>
                    </Col>
                </Row>}

                {obj['$campaign'] && obj['$campaign']['va'] && <Row>
                    <Col span={10} >
                        <Row className={'relative justify-between'}>
                            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Campaign</Text>
                            <SVG name={`ChevronRight`} />
                        </Row>
                    </Col>

                    <Col>
                        <Text level={7} type={'title'} extraClass={'ml-4'} weight={'thin'}>{obj['$campaign']['va']}</Text>
                    </Col>
                </Row>}

                {obj['$channel'] && obj['$channel']['va'] && <Row>
                    <Col span={10} >
                        <Row className={'relative justify-between'}>
                            <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Channel</Text>
                            <SVG name={`ChevronRight`} />
                        </Row>
                    </Col>

                    <Col>
                        <Text level={7} type={'title'} extraClass={'ml-4'} weight={'thin'}>{obj['$channel']['va']}</Text>
                    </Col>
                </Row>}
            </Col>
        )
    }

    const renderTitle = () => {
        let title = null;
        if (touchPointState === 'list') {
            title = (<Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Touchpoints</Text>);
        }
        if (touchPointState === 'add') {
            title = (<Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Add new Touchpoint</Text>);
        }
        return title;
    }

    const renderTitleActions = () => {
        let titleAction = null;
        if (touchPointState === 'list') {
            if (tabNo === "2") {
                titleAction = (
                    <Button size={'large'} onClick={() => {
                        setTouchPointState('add')
                    }}><SVG name={'plus'} extraClass={'mr-2'} size={16} />Add New</Button>)
            }
        }

        return titleAction;
    }

    const onTchSave = (tchObj) => {
        const tchPointRules = activeProject['hubspot_touch_points'] && activeProject['hubspot_touch_points']['hs_touch_point_rules'] ? [...activeProject['hubspot_touch_points']['hs_touch_point_rules']] : [];
        tchPointRules.push(tchObj);
        const projectDetails = { ...activeProject }
        udpateProjectDetails(activeProject.id, { 'hubspot_touch_points': { 'hs_touch_point_rules': tchPointRules } });
        fetchProjects();
        setTouchPointState('list');
    }

    const onTchCancel = () => {
        setTouchPointState('list');
    }

    const renderTouchPointContent = () => {
        let touchPointContent = null;
        if (touchPointState === 'list') {
            touchPointContent = (<Tabs activeKey={`${tabNo}`} onChange={callback} >
                <TabPane tab="Digital Marketing" key="1">
                    <MarketingInteractions />
                </TabPane>

                <TabPane tab="Hubspot" key="2">
                    <div className={`mb-10 pl-4 mt-10`}>
                        <Table className="fa-table--basic mt-4"
                            columns={columns}
                            dataSource={touchPointsData}
                            pagination={false}
                            loading={false}
                        />
                    </div>
                </TabPane>
            </Tabs>)
        }
        else if (touchPointState === 'add') {
            touchPointContent = (<TouchpointView onSave={onTchSave} onCancel={onTchCancel}> </TouchpointView>)
        }
        return touchPointContent;
    }

    return (<div>
        <>
            <Row>
                <Col span={12}>
                    {renderTitle()}
                </Col>
                <Col span={12}>
                    <div className={'flex justify-end'}>
                        {renderTitleActions()}
                    </div>
                </Col>
            </Row>
            <Row className={'mt-4'}>
                <Col span={24}>
                    <div className={'mt-6'}>
                        {renderTouchPointContent()}
                    </div>
                </Col>
            </Row>
        </>
    </div>);
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentProjectSettings: state.global.currentProjectSettings
});

const mapDispatchToProps = (dispatch) =>
    bindActionCreators(
        {
            getEventProperties,
            fetchProjects,
            udpateProjectDetails
        },
        dispatch
    );

export default connect(mapStateToProps, mapDispatchToProps)(Touchpoints);