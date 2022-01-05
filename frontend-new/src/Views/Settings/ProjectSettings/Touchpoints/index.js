
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

    const [touchPointState, setTouchPointState] = useState({state: 'list', index: 0});

    const columns = [
        {
            title: tabNo === "2"? 'Hubspot Object' : 'Salesforce Object',
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
        // {
        //     title: '',
        //     dataIndex: 'index',
        //     key: 'index',
        //     render: (obj) => { return renderTableActions(obj) }
        // },

    ];

    function callback(key) {
        setTabNo(key);
    }

    useEffect(() => {
        if(tabNo === "2") {
            setHubspotContactDate();
        }
        if(tabNo === "3") {
            setSalesforceContactData();
        }
        
    }, [activeProject, tabNo]);

    const setSalesforceContactData = () => {
        const touchpointObjs = activeProject['salesforce_touch_points'] && activeProject['salesforce_touch_points']['sf_touch_point_rules'] ? 
            [...activeProject['salesforce_touch_points']['sf_touch_point_rules'].map((rule, id) => ({...rule, 'index': id}))] 
            : [];
        setTouchPointsData(touchpointObjs);

        getEventProperties(activeProject.id, '$sf_contact_updated')
    }

    const setHubspotContactDate = () => {
        const touchpointObjs = activeProject['hubspot_touch_points'] && activeProject['hubspot_touch_points']['hs_touch_point_rules'] ? 
            [...activeProject['hubspot_touch_points']['hs_touch_point_rules'].map((rule, id) => ({...rule, 'index': id}))] 
            : [];
        setTouchPointsData(touchpointObjs);

        getEventProperties(activeProject.id, '$hubspot_contact_updated')
    }

    const renderTableActions = (index) => {
        return (<Button
            type='text'
            onClick={() => setTouchPointState({state: 'edit', index: index})}
            className={`fa-btn--custom ml-1 mr-1`}
        >
            <SVG name='more'></SVG>
        </Button>)
    }

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

                    <Col className={`fa-truncate-150`}>
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

                    <Col className={`fa-truncate-150`}>
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

                    <Col className={`fa-truncate-150`}>
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

                    <Col className={`fa-truncate-150`}>
                        <Text level={7} type={'title'} extraClass={'ml-4'} weight={'thin'}>{obj['$channel']['va']}</Text>
                    </Col>
                </Row>}
            </Col>
        )
    }

    const renderTitle = () => {
        let title = null;
        if (touchPointState.state === 'list') {
            title = (<Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Touchpoints</Text>);
        }
        if (touchPointState.state === 'add') {
            title = (<Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Add new Touchpoint</Text>);
        }
        return title;
    }

    const renderTitleActions = () => {
        let titleAction = null;
        if (touchPointState.state === 'list') {
            if (tabNo !== "1") {
                titleAction = (
                    <Button size={'large'} onClick={() => {
                        setTouchPointState({state: 'add', index: 0})
                    }}><SVG name={'plus'} extraClass={'mr-2'} size={16} />Add New</Button>)
            }
        }

        return titleAction;
    }

    const onTchSave = (tchObj) => {
        let tchPointRules = [];
        if(tabNo === '2') {
            tchPointRules = activeProject['hubspot_touch_points'] && activeProject['hubspot_touch_points']['hs_touch_point_rules'] ? [...activeProject['hubspot_touch_points']['hs_touch_point_rules']] : [];
            tchPointRules.push(tchObj);
            udpateProjectDetails(activeProject.id, { 'hubspot_touch_points': { 'hs_touch_point_rules': tchPointRules } });
        }
        else if (tabNo === '3') {
            tchPointRules = activeProject['salesforce_touch_points'] && activeProject['salesforce_touch_points']['sf_touch_point_rules'] ? [...activeProject['salesforce_touch_points']['sf_touch_point_rules']] : [];
            tchPointRules.push(tchObj);
            udpateProjectDetails(activeProject.id, { 'salesforce_touch_points': { 'sf_touch_point_rules': tchPointRules } });
        }
        fetchProjects();
        setTouchPointState({state: 'list', index: 0});
    }

    const onTchCancel = () => {
        setTouchPointState({state: 'list', index: 0});
    }

    const renderTouchPointContent = () => {
        let touchPointContent = null;
        if (touchPointState.state === 'list') {
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

                <TabPane tab="Salesforce" key="3">
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
        else if (touchPointState.state === 'add') {
            touchPointContent = (<TouchpointView tchType={tabNo} rule={null} onSave={onTchSave} onCancel={onTchCancel}> </TouchpointView>)
        }
        // else if (touchPointState.state === 'edit') {
        //     touchPointContent = (<TouchpointView rule={touchPointsData[touchPointState.index]} onSave={onTchSave} onCancel={onTchCancel}> </TouchpointView>)
        // }
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