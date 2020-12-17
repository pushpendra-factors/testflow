import React, {useState, useEffect} from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import styles from './index.module.scss';
import { SVG, Text } from '../../components/factorsComponents';
import ConversionGoalBlock from './ConversionGoalBlock';

import { fetchEventNames, getUserProperties, getEventProperties } from '../../reducers/coreQuery/middleware';
import { Button, Popover } from 'antd';
import MarkTouchpointBlock from './MarkTouchpointBlock';
import AttributionOptions from './AttributionOptions';
import LinkedEventsBlock from './LinkedEventsBlock';
import { QUERY_TYPE_EVENT } from '../../utils/constants';

const AttrQueryComposer = ({activeProject, 
        fetchEventNames, getEventProperties, 
        userProperties, eventProperties, runAttributionQuery}) => {

    // <--- Things that go into reducer state; // Jitesh
    const [eventGoal, setEventGoal] = useState({});
    const [touchPoint, setTouchPoint] = useState('');

    const [models, setModels] = useState([]);
    const [window, setWindow] = useState();

    const [linkedEvents, setLinkedEvents] = useState([]);

    // ---> 

    const [linkEvExpansion, setLinkEvExpansion] = useState(false);

    useEffect(() => {
        if (activeProject && activeProject.id) {
          fetchEventNames(activeProject.id);
          if(!userProperties.length) {
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
        setEventGoal(eventGoal);
    }

    const linkEventChange = (linkEvent, index) => {
        const currLinkedEvs = [...linkedEvents];
        if(index === undefined || index < 0) {
            currLinkedEvs.push(linkEvent);
        } else {
            currLinkedEvs[index] = linkEvent;
        }
        setLinkedEvents(currLinkedEvs);
    }

    const goalDel = () => {
        setEventGoal({});
    }

    const linkEventDel = (index) => {
        const currLinkedEvs = linkedEvents.filter((ev, i) => i !== index);
        setLinkedEvents(currLinkedEvs);
    }

    
    const renderConversionBlock = () => {
        if(eventGoal) {
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
        if(linkedEvents && linkedEvents.length) {
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
        setLinkEvExpansion(!linkEvExpansion);
    }

    const handleRunQuery = () => {
        runAttributionQuery(false)
    };

    const footer = () => {
        if (!eventGoal || !eventGoal?.label?.length) { return null; }
        
          return (
            <div className={`${styles.composer__footer} fa--query_block`}>
              <Popover
                className="fa-event-popover"
                trigger="click"
                visible={false}
              >
                <Button size={'large'}><SVG name={'calendar'} extraClass={'mr-1'} /> This Month </Button>
              </Popover>
              <Button size={'large'} type="primary" onClick={handleRunQuery}>Analyse</Button>
            </div>
          );
        
      };

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
                        <Text type={'title'} level={7} weight={'bold'}>OTHER OPTIONS</Text>
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
                            <SVG name={linkEvExpansion? 'minus' : 'plus'} color={'black'}></SVG>
                        </Button>
                    </div>
                    <div className={styles.composer__section__content}>
                        {linkEvExpansion && renderLinkedEvents()}
                    </div>
                </div>
            }

            { eventGoal?.label?.length && footer()}
        </div>
    )
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    userProperties: state.coreQuery.userProperties,
    eventProperties: state.coreQuery.eventProperties
});
  
const mapDispatchToProps = dispatch => bindActionCreators({
    fetchEventNames,
    getEventProperties,
    getUserProperties
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(AttrQueryComposer);