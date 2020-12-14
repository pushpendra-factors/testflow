import React, {useState, useEffect} from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import styles from './index.module.scss';
import { SVG, Text } from '../../components/factorsComponents';
import ConversionGoalBlock from './ConversionGoalBlock';

import { fetchEventNames, getUserProperties, getEventProperties } from '../../reducers/coreQuery/middleware';
import { Button } from 'antd';

const AttrQueryComposer = ({activeProject, fetchEventNames, userProperties, runAttributionQuery}) => {

    const [eventGoal, setEventGoal] = useState({
        label: 'Order Confirmed'
    });

    useEffect(() => {
        if (activeProject && activeProject.id) {
          fetchEventNames(activeProject.id);
          if(!userProperties.length) {
            getUserProperties(activeProject.id, 'analysis');
          }
        }
    }, [activeProject]);

    
    const renderConversionBlock = () => {
        if(eventGoal) {
            return (<ConversionGoalBlock eventGoal={eventGoal}></ConversionGoalBlock>)
        } else {
            return (<ConversionGoalBlock></ConversionGoalBlock>)
        }
        
    }


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

            <div className={`${styles.composer__section} fa--query_block`}>
                <div className={styles.composer__section__title}>
                    <Text type={'title'} level={7} weight={'bold'}>MARKETING TOUCHPOINTS</Text>
                </div>
                <div className={styles.composer__section__content}>
                    <ConversionGoalBlock></ConversionGoalBlock>
                </div>
            </div>

            <div className={`${styles.composer__section} fa--query_block`}>
                <div className={styles.composer__section__title}>
                    <Text type={'title'} level={7} weight={'bold'}>OTHER OPTIONS</Text>
                </div>
                <div className={styles.composer__section__content}>
                    <ConversionGoalBlock></ConversionGoalBlock>
                </div>
            </div>

            <div className={`${styles.composer__section} fa--query_block`}>
                <div className={styles.composer__section__title}>
                    <Text type={'title'} level={7} weight={'bold'}>LINKED EVENTS</Text>
                </div>
                <div className={styles.composer__section__content}>
                    <ConversionGoalBlock></ConversionGoalBlock>
                </div>
                <Button onClick={runAttributionQuery.bind(this, false)}>Run Query</Button>
            </div>
        </div>
    )
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    eventProperties: state.coreQuery.eventProperties,
    userProperties: state.coreQuery.userProperties
  });
  
  const mapDispatchToProps = dispatch => bindActionCreators({
    fetchEventNames,
    getEventProperties,
    getUserProperties
  }, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(AttrQueryComposer);