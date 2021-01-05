import React, {useState, useEffect} from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import styles from './index.module.scss';
import { SVG, Text } from '../../components/factorsComponents';
import { Button, Popover } from 'antd';

import ChannelBlock from './ChannelBlock';

import {getCampaignConfigData, setCampChannel,
    setCampMeasures, setCampFilters

} from 'Reducers/coreQuery/middleware';
import {fetchCampaignConfig} from 'Reducers/coreQuery/services';
import MeasuresBlock from './MeasuresBlock';
import FilterBlock from '../QueryComposer/FilterBlock';

const CampQueryComposer = ({activeProject, channel, 
    getCampaignConfigData, 
    setCampChannel, measures,
    setCampMeasures, campaign_config,
    handleRunQuery, filters,
    setCampFilters
}) => {

    const [filterProps, setFilterProperties] = useState({});

    const [filterDD, setFilterDD] = useState(false);

    useEffect(() => {
        if(campaign_config.properties) {
            const props = {};
            campaign_config.properties.forEach((prop, i) => {
                props[prop.label] = prop.values;
            })
            setFilterProperties(props);
        }
    }, [campaign_config])

    useEffect(()=>{
        if(activeProject && activeProject.id && channel) {
            getCampaignConfigData(activeProject.id, channel);
            setMeasuresToState([]);
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

    const addFilter = (val) => {
        const fltrs = [...filters];
        const filt = fltrs.filter(fil => JSON.stringify(fil) === JSON.stringify(val));
        if (filt && filt.length) return;
        fltrs.push(val);
        setCampFilters(fltrs);
        closeFilter();
    };

    const closeFilter = () => {
        setFilterDD(false);
    };

    const delFilter = (index) => {
        const fltrs = filters.filter((v, i) => i !== index);
        setCampFilters(fltrs);
    }

    const renderFilterBlock = () => {
        if(filterProps) {
            const filtrs = [];

            filters.forEach((filt, id) => {
                filtrs.push(
                    <div key={id} className={id !== 0? `mt-4` : null}>
                        <FilterBlock activeProject={activeProject} 
                            index={id}
                            blockType={'global'} filterType={'channel'} 
                            filter={filt}
                            extraClass={styles.filterSelect}
                            delBtnClass={styles.filterDelBtn}
                            delIcon={`trash`}
                            deleteFilter={delFilter}
                            typeProps={{channel: channel}} filterProps={filterProps}
                            propsConstants={Object.keys(filterProps)}
                        ></FilterBlock>
                    </div>
                )
            })

            if(filterDD) {
                filtrs.push(  
                    <div key={filtrs.length} className={`mt-4`}>
                        <FilterBlock activeProject={activeProject} 
                            blockType={'global'} filterType={'channel'} 
                            extraClass={styles.filterSelect}
                            delBtnClass={styles.filterDelBtn}
                            typeProps={{channel: channel}} filterProps={filterProps}
                            propsConstants={Object.keys(filterProps)}
                            insertFilter={addFilter}
                            closeFilter={closeFilter}
                        ></FilterBlock>
                    </div>
                )
            } else {
                filtrs.push(
                    <div key={filtrs.length} className={`flex mt-4`}>
                        <div className={'fa--query_block--add-event flex justify-center items-center mr-2'}>
                            <SVG name={'plus'} color={'purple'}></SVG>
                        </div>

                        <Button size={'large'} type="link" onClick={() => setFilterDD(true)}>Add new</Button>
                    </div>
                )
            }

            
            
            return (<div className={styles.block}>{filtrs}</div>);
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

            {   channel && measures && measures.length? 
                    <div className={`${styles.composer__section} fa--query_block`}>
                        <div className={styles.composer__section__title}>
                            <Text type={'title'} level={7} weight={'bold'}>Filter</Text>
                        </div>
                        <div className={styles.composer__section__content}>
                            {renderFilterBlock()}
                        </div>
                    </div>
                :
                null
            }


            { channel && measures && measures.length? footer() : null}
        </div>
    );
}


const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    campaign_config: state.coreQuery.campaign_config,
    measures: state.coreQuery.camp_measures,
    filters: state.coreQuery.camp_filters,
    channel: state.coreQuery.camp_channels
});
  
const mapDispatchToProps = dispatch => bindActionCreators({
    setCampChannel,
    setCampMeasures,
    setCampFilters,
    getCampaignConfigData
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(CampQueryComposer);