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
    getFilters
} from '../../../../../Views/CoreQuery/utils';

const TouchpointView = ({ activeProject, eventProperties, filters = [], onCancel, onSave }) => {
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


    const setValuesByProps = (props) => {
        fetchEventPropertyValues(activeProject.id, '$hubspot_contact_updated', props[1]).then(res => {
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
        const filterDD = Object.assign({}, filterDropDownOptions);
        const propState = [];
        filterDropDownOptions.props.forEach((k) => {
            propState.push({ label: k.label, icon: 'event', values: eventProperties['$hubspot_contact_updated'] });
        });

        const dateTypepoperties = [];
        eventProperties['$hubspot_contact_updated']?.forEach((prop) => { if (prop[2] === 'datetime') { dateTypepoperties.push(prop) } }
        );
        setDateTypeProps(dateTypepoperties);
        filterDD.props = propState;
        setFiltDD(filterDD);

    }, [eventProperties]);

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
            <Col span={7} className={`justify-items-end`}>
                <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>Select Hubspot Contact field</Text>
            </Col>

            <Col span={12}>
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

    const setTimestampProp = (val) => {
        setTouchPointPropRef(val[1]);
        setDateTypeDD(false);
    }

    const renderTimestampSelector = () => {
        return (<div className={`mt-8`}>
            <Row className={`mt-2`}>
                <Text level={7} type={'title'} extraClass={'m-0'} weight={'bold'}>Touchpoint Timestamp</Text>
            </Row>
            <Row className={`mt-4`}>
                <Radio.Group onChange={setTimestampRef} value={timestampRef}>
                    <Radio value={`LAST_MODIFIED_TIME_REF`}>Factors Last modified time</Radio>
                    <Radio value={``}>Select a property</Radio>
                </Radio.Group>
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
        propertyMap['$source']['va'] = val?.target?.value;
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
                    <Input value={propertyMap['$source']['va']} onChange={setPropSource}></Input>
                </Col>
            </Row>

            <Row className={`mt-4`}>
                <Col span={7}>
                    <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>Campaign</Text>
                </Col>

                <Col>
                    <Select className={'fa-select w-full'} size={'large'} value={propertyMap['$campaign']['va']} onSelect={setPropCampaign} defaultValue={``}>
                        <Option value={``}>Select Campaign Property </Option>
                        {eventProperties['$hubspot_contact_updated']?.map((prop) => {
                            return (<Option value={prop[1]}> {prop[0]} </Option>)
                        })}
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
            <Row  className={`mt-24 relative justify-end`}>
                <Col span={10}>
                    <Button size={'large'} onClick={() => onCancel()}>Cancel</Button>
                    <Button size={'large'} type="primary" className={'ml-2'}
                        htmlType="submit" onClick={onSaveToucPoint}
                    >Save</Button>
                </Col>
            </Row>
        )
    }

    return (
        <div>
            {renderFilterBlock()}

            {renderTimestampSelector()}

            {renderPropertyMap()}

            {renderFooterActions()}
        </div>
    )
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    eventProperties: state.coreQuery.eventProperties
});

export default connect(mapStateToProps, {})(TouchpointView);