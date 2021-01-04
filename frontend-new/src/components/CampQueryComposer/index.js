import React, {useState, useEffect} from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import styles from './index.module.scss';
import { SVG, Text } from '../../components/factorsComponents';
import { Button, Popover } from 'antd';

import ChannelBlock from './ChannelBlock';

import {getCampaignConfigData, setCampChannel,
    setCampMeasures

} from 'Reducers/coreQuery/middleware';
import {fetchCampaignConfig} from 'Reducers/coreQuery/services';
import MeasuresBlock from './MeasuresBlock';

const CampQueryComposer = ({activeProject, channel, 
    getCampaignConfigData, 
    setCampChannel, measures,
    setCampMeasures, campaign_config,
    handleRunQuery
}) => {

    useEffect(()=>{
        if(activeProject && activeProject.id && channel) {
            getCampaignConfigData(activeProject.id, channel);
        }
    }, [activeProject, channel])


    const setChannel = (chan) => {
        setCampChannel(chan);
    }

    const renderChannelBlock = () => {
        if(channel) {
            return <ChannelBlock channel={channel} onChannelSelect={setChannel}></ChannelBlock>
        } else {
            return <ChannelBlock onChannelSelect={setChannel}></ChannelBlock>
        }

    }

    const setMeasuresToState = (msrs) => {
        setCampMeasures(msrs);
    }

    const renderMeasuresBlock = () => {
        if(measures) {
            return <MeasuresBlock measures={measures} 
            measures_metrics={campaign_config.metrics}
            onMeasureSelect={setMeasuresToState}></MeasuresBlock>
        }
        
    }

    const footer = () => {
          return (
            <div className={`${styles.composer__footer} fa--query_block`}>
              <Popover
                className="fa-event-popover"
                trigger="click"
                visible={false}
              >
                <Button size={'large'}><SVG name={'calendar'} extraClass={'mr-1'} /> This Month </Button>
              </Popover>
              <Button size={'large'} type="primary" onClick={handleRunQuery}>Run Query</Button>
            </div>
          );
        
      };

    return (
        <div className={styles.composer}>
            <div className={`${styles.composer__section} fa--query_block`}>
                <div className={styles.composer__section__title}>
                    <Text type={'title'} level={7} weight={'bold'}>Select Channel</Text>
                </div>
                <div className={styles.composer__section__content}>
                    {renderChannelBlock()}
                </div>
            </div>

            <div className={`${styles.composer__section} fa--query_block`}>
                <div className={styles.composer__section__title}>
                    <Text type={'title'} level={7} weight={'bold'}>MEASURES</Text>
                </div>
                <div className={styles.composer__section__content}>
                    {renderMeasuresBlock()}
                </div>
            </div>


            { channel && measures && measures.length && footer()}
        </div>
    );
}


const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    campaign_config: state.coreQuery.campaign_config,
    measures: state.coreQuery.camp_measures,
    channel: state.coreQuery.camp_channels
});
  
const mapDispatchToProps = dispatch => bindActionCreators({
    setCampChannel,
    setCampMeasures,
    getCampaignConfigData
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(CampQueryComposer);