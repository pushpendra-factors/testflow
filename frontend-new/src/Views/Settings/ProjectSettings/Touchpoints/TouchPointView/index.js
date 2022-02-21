import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';
import { Text, SVG } from 'factorsComponents';
import {
    Row, Col, Button, Radio, Input, Select
} from 'antd';

import FaFilterSelect from 'Components/FaFilterSelect';
import { DEFAULT_OPERATOR_PROPS } from 'Components/FaFilterSelect/utils';

import { fetchEventPropertyValues } from 'Reducers/coreQuery/services';
import FaSelect from '../../../../../components/FaSelect';

import {
    getFilters, getStateFromFilters
} from '../../../../../Views/CoreQuery/utils';

const TouchpointView = ({ activeProject, tchType = '2', eventProperties, userProperties, rule, onCancel, onSave }) => {
    const [dropDownValues, setDropDownValues] = useState({});
    const [filterDD, setFilterDD] = useState(false);
    const [timestampRef, setTimestampRefState] = useState('LAST_MODIFIED_TIME_REF');
    //touch_point_time_ref
    const [touchPointPropRef, setTouchPointPropRef] = useState('');
    const [timestampPropertyRef, setTimestampPropRef] = useState(false);
    const [dateTypeDD, setDateTypeDD] = useState(false);
    const [dateTypeProps, setDateTypeProps] = useState([]);
    //filters
    const [newFilterStates, setNewFilterStates] = useState([]);

    //property map
    const [propertyMap, setPropertyMap] = useState({
        "$campaign": {
            "ty": "Property",
            "va": ""
        },
        "$channel": {
            "ty": "Constant",
            "va": ""
        },
        "$source": {
            "ty": "Constant",
            "va": ""
        },
        "$type": {
            "ty": "Constant",
            "va": ""
        }
    });

    const [filterDropDownOptions, setFiltDD] = useState({
        props: [
            {
                label: 'event',
                icon: 'event',
            },
        ],
        operator: DEFAULT_OPERATOR_PROPS,
    });

    useEffect(()=>{
        if(rule) {
            const filterState = getStateFromFilters(rule.filters);
            setNewFilterStates(filterState);
            setPropertyMap(rule.properties_map);
            if(rule.touchPointPropRef === 'LAST_MODIFIED_TIME_REF') {
                setTimestampRefState('LAST_MODIFIED_TIME_REF');
                setTimestampPropRef(false);
                setTouchPointPropRef('LAST_MODIFIED_TIME_REF');
            } else {
                setTimestampRefState(``);
                setTouchPointPropRef(rule.touch_point_time_ref)
                setTimestampPropRef(true);
                setDateTypeDD(false);
            }
        }
        
    }, [rule])


    const setValuesByProps = (props) => {
        const eventToCall = tchType === '2' ? 
            '$hubspot_contact_updated' : timestampRef === 'campaign_member_created_date'? '$sf_campaign_member_created' :  '$sf_campaign_member_updated';
        fetchEventPropertyValues(activeProject.id, eventToCall, props[1]).then(res => {
            const ddValues = Object.assign({}, dropDownValues);
            ddValues[props[1]] = [...res.data, '$none'];
            setDropDownValues(ddValues);
        }).catch(err => {
            const ddValues = Object.assign({}, dropDownValues);
            ddValues[props[0]] = ['$none'];
            setDropDownValues(ddValues);
        });
    }

    useEffect(() => {
        const eventToCall = tchType === '2' ? 
            '$hubspot_contact_updated' : timestampRef === 'campaign_member_created_date'? '$sf_campaign_member_created' :  '$sf_campaign_member_updated';
        const tchUserProps = [];
        const filterDD = Object.assign({}, filterDropDownOptions);
        const propState = [];
        if(tchType === '2') {
            userProperties.forEach((prop) => {if(prop[1]?.startsWith('$hubspot')){tchUserProps.push(prop)}})
        } else if(tchType === '3') {
            userProperties.forEach((prop) => {if(prop[1]?.startsWith('$salesforce')){tchUserProps.push(prop)}})
        }
        filterDropDownOptions.props.forEach((k) => {
            propState.push({ label: k.label, icon: 'event', values: [...eventProperties[eventToCall], ...tchUserProps] });
        });

        const dateTypepoperties = [];
        eventProperties[eventToCall]?.forEach((prop) => { if (prop[2] === 'datetime') { dateTypepoperties.push(prop) } });
        tchUserProps.forEach((prop) => { if (prop[2] === 'datetime') { dateTypepoperties.push(prop)}});
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
            <Col span={9} className={`justify-items-end`}>
                <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>Select {tchType === '2'? `Hubspot Contact`: 'Salesforce Campaign Member'} fields</Text>
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
        for(let i =0;i<propKeys.length; i++) {
            propertyMap[propKeys[i]]['va']? null : isReady = false;
            if(!isReady) {break}
        }
        return isReady;
    }

    const renderTimestampRenderOption = () => {
        let radioGroupElement = null;
        if(tchType === '2') {
            radioGroupElement = (<Radio.Group onChange={setTimestampRef} value={timestampRef}>
                <Radio value={`LAST_MODIFIED_TIME_REF`}>Factors Last modified time</Radio>
                <Radio value={``}>Select a property</Radio>
            </Radio.Group>)
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
                <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>Touchpoint Timestamp</Text>
            </Row>
            <Row className={`mt-4`}>
                {renderTimestampRenderOption()}
            </Row>
            <Row className={`mt-2`}>
                {timestampPropertyRef &&
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
        const propMap = Object.assign({}, propertyMap);
        propertyMap['$source']['va'] = val;
        setPropertyMap(propMap);
    }

    const setPropCampaign = (val) => {
        const propMap = Object.assign({}, propertyMap);
        propertyMap['$campaign']['va'] = val;
        setPropertyMap(propMap);
    }

    const setPropChannel = (val) => {
        const propMap = Object.assign({}, propertyMap);
        propertyMap['$channel']['va'] = val?.target?.value;
        setPropertyMap(propMap);
    }

    const renderEventPropertyCampOptions = () => {
        const eventToCall = tchType === '2' ? 
            '$hubspot_contact_updated' : timestampRef === 'campaign_member_created_date'? '$sf_campaign_member_created' :  '$sf_campaign_member_updated';
        const propertiesMp = eventProperties[eventToCall]? [...eventProperties[eventToCall]]: [];
        if(tchType === '2') {
            userProperties.forEach((prop) => {if(prop[1]?.startsWith('$hubspot')){propertiesMp.push(prop)}})
        } else if(tchType === '3') {
            userProperties.forEach((prop) => {if(prop[1]?.startsWith('$salesforce')){propertiesMp.push(prop)}})
        }
        return propertiesMp?.map((prop) => {
            return (<Option value={prop[1]}> {prop[0]} </Option>);
        });
    }

    const renderPropertyMap = () => {
        return (<div className={`border-top--thin pt-5 mt-8 `}>
            <Row>
                <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>Map the properties</Text>
            </Row>
            <Row className={`mt-10`}>
                <Col span={7}>
                    <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Type</Text>
                </Col>

                <Col>
                    <Select className={'fa-select w-full'} size={'large'} value={propertyMap['$type']['va']} onSelect={setPropType} defaultValue={``}>
                        <Option value={``}>Select Type </Option>
                        <Option value="tactic">Tactic</Option>
                        <Option value="offer">Offer</Option>
                    </Select>
                </Col>
            </Row>

            <Row className={`mt-4`}>
                <Col span={7}>
                    <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Source</Text>
                </Col>

                <Col>
                    {   tchType === '2'?
                        <Input value={propertyMap['$source']['va']} onChange={setPropSource}></Input>
                        :
                        <Select className={'fa-select w-full'} size={'large'} value={propertyMap['$source']['va']} onSelect={setPropSource} defaultValue={``}>
                            <Option value={``}>Select Source Property </Option>
                            {renderEventPropertyCampOptions()}
                        </Select>
                    }
                </Col>
            </Row>

            <Row className={`mt-4`}>
                <Col span={7}>
                    <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Campaign</Text>
                </Col>

                <Col>
                    <Select className={'fa-select w-full'} size={'large'} value={propertyMap['$campaign']['va']} onSelect={setPropCampaign} defaultValue={``}>
                        <Option value={``}>Select Campaign Property </Option>
                        {renderEventPropertyCampOptions()}
                    </Select>
                </Col>
            </Row>

            <Row className={`mt-4`}>
                <Col span={7}>
                    <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Channel</Text>
                </Col>

                <Col>
                    <Input value={propertyMap['$channel']['va']} onChange={setPropChannel}></Input>
                </Col>
            </Row>

        </div>);
    }

    const onSaveToucPoint = () => {
        // Prep settings obj;

        const touchPointObj = {
            //parse and set filterstate
            "filters": getFilters(newFilterStates),
            // set propMap
            "properties_map": propertyMap,
            "touch_point_time_ref": touchPointPropRef
        }
        onSave(touchPointObj);
    }

    const renderFooterActions = () => {
        return (
            <div>
                <Row  className={`mt-24 relative justify-end`}>
                    <Col span={10}>
                        <Button size={'large'} onClick={() => onCancel()}>Cancel</Button>
                        <Button  size={'large'} type="primary" className={'ml-2'}
                            htmlType="submit" onClick={onSaveToucPoint}
                        >Save</Button>
                    </Col>
                </Row>
            </div>
        )
    }

    return (
        <div>

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

export default connect(mapStateToProps, {})(TouchpointView);