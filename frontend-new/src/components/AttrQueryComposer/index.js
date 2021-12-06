import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import styles from './index.module.scss';
import { SVG, Text } from '../../components/factorsComponents';
import ConversionGoalBlock from './ConversionGoalBlock';
import FaDatepicker from '../../components/FaDatepicker';
import ComposerBlock from '../QueryCommons/ComposerBlock';

import {
    fetchEventNames,
    getUserProperties,
    getEventProperties,
    setGoalEvent,
    setTouchPoint,
    setModels, setWindow,
    setLinkedEvents, setAttrDateRange,
    getCampaignConfigData
} from '../../reducers/coreQuery/middleware';
import { Button, Tooltip } from 'antd';
import MarkTouchpointBlock from './MarkTouchpointBlock';
import AttributionOptions from './AttributionOptions';
import LinkedEventsBlock from './LinkedEventsBlock';
import { QUERY_TYPE_EVENT } from '../../utils/constants';

const AttrQueryComposer = ({ activeProject,
    fetchEventNames, getEventProperties,
    userProperties, eventProperties,
    runAttributionQuery, eventGoal, setGoalEvent,
    touchPoint, setTouchPoint, models, setModels,
    window, setWindow, linkedEvents, setLinkedEvents,
    setAttrDateRange, dateRange,
    collapse = false, setCollapse
}) => {

    const [linkEvExpansion, setLinkEvExpansion] = useState(true);
    const [convGblockOpen, setConvGblockOpen] = useState(true);
    const [tchPointblockOpen, setTchPointblockOpen] = useState(true);
    const [criteriablockOpen, setCriteriablockOpen] = useState(true);

    useEffect(() => {
        if (activeProject && activeProject.id) {
            getCampaignConfigData(activeProject.id, "all_ads")
            fetchEventNames(activeProject.id);
            if (!userProperties.length) {
                getUserProperties(activeProject.id, 'analysis');
            }
        }
    }, [activeProject]);

    useEffect(() => {

        if (!eventProperties[eventGoal?.label]) {
            getEventProperties(activeProject.id, eventGoal.label);
        }
    }, [eventGoal]);

    useEffect(() => {
        linkedEvents.forEach((ev, index) => {
            if (!eventProperties[ev.label]) {
                getEventProperties(activeProject.id, ev.label);
            }
        })

    }, [linkedEvents]);

    const goalChange = (eventGoal) => {
        setGoalEvent(eventGoal);
    }

    const linkEventChange = (linkEvent, index) => {
        const currLinkedEvs = [...linkedEvents];
        if (index === undefined || index < 0) {
            currLinkedEvs.push(linkEvent);
        } else {
            currLinkedEvs[index] = linkEvent;
        }
        setLinkedEvents(currLinkedEvs);
    }

    const goalDel = () => {
        setGoalEvent({});
    }

    const linkEventDel = (index) => {
        const currLinkedEvs = linkedEvents.filter((ev, i) => i !== index);
        setLinkedEvents(currLinkedEvs);
    }


    const renderConversionBlock = () => {
        if (eventGoal) {
            return (
                <ConversionGoalBlock eventGoal={eventGoal}
                    eventGoalChange={goalChange}
                    delEvent={goalDel}
                >
                </ConversionGoalBlock>)
        } else {
            return (<ConversionGoalBlock></ConversionGoalBlock>)
        }
    }

    const renderMarkTouchpointBlock = () => {
        return (
            <MarkTouchpointBlock
                touchPoint={touchPoint}
                setTouchpoint={(tchPoint) => setTouchPoint(tchPoint)}
            >

            </MarkTouchpointBlock>)
    }

    const renderAttributionOptions = () => {
        return (<AttributionOptions
            models={models}
            setModelOpt={(val) => setModels(val)}
            window={window}
            setWindowOpt={(win) => setWindow(win)}
        >
        </AttributionOptions>);
    }

    const renderLinkedEvents = () => {

        const linkEventsList = [];
        if (linkedEvents && linkedEvents.length) {
            linkedEvents.forEach((ev, index) => {
                linkEventsList.push(
                    <LinkedEventsBlock
                        linkEvent={ev}
                        linkEventChange={(ev) => linkEventChange(ev, index)}
                        delLinkEvent={() => linkEventDel(index)}
                    >
                    </LinkedEventsBlock>
                )
            })

        }

        linkEventsList.push(<LinkedEventsBlock linkEventChange={(ev) => linkEventChange(ev, -1)}></LinkedEventsBlock>)

        return linkEventsList;

    }

    const toggleLinkEvExpansion = () => {
        if (models.length > 1) return null;
        setLinkEvExpansion(!linkEvExpansion);
    }

    const handleRunQuery = () => {
        runAttributionQuery(false)
    };

    const setDateRange = (ranges) => {
        const dtRange = Object.assign({}, dateRange);
        if (ranges && ranges.startDate) {
            if (Array.isArray(ranges.startDate)) {
                dtRange.from = ranges.startDate[0]
                dtRange.to = ranges.startDate[1];
            } else {
                dtRange.from = ranges.startDate;
                dtRange.to = ranges.endDate;
            }
        }
        setAttrDateRange(dtRange);
    }

    const footer = () => {
        if (!eventGoal || !eventGoal?.label?.length) { return null; }

        return (
            <div className={`${!collapse ? styles.composer__footer : styles.composer_footer_right}`}>
                {!collapse ? <FaDatepicker
                    customPicker
                    presetRange
                    buttonSize={`large`}
                    className={`mr-2`}
                    monthPicker
                    range={
                        {
                            startDate: dateRange.from,
                            endDate: dateRange.to
                        }
                    }
                    placement="topRight" onSelect={setDateRange} /> : <Button className={`mr-2`} size={'large'} type={'default'} onClick={() => setCollapse(false)}>
                    <SVG name={`arrowUp`} size={20} extraClass={`mr-1`}></SVG>Collapse all</Button>}
                <Button className={`ml-2`} size={'large'} type='primary' onClick={handleRunQuery}>Run Analysis</Button>
            </div>
        );
    };

    try {
        return (
            <div className={`${styles.composer}`}>
                <ComposerBlock
                    blockTitle={'CONVERSION GOAL'}
                    isOpen={convGblockOpen}
                    showIcon={true}
                    onClick={() => setConvGblockOpen(!convGblockOpen)}
                    extraClass={`no-padding-l no-padding-r`}
                    >
                    {renderConversionBlock()}
                </ComposerBlock>

                {eventGoal?.label?.length &&
                    <ComposerBlock
                        blockTitle={'MARKETING TOUCHPOINTS'}
                        isOpen={tchPointblockOpen}
                        showIcon={true}
                        onClick={() => setTchPointblockOpen(!tchPointblockOpen)}
                        extraClass={`no-padding-l no-padding-r`}
                        >
                        {renderMarkTouchpointBlock()}
                    </ComposerBlock>
                }

                {eventGoal?.label?.length &&
                    <ComposerBlock
                        blockTitle={'CRITERIA'}
                        isOpen={criteriablockOpen}
                        showIcon={true}
                        onClick={() => setCriteriablockOpen(!criteriablockOpen)}
                        extraClass={`no-padding-l no-padding-r`}
                        >
                        {renderAttributionOptions()}
                    </ComposerBlock>
                }

                {eventGoal?.label?.length &&
                    <ComposerBlock
                        blockTitle={'LINKED EVENTS'}
                        isOpen={linkEvExpansion}
                        showIcon={true}
                        onClick={() => toggleLinkEvExpansion()}
                        extraClass={`no-padding-l no-padding-r`}
                        >
                        {linkEvExpansion && models.length <= 1 && renderLinkedEvents()}
                    </ComposerBlock>
                }

                {eventGoal?.label?.length && footer()}
            </div>
        )
    } catch (err) { console.log(err) };
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    userProperties: state.coreQuery.userProperties,
    eventProperties: state.coreQuery.eventProperties,
    eventGoal: state.coreQuery.eventGoal,
    touchPoint: state.coreQuery.touchpoint,
    models: state.coreQuery.models,
    window: state.coreQuery.window,
    linkedEvents: state.coreQuery.linkedEvents,
    dateRange: state.coreQuery.attr_dateRange
});

const mapDispatchToProps = dispatch => bindActionCreators({
    fetchEventNames,
    getEventProperties,
    getUserProperties,
    getCampaignConfigData,
    setGoalEvent,
    setTouchPoint,
    setAttrDateRange,
    setModels,
    setWindow,
    setLinkedEvents
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(AttrQueryComposer);