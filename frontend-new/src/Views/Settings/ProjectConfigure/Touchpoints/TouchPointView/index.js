import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Text, SVG } from 'factorsComponents';
import {
    Row, Col, Button, Radio, Input, Select, Tooltip
} from 'antd';

import { getEventProperties } from 'Reducers/coreQuery/middleware';

import FaFilterSelect from 'Components/FaFilterSelect';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';

import { fetchEventPropertyValues } from 'Reducers/coreQuery/services';
import FaSelect from '../../../../../components/FaSelect';

import {
    getFiltersWithoutOrProperty, getStateFromFilters
} from '../../../../../Views/CoreQuery/utils';

const TouchpointView = ({ activeProject, tchType = '2', getEventProperties, eventProperties, userProperties, rule, onCancel, onSave }) => {
    const [dropDownValues, setDropDownValues] = useState({});
    const [filterDD, setFilterDD] = useState(false);

    const [tchRuleType, setTchRuleType] = useState('hubspot_contact_fields');

    const [timestampRef, setTimestampRefState] = tchType === '2' ? useState('LAST_MODIFIED_TIME_REF') : useState('campaign_member_created_date');
    //touch_point_time_ref
    const [touchPointPropRef, setTouchPointPropRef] = tchType === '2' ? useState('LAST_MODIFIED_TIME_REF') : useState('campaign_member_created_date');
    const [timestampPropertyRef, setTimestampPropRef] = useState(false);
    const [dateTypeDD, setDateTypeDD] = useState(false);
    const [dateTypeProps, setDateTypeProps] = useState([]);
    //filters
    const [newFilterStates, setNewFilterStates] = useState([]);

    //Search Keys 
    const [searchSour, setSearchSour] = useState({
        'source': '',
        'campaign': '',
        'channel': ''
    });

    //property map
    const [propertyMap, setPropertyMap] = useState({
        "$campaign": {
            "ty": "Property",
            "va": ""
        },
        "$channel": {
            "ty": "Property",
            "va": ""
        },
        "$source": {
            "ty": "Property",
            "va": ""
        },
        "$type": {
            "ty": "Property",
            "va": ""
        }
    });

    const [filterDropDownOptions, setFiltDD] = useState({
        props: [
            {
                label: '',
                icon: 'event',
            },
        ],
        operator: DEFAULT_OPERATOR_PROPS,
    });

    useEffect(() => {
        if (rule) {
            const filterState = getStateFromFilters(rule.filters);
            chainEventPropertyValues(filterState)
            setNewFilterStates(filterState);
            setPropertyMap(rule.properties_map);
            if (rule.touchPointPropRef === 'LAST_MODIFIED_TIME_REF') {
                setTimestampRefState('LAST_MODIFIED_TIME_REF');
                setTimestampPropRef(false);
                setTouchPointPropRef('LAST_MODIFIED_TIME_REF');
            } else {
                setTimestampRefState(``);
                setTouchPointPropRef(rule.touch_point_time_ref)
                setTchRuleType(rule.rule_type);
                setTimestampPropRef(true);
                setDateTypeDD(false);
            }
        }

    }, [rule])

    useEffect(() => {
        if(tchType === '2') {
            const eventToCall = getEventToCall();
            getEventProperties(activeProject.id, eventToCall);
        }
    }, [tchRuleType])

    const chainEventPropertyValues = (filters) => {
        const eventToCall = returnEventToCall();
        filters.forEach((filt) => {
            const prop = filt.props;
            const propToCall = prop.length > 3? prop[1] : prop[0];
            const propCallBack = (data) => setPropData(propToCall, data);
            console.log(propToCall);
            fetchEventPropertyValues(activeProject.id, eventToCall, propToCall).then(res => {
                propCallBack(res.data)
            });
        });   
    }

    const returnEventToCall = () => {
        return tchType === '2' ?
        getEventToCall() : timestampRef === 'campaign_member_created_date' ? '$sf_campaign_member_created' : '$sf_campaign_member_updated';
    }

    const setPropData = (propToCall, data) => {
        const ddValues = Object.assign({}, dropDownValues);
        ddValues[propToCall] = [...data, '$none'];
        setDropDownValues(ddValues);
    }
    

    const setValuesByProps = (props) => {
        const eventToCall = returnEventToCall();
        const propToCall = props.length > 3? props[1] : props[0];
        if(dropDownValues[propToCall]?.length >= 1) {
            return null;
        }
        fetchEventPropertyValues(activeProject.id, eventToCall, propToCall).then(res => {
            setPropData(propToCall, res.data);
        }).catch(err => {
            const ddValues = Object.assign({}, dropDownValues);
            ddValues[propToCall] = ['$none'];
            setDropDownValues(ddValues);
        });
    }

    const getEventToCall = () => {
        if(tchRuleType === 'Emails') {
            return '$hubspot_engagement_email';
        }
        else if(tchRuleType === 'hubspot_contact_fields') {
            return '$hubspot_contact_updated';
        }
    }

    useEffect(() => {
        const eventToCall = tchType === '2' ?
        getEventToCall() : timestampRef === 'campaign_member_created_date' ? '$sf_campaign_member_created' : '$sf_campaign_member_updated';
        const tchUserProps = [];
        const filterDD = Object.assign({}, filterDropDownOptions);
        const propState = [];
        const eventProps = [];
        if (tchType === '2') {
            const startsWith = tchRuleType === 'Emails'? '$hubspot_engagement' : '$hubspot_contact';  
            eventProperties[eventToCall] ?
                eventProperties[eventToCall].forEach((prop) => { if (prop[1]?.startsWith(startsWith)) { eventProps.push(prop) } }) : null;
            userProperties.forEach((prop) => { if (prop[1]?.startsWith(startsWith)) { tchUserProps.push(prop) } })
        } else if (tchType === '3') {
            eventProperties[eventToCall] ?
                eventProperties[eventToCall].forEach((prop) => { if (prop[1]?.startsWith('$salesforce_campaign')) { eventProps.push(prop) } }) : null;
            userProperties.forEach((prop) => {
                if (prop[1]?.startsWith('$salesforce_campaign')) { tchUserProps.push(prop) }
            });
        }

        filterDropDownOptions.props.forEach((k) => {
            propState.push({ label: k.label, icon: 'event', values: [...eventProps, ...tchUserProps] });
        });

        const dateTypepoperties = [];
        eventProps.forEach((prop) => { if (prop[2] === 'datetime') { dateTypepoperties.push(prop) } });
        tchUserProps.forEach((prop) => { if (prop[2] === 'datetime') { dateTypepoperties.push(prop) } });
        setDateTypeProps(dateTypepoperties);
        filterDD.props = propState;
        setFiltDD(filterDD);

    }, [eventProperties, timestampRef]);

    const applyFilter = (fil, index) => {

        const filtState = [...newFilterStates];
        if (index && index < 0) {
            filtState.push(fil);
        } else {
            filtState[index] = fil;
        }
        setNewFilterStates(filtState);
        setFilterDD(false);
    };

    const closeFilter = (index) => {
        const filtrs = [...newFilterStates];
        filtrs.splice(index, 1);
        setNewFilterStates(filtrs);
    }

    const addFilter = () => { console.log("Added") }

    const renderFilters = () => {
        const filterRows = []
        if (newFilterStates) {
            newFilterStates.forEach((filter, index) => {
                filterRows.push((
                    <Row className={`mt-2`}>
                        <FaFilterSelect
                            filter={filter}
                            propOpts={filterDropDownOptions.props}
                            operatorOpts={filterDropDownOptions.operator}
                            valueOpts={dropDownValues}
                            applyFilter={(filt) => applyFilter(filt, index)}
                            setValuesByProps={setValuesByProps}
                        >

                        </FaFilterSelect>
                        <Button className={`ml-2`} icon={<SVG name={'remove'} />} onClick={() => closeFilter(index)}></Button>
                    </Row>
                ))
            })
        }

        filterRows.push((<Row className={`mt-2`}>
            {filterDD ?
                <>
                    <FaFilterSelect
                        propOpts={filterDropDownOptions.props}
                        operatorOpts={filterDropDownOptions.operator}
                        valueOpts={dropDownValues}
                        applyFilter={(filt) => applyFilter(filt, -1)}
                        setValuesByProps={setValuesByProps}
                    >

                    </FaFilterSelect>
                    <Button className={`ml-2`} icon={<SVG name={'remove'} />} onClick={() => setFilterDD(false)}></Button>
                </>
                :
                <Button size={'large'} type={'text'}
                    onClick={() => setFilterDD(true)}
                ><SVG name={'plus'}
                    extraClass={'mr-1'} />{'Add Filter'}
                </Button>
            }


        </Row>))

        return filterRows;
    }

    const renderFilterBlock = () => {
        return (<Row className={`mt-4`}>
            <Col span={6} className={`justify-items-start`}>
                <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>Add a Touchpoint Rule<sup>*</sup></Text>
            </Col>

            <Col span={14}>
                {renderFilters().map(component => component)}
            </Col>

        </Row>);
    }

    const setTimestampRef = (val) => {
        if (val?.target?.value === `LAST_MODIFIED_TIME_REF`) {
            setTimestampRefState('LAST_MODIFIED_TIME_REF');
            setTimestampPropRef(false);
            setTouchPointPropRef('LAST_MODIFIED_TIME_REF');
        } else {
            setTimestampRefState(``);
            setTouchPointPropRef('')
            setTimestampPropRef(true);
        }
    }

    const setTimestampRefSF = (val) => {
        const timeStVal = val?.target?.value;
        setTimestampRefState(timeStVal);
        setTimestampPropRef(false);
        setTouchPointPropRef(timeStVal);
    }

    const setTimestampProp = (val) => {
        setTouchPointPropRef(val[1]);
        setDateTypeDD(false);
    }

    const validateInputs = () => {
        let isReady = true;
        const propKeys = Object.keys(propertyMap);
        for (let i = 0; i < propKeys.length; i++) {
            propertyMap[propKeys[i]]['va'] ? null : isReady = false;
            if (!isReady) { break }
        }
        return isReady;
    }

    const validateRuleInfo = () => {
        if (newFilterStates.length < 1 || !touchPointPropRef) {
            return false;
        }
        return true;
    }

    const setTimestampRefEmail = (val) => {
        setTimestampRefState(val);
        setTimestampPropRef(false);
        setTouchPointPropRef(val);
    }

    const getTimestampOptionByRule = () => {
        if(tchRuleType === 'hubspot_contact_fields') {
            return (<Radio.Group onChange={setTimestampRef} value={timestampRef}>
                <Radio value={`LAST_MODIFIED_TIME_REF`}>Factors Last modified time</Radio>
                <Radio value={``}>Select a property</Radio>
            </Radio.Group>);
        }
        else if(tchRuleType === 'Emails') {
            return (<Radio.Group onChange={() => setTimestampRefEmail('$hubspot_engagement_timestamp')} value={timestampRef} defaultValue={`$hubspot_engagement_timestamp`}>
                <Radio value={`$hubspot_engagement_timestamp`}>Email Timestamp</Radio>
                {/* <Radio value={`email_replied_timestamp`}>Email Replied Timestamp</Radio> */}
            </Radio.Group>);
        }
    }

    const renderTimestampRenderOption = () => {
        let radioGroupElement = null;
        if (tchType === '2') {
            radioGroupElement = getTimestampOptionByRule();
        }
        else if (tchType === '3') {
            radioGroupElement = (<Radio.Group onChange={setTimestampRefSF} value={timestampRef}>
                <Radio value={`campaign_member_created_date`}>Campaign Created Date</Radio>
                <Radio value={`campaign_member_first_responded_date`}>Campaign First Responded Date</Radio>
            </Radio.Group>)
        }

        return radioGroupElement;
    }

    const renderTimestampSelector = () => {
        return (<div className={`mt-8`}>
            <Row className={`mt-2`}>
                <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>Touchpoint Timestamp<sup>*</sup></Text>
            </Row>
            <Row className={`mt-4`}>
                {renderTimestampRenderOption()}
            </Row>
            <Row className={`mt-2`}>
                {tchRuleType === 'hubspot_contact_fields' && timestampPropertyRef &&
                    <div className={`relative`}>
                        <Button type='link' onClick={() => setDateTypeDD(!dateTypeDD)}>
                            {touchPointPropRef ? touchPointPropRef : 'Select Date type property'}
                        </Button>
                        {dateTypeDD && <FaSelect
                            optionClick={(opt) => setTimestampProp(opt)}
                            onClickOutside={() => setDateTypeDD(false)}
                            options={dateTypeProps} >
                        </FaSelect>}
                    </div>
                }
            </Row>
        </div>);
    }

    const setPropType = (val) => {
        const propMap = Object.assign({}, propertyMap);
        propertyMap['$type']['va'] = val;
        setPropertyMap(propMap);
    }

    const setPropSource = (val) => {
        let propMap = Object.assign({}, propertyMap);
        propertyMap['$source']['va'] = val;
        if (val === searchSour['source']) {
            propMap = setSearchValue('source', propMap);
        }
        setSearchSour({ 'source': '', 'campaign': '', 'channel': '' });
        setPropertyMap(propMap);
    }

    const setSearchValue = (type, propMap) => {
        propertyMap['$' + type]['va'] = searchSour[type];
        propMap['$' + type]['ty'] = 'Constant';
        return propMap;
    }

    const setPropCampaign = (val) => {
        setSearchSour({ 'source': '', 'campaign': '', 'channel': '' });
        let propMap = Object.assign({}, propertyMap);
        propertyMap['$campaign']['va'] = val;
        if (val === searchSour['campaign']) {
            propMap = setSearchValue('campaign', propMap);
        }
        setPropertyMap(propMap);
    }

    const setPropChannel = (val) => {
        let propMap = Object.assign({}, propertyMap);
        propertyMap['$channel']['va'] = val;
        if (val === searchSour['channel']) {
            propMap = setSearchValue('channel', propMap);
        }
        setPropertyMap(propMap);
    }

    const isSearchProps = (dropDownType, prop) => {
        if (dropDownType && searchSour[dropDownType] && prop[1]?.search(searchSour[dropDownType])) {
            return true;
        }
        if (dropDownType && !searchSour[dropDownType]) {
            return true;
        }
        return false;
    }

    const propOption = (item) => {
        return (<Tooltip title={item} placement={'right'}>
            <div style={{ width: "210px" }}>
                <div
                    style={{
                        maxWidth: "200px",
                        overflow: "hidden",
                        whiteSpace: "nowrap",
                        textOverflow: "ellipsis"
                    }}
                >
                    {item}
                </div>
            </div> </Tooltip>);
    }

    const renderEventPropertyCampOptions = (dropDownType) => {
        const eventToCall = tchType === '2' ?
            getEventToCall() : timestampRef === 'campaign_member_created_date' ? '$sf_campaign_member_created' : '$sf_campaign_member_updated';
        const propertiesMp = [];
        if (tchType === '2') {
            const startsWith = tchRuleType === 'Emails'? '$hubspot_engagement' : '$hubspot_contact';
            eventProperties[eventToCall]?.forEach((prop) => {
                if (prop[1]?.startsWith(startsWith) && isSearchProps(dropDownType, prop)) {
                    propertiesMp.push(
                        <Option key={prop[1]} value={prop[1]}> {propOption(prop[0])}  </Option>
                    );
                }
            });
            userProperties.forEach((prop) => {
                if (prop[1]?.startsWith(startsWith) && isSearchProps(dropDownType, prop)) {
                    propertiesMp.push(
                        <Option key={prop[1]} value={prop[1]}> {propOption(prop[0])} </Option>
                    );
                }
            });
        } else if (tchType === '3') {
            eventProperties[eventToCall]?.forEach((prop) => {
                if (prop[1]?.startsWith('$salesforce') && isSearchProps(dropDownType, prop)) {
                    propertiesMp.push(
                        <Option key={prop[1]} value={prop[1]}> {propOption(prop[0])}  </Option>);
                }
            });
            userProperties.forEach((prop) => {
                if (prop[1]?.startsWith('$salesforce') && isSearchProps(dropDownType, prop)) {
                    propertiesMp.push(
                        <Option key={prop[1]} value={prop[1]}> {propOption(prop[0])}  </Option>);
                }
            });
        }
        if (dropDownType && searchSour[dropDownType]) {
            propertiesMp.push((<Option value={searchSour[dropDownType]}> <span>Select: </span> {searchSour[dropDownType]} </Option>));
        }
        return propertiesMp;
    }

    const setSearch = (key, val) => {
        const srch = Object.assign({}, searchSour);
        srch[key] = val;
        setSearchSour(srch);
    }

    const renderPropertyMap = () => {
        return (<div className={`border-top--thin pt-5 mt-8 `}>
            <Row>
                <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>Map the properties<sup>*</sup></Text>
            </Row>
            <Row className={`mt-10`}>
                <Col span={7}>
                    <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Type</Text>
                </Col>
                <Col>
                    <Select
                        className={'fa-select w-full'}
                        size={'large'} value={propertyMap['$type']['va']} onSelect={setPropType} defaultValue={``}

                    >
                        <Option value={``}>Select Type </Option>
                        <Option value="tactic">Tactic</Option>
                        {tchRuleType!=='Emails' && <Option value="offer">Offer</Option>}
                    </Select>
                </Col>
            </Row>

            <Row className={`mt-4`}>
                <Col span={7}>
                    <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Source</Text>
                </Col>

                <Col>
                    {
                        <Select
                            showSearch
                            onSearch={(val) => setSearch('source', val)}
                            className={'fa-select w-full'} size={'large'}
                            value={propertyMap['$source']['va']} onSelect={setPropSource}
                            defaultValue={``} style={{ minWidth: '200px', maxWidth: '210px' }}>
                            {searchSour['source'] ? null :
                                <Option value={``}>Select Source Property </Option>
                            }
                            {renderEventPropertyCampOptions('source')}
                        </Select>
                    }
                </Col>
            </Row>

            <Row className={`mt-4`}>
                <Col span={7}>
                    <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Campaign</Text>
                </Col>

                <Col>
                    <Select
                        showSearch
                        onSearch={(val) => setSearch('campaign', val)}
                        className={'fa-select w-full'} style={{ minWidth: '200px', maxWidth: '210px' }}
                        size={'large'} value={propertyMap['$campaign']['va']} onSelect={setPropCampaign} defaultValue={``}>
                        {searchSour['campaign'] ? null :
                            <Option value={``}>Select Campaign Property </Option>
                        }

                        {renderEventPropertyCampOptions('campaign')}
                    </Select>
                </Col>
            </Row>

            <Row className={`mt-4`}>
                <Col span={7}>
                    <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Channel</Text>
                </Col>

                <Col>
                    <Select
                        showSearch
                        onSearch={(val) => setSearch('channel', val)}
                        className={'fa-select w-full'} style={{ minWidth: '200px', maxWidth: '210px' }}
                        size={'large'} value={propertyMap['$channel']['va']} onSelect={setPropChannel} defaultValue={``}>
                        {searchSour['channel'] ? null :
                            <Option value={``}>Select Channel Property </Option>
                        }

                        {renderEventPropertyCampOptions('channel')}
                    </Select>
                </Col>
            </Row>

        </div>);
    }

    const onSaveToucPoint = () => {
        // Prep settings obj;

        const touchPointObj = {
            //parse and set filterstate
            "filters": getFiltersWithoutOrProperty(newFilterStates),
            // set propMap
            "properties_map": propertyMap,
            "touch_point_time_ref": touchPointPropRef,
        }
        if(tchType === '2') {
            touchPointObj["rule_type"] = tchRuleType
        }
        onSave(touchPointObj);
    }

    const renderFooterActions = () => {
        return (
            <div>
                <Row className={`mt-20 relative justify-start`}>
                    <Text level={7} type={'title'} extraClass={'m-0 italic'} weight={'thin'}><sup>*</sup> All these fields are mandatory</Text>
                </Row>
                <Row className={`border-top--thin mt-4 relative justify-start`}>
                    <Col className={`mt-6`} span={10}>
                        <Button size={'large'} onClick={() => onCancel()}>Cancel</Button>
                        <Button disabled={!validateRuleInfo() || !validateInputs()} size={'large'} type="primary" className={'ml-2'}
                            htmlType="submit" onClick={onSaveToucPoint}
                        >Save</Button>
                    </Col>
                </Row>
            </div>
        )
    }

    const renderTchRuleTypeOptions = () => {
        return (
            <Col>
                    <Select
                        className={'fa-select w-64'}
                        size={'large'} value={tchRuleType} onSelect={setTchRuleType} defaultValue={``}

                    >
                        <Option value="Emails">Email</Option>
                        <Option value="hubspot_contact_fields">Change in Hubspot contact field value</Option>
                    </Select>
                </Col>
        )
    }

    const renderTchRuleType = () => {
        if(tchType === '3') {return;}
        return (<div className={`mt-8`}>
            <Row className={`mt-2`}>
                <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Create a touchpoint using<sup>*</sup></Text>
            </Row>
            <Row className={`mt-4`}>
                {renderTchRuleTypeOptions()}
            </Row>
        </div>
        )
    }

    return (
        <div>

            {renderTchRuleType()}

            {renderTimestampSelector()}

            {renderFilterBlock()}

            {renderPropertyMap()}

            {renderFooterActions()}
        </div>
    )
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    eventProperties: state.coreQuery.eventProperties,
    userProperties: state.coreQuery.userProperties
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getEventProperties,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(TouchpointView);