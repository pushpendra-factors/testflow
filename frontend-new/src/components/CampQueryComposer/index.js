import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import styles from './index.module.scss';
import { SVG, Text } from '../../components/factorsComponents';
import { Button, Popover } from 'antd';

import ChannelBlock from './ChannelBlock';

import GroupSelect from "../QueryComposer/GroupSelect";

import {
    getCampaignConfigData, setCampChannel,
    setCampMeasures, setCampFilters, setCampGroupBy,
    setCampDateRange
} from '../../reducers/coreQuery/middleware';
import MeasuresBlock from './MeasuresBlock';
import FilterBlock from '../QueryComposer/FilterBlock';

import FaDatepicker from '../../components/FaDatepicker';

const CampQueryComposer = ({ activeProject, channel,
    getCampaignConfigData,
    setCampChannel, measures,
    setCampMeasures, campaign_config,
    filters, setCampFilters,
    groupBy, setCampGroupBy,
    handleRunQuery, dateRange,
    setCampDateRange
}) => {

    const [filterProps, setFilterProperties] = useState({});
    const [groupByProps, setGroupByProps] = useState([]);
    const [filterDD, setFilterDD] = useState(false);
    const [groupByDD, setGroupByDD] = useState([false]);
    const [dateRangePopover, setDateRangePopover] = useState(false);

    useEffect(() => {
        if (campaign_config.properties) {
            const props = {};
            const groupProps = [];
            campaign_config.properties.forEach((prop, i) => {
                props[prop.label] = prop.values;
                groupProps.push({
                    label: prop.label,
                    icon: prop.icon,
                    values: prop.values
                });
            });
            setFilterProperties(props);
            setGroupByProps(groupProps);
        }
    }, [campaign_config])

    useEffect(() => {
        if (activeProject && activeProject.id && channel) {
            getCampaignConfigData(activeProject.id, channel);
            // setMeasuresToState([]);
        }
    }, [activeProject, channel])


    const setChannel = (chan) => {
        setCampChannel(chan);
    }

    const renderChannelBlock = () => {
        if (channel) {
            return <ChannelBlock channel={channel} onChannelSelect={setChannel}></ChannelBlock>
        } else {
            return <ChannelBlock onChannelSelect={setChannel}></ChannelBlock>
        }

    }

    const setMeasuresToState = (msrs) => {
        setCampMeasures(msrs);
    }

    const renderMeasuresBlock = () => {
        if (measures) {
            return <MeasuresBlock measures={measures}
                measures_metrics={campaign_config.metrics}
                onMeasureSelect={setMeasuresToState}></MeasuresBlock>
        }

    }

    const addFilter = (val) => {
        const fltrs = [...filters];
        const value = Object.assign({}, val);
        val.props[2] = val.props[2].replace(' ', '_');
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
        if (filterProps) {
            const filtrs = [];

            filters.forEach((filt, id) => {
                filtrs.push(
                    <div key={id} className={id !== 0 ? `mt-4` : null}>
                        <FilterBlock activeProject={activeProject}
                            index={id}
                            blockType={'global'} filterType={'channel'}
                            filter={filt}
                            extraClass={styles.filterSelect}
                            delBtnClass={styles.filterDelBtn}
                            delIcon={`trash`}
                            deleteFilter={delFilter}
                            typeProps={{ channel: channel }} filterProps={filterProps}
                            propsConstants={Object.keys(filterProps)}
                        ></FilterBlock>
                    </div>
                )
            })

            if (filterDD) {
                filtrs.push(
                    <div key={filtrs.length} className={`mt-4`}>
                        <FilterBlock activeProject={activeProject}
                            blockType={'global'} filterType={'channel'}
                            extraClass={styles.filterSelect}
                            delBtnClass={styles.filterDelBtn}
                            typeProps={{ channel: channel }} filterProps={filterProps}
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

    const triggerGroupDD = (index) => {
        const grpDD = [...groupByDD];
        grpDD[index] = !grpDD[index];
        setGroupByDD(grpDD);
    }

    const renderGroupByBlock = () => {
        const groupByComponents = [];
        groupBy.forEach((gbp, index) => {
            groupByComponents.push(renderGroupBy(index));
        });
        groupByComponents.push(renderGroupBy(groupByComponents.length, true));
        return (<div className={styles.block}>
            {groupByComponents}
        </div>)
    }

    const onGroupBySet = (gbp, index) => {
        const newGroupByState = [...groupBy];
        const gbpState = {};
        gbpState.prop_category = gbp[0].replace(' ', '_');
        gbpState.property = gbp[1][0];
        gbpState.prop_type = gbp[1][1];
        if (newGroupByState[index]) {
            newGroupByState[index] = gbpState;
        } else {
            newGroupByState.push(gbpState);
        }
        setCampGroupBy(newGroupByState);
        triggerGroupDD(index);
    }

    const delGbpOption = (index) => {
        const newGroupByState = [...groupBy.filter((gb, i) => i !== index)];
        setCampGroupBy(newGroupByState);
    }

    const renderGroupBy = (index, init = false) => {
        return (<div key={0} className={` ${styles.groupItem} flex justify-start items-center mt-4`} >
            {!groupByDD[index] &&
                <>

                    {init === false ? <Button size={'large'}
                        type="text"
                        onClick={() => delGbpOption(index)}
                        className={`${styles.gbpRemove}`}>
                        <SVG name="trash"></SVG></Button> : null}

                    {init === true &&
                        <div className={'fa--query_block--add-event flex justify-center items-center mr-2'}>
                            <SVG name={'plus'} color={'purple'}></SVG>
                        </div>
                    }

                    <Button size={'large'} type="link" onClick={() => triggerGroupDD(index)}>
                        {init === true ?
                            <>Add new </> :
                            <><SVG name={groupBy[index].prop_category}></SVG>
                                <span className={`ml-2`}>
                                    {groupBy[index]?.property}
                                </span></>
                        }
                    </Button>

                </>

            }
            {groupByDD[index]
                ? (<GroupSelect groupedProperties={groupByProps}
                    placeholder="Select Property"
                    optionClick={(group, val) => onGroupBySet([group, val], index)}
                    onClickOutside={() => triggerGroupDD(index)}
                >
                </GroupSelect>
                )

                : null
            }
        </div>);
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

        setCampDateRange(dtRange);
    }

    const runCampaignsQuery = useCallback(() => {
      handleRunQuery(false, null);
    }, [handleRunQuery]);

    const footer = () => {
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
                <Button size={'large'} type="primary" onClick={runCampaignsQuery}>Run Query</Button>
            </div>
        );

    };

    try {

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

                {   channel && measures && measures.length ?
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


                {   channel && measures && measures.length ?
                    <div className={`${styles.composer__section} fa--query_block`}>
                        <div className={styles.composer__section__title}>
                            <Text type={'title'} level={7} weight={'bold'}>Group By</Text>
                        </div>
                        <div className={styles.composer__section__content}>
                            {renderGroupByBlock()}
                        </div>
                    </div>
                    :
                    null
                }

                { channel && measures && measures.length ? footer() : null}
            </div>
        );
    } catch (err) { console.log(err) }
}


const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    campaign_config: state.coreQuery.campaign_config,
    measures: state.coreQuery.camp_measures,
    filters: state.coreQuery.camp_filters,
    channel: state.coreQuery.camp_channels,
    groupBy: state.coreQuery.camp_groupBy,
    dateRange: state.coreQuery.camp_dateRange
});

const mapDispatchToProps = dispatch => bindActionCreators({
    setCampChannel,
    setCampMeasures,
    setCampFilters,
    getCampaignConfigData,
    setCampGroupBy,
    setCampDateRange
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(CampQueryComposer);