import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import styles from './index.module.scss';
import { SVG, Text } from '../../components/factorsComponents';
import ConversionGoalBlock from './ConversionGoalBlock';

import FaDatepicker from '../../components/FaDatepicker';

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
import { Button, Popover } from 'antd';
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
    setAttrDateRange, dateRange
}) => {

    const [linkEvExpansion, setLinkEvExpansion] = useState(false);

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
        if(models.length > 1) return null;
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
            <div className={`${styles.composer__footer} fa--query_block`}>
                <FaDatepicker customPicker presetRange
                    monthPicker quarterPicker
                    range={
                        {
                            startDate: dateRange.from,
                            endDate: dateRange.to
                        }
                    }
                    placement="topRight" onSelect={setDateRange} />

                <Button size={'large'} type="primary" onClick={handleRunQuery}>Analyse</Button>
            </div>
        );

    };

    try {
        return (
            <div className={`${styles.composer}`}>
                <div className={`${styles.composer__section} fa--query_block`}>
                    <div className={styles.composer__section__title}>
                        <Text type={'title'} level={7} weight={'bold'}>CONVERSION GOAL</Text>
                    </div>
                    <div className={styles.composer__section__content}>
                        {renderConversionBlock()}
                    </div>
                </div>

                { eventGoal?.label?.length &&
                    <div className={`${styles.composer__section} fa--query_block`}>
                        <div className={styles.composer__section__title}>
                            <Text type={'title'} level={7} weight={'bold'}>MARKETING TOUCHPOINTS</Text>
                        </div>
                        <div className={styles.composer__section__content}>
                            {renderMarkTouchpointBlock()}
                        </div>
                    </div>
                }

                { eventGoal?.label?.length &&
                    <div className={`${styles.composer__section} fa--query_block`}>
                        <div className={styles.composer__section__title}>
                            <Text type={'title'} level={7} weight={'bold'}>CRITERIA</Text>
                        </div>
                        <div className={styles.composer__section__content}>
                            {renderAttributionOptions()}
                        </div>
                    </div>
                }

                { eventGoal?.label?.length &&
                    <div className={`${styles.composer__section} fa--query_block`}>
                        <div className={styles.composer__section__title}>
                            <Text type={'title'} level={7} weight={'bold'}>LINKED EVENTS</Text>
                            <Button type={'text'} onClick={toggleLinkEvExpansion}>
                                <SVG name={linkEvExpansion && models.length <=1 ? 'minus' : 'plus'} color={'black'}></SVG>
                            </Button>
                        </div>
                        <div className={styles.composer__section__content}>
                            {linkEvExpansion && models.length <= 1 && renderLinkedEvents()}
                        </div>
                    </div>
                }

                { eventGoal?.label?.length && footer()}
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