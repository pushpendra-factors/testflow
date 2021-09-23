import React, { useEffect, useState } from 'react';
import { Tabs, Row, Col, Tooltip, Select, Button, Collapse, Tag, Spin, message } from 'antd'; 
import { SVG, Text, Number } from 'factorsComponents';
import { connect, useDispatch } from 'react-redux';
import _ from 'lodash';
import moment from 'moment';
import { fetchWeeklyIngishts, updateInsightFeedback } from 'Reducers/insights';

const { Option } = Select;

const { Panel } = Collapse;

const NoData = ({data}) => { 
    
    let text = {};
    text.title = 'No Insights available for this query!',
    text.subtitle = 'We are currently not supporting insights for this type of query.',
    text.info = false;

    if(data == 'not-yet-available'){
        text.title = 'Preparing Insights',
        text.subtitle = 'Insights would require a minimum of one week’s data. Please come back in a weeks time.',
        text.info = false

    }  
    if(data == 'add-to-dashboard'){ 
            text.title = 'No Insights available yet!',
            text.subtitle = 'This Analysis is not saved yet. To start seeing insights, follow the following steps:',
            text.info = true
    }  

    return (
        <div className={'flex flex-col items-center pt-10'}>
            <img src="assets/images/weekly-insights-no-data.png" className={'mb-2'} style={{ maxWidth: '175px', width: '100%' }} />
            <Text type={"title"} level={7} weight={'bold'} extraClass={"m-0"}>{text.title}</Text>
            <Text type={"title"} level={8} weight={'thin'} color={'grey'} extraClass={"m-0 mb-4"}>{text.subtitle}</Text>

            {text.info && <>
                <div className={'flex items-center mt-4'}>
                    <Button>Save Analysis</Button>
                    <Text type={"title"} level={8} weight={'thin'} color={'grey'} extraClass={"m-0 mx-2"}>and then,</Text>
                    <Button>Add to Dashboard</Button>
                </div>
            </>} 
        </div>
    )
}

const WeeklyInishgtsResults = ({data, activeInsight, requestQuery,activeProject , queryType, queryTitle, eventPropNames, userPropNames, fetchWeeklyIngishts, updateInsightFeedback }) => {    

    const [defaultActive, setDefaultActive] = useState(null);
    const [expandAll, setExpandAll] = useState(true);
    const [loading, setLoading] = useState(false);

    const TagIconSize = 14;
    const UpIcon = 'growthUp';
    const DownIcon = 'growthDown'; 

  const matchEventName = (item) => {   
      let findItem = eventPropNames?.[item] || userPropNames?.[item] 
    return findItem ? findItem : item
  }

    const panelActive = (panelNo) =>{ 
        setDefaultActive(_.map(panelNo, _.parseInt));
    }
     
    
    const togglePanels = () =>{
        
        if(expandAll){
            const panelCount = data?.actual_metrics?.length;
            const activePanel = _.range(panelCount)
            setDefaultActive(activePanel);
            setExpandAll(false);
        }
        else{
            setDefaultActive(null);
            setExpandAll(true);
        }
    }

    const UserRatingComp = ({item, index, actualData}) =>{

        const [isUpvote, seIsUpvote] = useState(false);
        const [isDownvote, seIsDownvote] = useState(false);

        const userRating = (e, item, index, rating) =>{
            e.stopPropagation(); 
            let data = {
                "feature":"weekly_insights",
                "property":{
                   "key":item?.key,
                   "date":"",
                   "order":index,
                   "value":item?.value,
                   "entity":item?.entity,
                   "query_id":actualData?.query_id,
                   "is_increased":item?.actual_values?.isIncrease
                },
                "vote_type":0
             }
    
            if(rating===1){
                data.vote_type = 1;
                seIsUpvote(true);
                seIsDownvote(false);
            }
            else if(rating===2){
                seIsUpvote(false);
                seIsDownvote(true);
                data.vote_type = 2
            }  
            updateInsightFeedback(activeProject?.id,data).then(()=>{
                // message.success('Successfully maked your feedback!');
            }).catch((err) => { 
                message.error('feedback submission failed!');
                console.log('feedback submission failed!',err);
              });
            // console.log('clicked rating',rating )
        }

        return(
            <div className={'flex items-center mx-4 insights-rating--block'}>
                <Text type={"title"} color={'grey'} level={8} extraClass={"m-0 mx-2"}>{`Was this useful?`}</Text>
                <Button onClick={(e)=>userRating(e, item, index, 1)}  size={'small'} icon={<SVG name={isUpvote? 'ThumbsUp_S': 'ThumbsUp'} color={isUpvote? 'blue': 'grey'} size={12} />} className={'ml-1'} />
                <Button onClick={(e)=>userRating(e, item, index, 2)} size={'small'} icon={<SVG name={isDownvote? 'ThumbsDown_S': 'ThumbsDown'} size={12} color={isDownvote? 'blue': 'ThumbsUp'} />} className={'ml-1'} /> 
        </div>
        )
    }

   

    const highlightCard = (data, title, margin = false, isPercent = false) => {
        return (<div className={`flex items-center mt-4 border--thin-2 py-4 px-8 border-radius--sm  w-full ${margin ? 'mx-4' : ''}`} style={{maxWidth: '400px'}}>
            <div className={'flex items-center'}>
                {data.isIncrease ? <SVG name={UpIcon} size={24} color={'green'} /> : <SVG name={DownIcon} size={24} color={'red'} />}
                <Text type={"title"} level={4} weight={'bold'} extraClass={"m-0 ml-2"}><Number suffix={'%'} number={data.percentage} /></Text>
            </div>
            <div className={'flex flex-col ml-4'}>
                <Text type={"title"} level={8} weight={'bold'} extraClass={"m-0 uppercase"}>{title}</Text>
                <Text type={"title"} level={8} color={'grey'} extraClass={"m-0"}>{`(`}<Number number={data.w1} suffix={isPercent? "%" : ''} /> {` -> `}<Number number={data.w2} suffix={isPercent? "%" : ''} />{`)`}</Text>
            </div>

        </div>)
    }

    const genHeader = (item,index=0, actualData=false) => {  
        const data = item.actual_values;  
        return (
            <div className={'flex justify-between items-center py-2 insights-rating--container'}>
                <div className={'flex  items-center'}>
                    <Tag color={data.isIncrease ? 'green' : "red"} className={`${data.isIncrease ? 'fa-tag--green' : "fa-tag--red"}`}>
                        {data.isIncrease ? <SVG name={UpIcon} size={TagIconSize} color={'green'} /> : <SVG name={DownIcon} size={TagIconSize} color={'red'} />}
                        <Number suffix={'%'} number={data.percentage} />
                    </Tag>
                    <Text type={"title"} level={7} extraClass={"m-0 mx-2"}>{`${data.isIncrease ? 'Increase' : 'Decrease'}  where`}</Text>
                    <Tag className={'m-0 mx-2'} className={'fa-tag--regular fa-tag--highlight'}>{`${matchEventName(item.key)}`}</Tag>
                    <Text type={"title"} level={7} extraClass={"m-0 ml-2"}>is</Text>
                    <Text type={"title"} level={7} weight={'bold'} extraClass={"m-0 ml-1 mr-2"}>{`${item.value}`}</Text>
                    <Text type={"title"} weight={'thin'} color={'grey'} level={8} extraClass={"m-0"}>{`(`}<Number number={data?.w1}   /> {` -> `}<Number number={data?.w2}   />{`)`}</Text>
                </div>
                <div className={'flex  items-center'}>
                 <UserRatingComp 
                 item={item}
                 index={index}
                 actualData={actualData}
                 />
                    <Tag className={'fa-tag--grey uppercase'}>{item.type}</Tag>
                </div>

            </div>
        )
    }

    const genBody = (item)=> {
        const prevalance = item.change_in_prevalance;
        const conversion = item.change_in_conversion;   
        if(item?.type== 'distribution' ){
            const data = item?.change_in_distribution;
            const value1 = item?.change_in_distribution?.w1;
            const value2 = item?.change_in_distribution?.w2;
            return( 
                    <div className={'flex  items-center pl-10'}> 
                        <Text type={"title"} weight={'thin'} color={'grey'} level={8} extraClass={"m-0 mr-2"}> {`Share of`}</Text> 
                        <Tag className={'m-0 mx-2'} className={'fa-tag--regular fa-tag--highlight'}>{`${matchEventName(item.key)}`}</Tag>
                        <Text type={"title"} weight={'thin'} color={'grey'} level={8} extraClass={"m-0 ml-1"}> {`is`}</Text> 
                        <Text type={"title"} weight={'bold'} color={'grey'} level={8} extraClass={"m-0 ml-1"}>{item.value}</Text>
                        <Text type={"title"} weight={'thin'} color={'grey'} level={8} extraClass={"m-0 mx-1"}>{`${data.isIncrease ? 'increased' : 'decreased'} by`}</Text>  
                        <Tag color={data.isIncrease ? 'green' : "red"} className={`m-0 mx-1 ${data.isIncrease ? 'fa-tag--green' : "fa-tag--red"}`}>
                            {data.isIncrease ? <SVG name={UpIcon} size={TagIconSize} color={'green'} /> : <SVG name={DownIcon} size={TagIconSize} color={'red'} />}
                            <Number suffix={'%'} number={data?.percentage} />
                        </Tag>
                        <Text type={"title"} weight={'thin'} color={'grey'} level={8} extraClass={"m-0 ml-1"}> {` from `} <Number number={value1} suffix={'%'}  />{` to `}<Number number={value2} suffix={'%'}  /> </Text> 
                    </div>  
            )
        }  
        return (
            <div className={'flex  items-center pl-10'}>

                <Tag className={'fa-tag--regular flex items-center'}>
                    {prevalance.isIncrease ? <SVG name={UpIcon} size={TagIconSize} color={'green'} /> : <SVG name={DownIcon} size={TagIconSize} color={'red'} />}
                    <Number suffix={'%'} number={prevalance.percentage} />
                </Tag>
                <Text type={"title"} level={8} color={'grey'} extraClass={"m-0 mx-2"}>Change in Prevailance</Text>
                <Text type={"title"} weight={'thin'} color={'grey'} level={8} extraClass={"m-0 mr-4"}>{`(`}<Number number={prevalance.w1} /> {` -> `}<Number number={prevalance.w2} />{`)`}</Text>

                <Tag className={'fa-tag--regular flex items-center ml-4'}>
                    {conversion.isIncrease ? <SVG name={UpIcon} size={TagIconSize} color={'green'} /> : <SVG name={DownIcon} size={TagIconSize} color={'red'} />}
                    <Number suffix={'%'} number={conversion.percentage} />
                </Tag>
                <Text type={"title"} level={8} color={'grey'} extraClass={"m-0 mx-2"}>Change in Conversion</Text>
                <Text type={"title"} weight={'thin'} color={'grey'} level={8} extraClass={"m-0 mr-4"}>{`(`}<Number number={conversion.w1} suffix={'%'}  /> {` -> `}<Number number={conversion.w2} suffix={'%'}  />{`)`}</Text>

            </div>
        )
    }


    const dateData = activeInsight?.InsightsRange;

    let dataObjArr = Object.keys(dateData).map((item,index)=>{
        return {
            text:`${moment.unix(item).format("MMM DD, YYYY")} - ${moment.unix(item).endOf('week').format("MMM DD, YYYY")}`,
            value: item
        }
    }); 
    
    
    const dataOptions = dataObjArr.map((item,index)=>{ 
        return  <Option value={item.text}>{item.text}</Option> 
    })
    
    const handleChangeWeek = (value) => {
        setLoading(true)
        let dataObjItem = dataObjArr?.find((item)=>{ 
            return item.text == value
        })
        let dataObjVal = dataObjItem?.value;
        fetchWeeklyIngishts(
            activeProject?.id,
            activeInsight?.id,
            dataObjVal,
            dateData?.[dataObjVal][0],
            activeInsight?.isDashboard
        ).then(()=>{
            setLoading(false);
        }).catch((e) => {
            setLoading(false);
            console.log('weekly-ingishts fetch error', e);
          }); 
    }
    let insightsLen =  Object.keys(dateData)?.length || 0; 
    const defaultDate = `${moment.unix(Object.keys(dateData)[insightsLen-1]).format("MMM DD, YYYY")} - ${moment.unix(Object.keys(dateData)[insightsLen-1]).endOf('week').format("MMM DD, YYYY")}`;
    // const WeekData = `${moment.unix(1624147200).format("MMM DD, YYYY")} - ${moment.unix(1624147200).endOf('week').format("MMM DD, YYYY")}`; 
    const baseName = requestQuery?.cl == "funnel" ? requestQuery?.ewp[0].na : "Sessions";

    
    
    return (
        <div className=''>
                <Row gutter={[24, 24]}>
                    <Col span={12}>
                        <div className={'flex items-center mt-6'}>
                            <Text type={"title"} level={7} color={'grey'} weight={''} extraClass={"m-0"}>Insights for</Text> 
                           
                            {/* <Text type={"title"} level={7} weight={'bold'} extraClass={"m-0 ml-2"}>{defaultDate}</Text>  */}
                           <div className={'ml-2'}> 
                                <Select loading={loading} disabled={loading} className={'fa-select'} defaultValue={defaultDate} style={{ width: 240 }} onChange={handleChangeWeek}>
                                        {dataOptions}
                                </Select>
                            </div>
                            
                            <Text type={"title"} level={7} color={'grey'} weight={''} extraClass={"m-0 ml-2"}> compared to</Text> 
                            <Text type={"title"} level={7} weight={'bold'} extraClass={"m-0 ml-2"}> Week Before</Text> 
                        </div>
                    </Col>
                    <Col span={12}>
                        <div className={'flex justify-end items-center mt-6'}>
                            <Button type={'text'} style={{minWidth: '170px'}} onClick={togglePanels}>{expandAll ?  <SVG size={16} name={'SortDown'} /> : <SVG size={16} name={'SortUp'} /> } {expandAll ? 'Expand Insights' : 'Collapse Insights' }</Button> 
                        </div>
                    </Col>
                    
                </Row>
            <div className={'fa-container mt-0'}>
                { loading ?  <div className='flex justify-center items-center mt-10'>
        <Spin  />
      </div> : <>
                <Row>
                    <Col span={24}> 
                        <div className={'flex items-baseline justify-between'}> 
                            <Text type={"title"} level={3} weight={'bold'} extraClass={"m-0 mt-2"}>{queryTitle}</Text> 

                            {data?.baseline &&
                            <div className={'flex items-baseline justify-end'}>
                                <Text type={"title"} level={7}  extraClass={"m-0"}>{`Baseline :`}</Text> 
                                <Text type={"title"} level={7} weight={'bold'} extraClass={"m-0 ml-1"}>{data?.baseline ? matchEventName(data?.baseline) : `Sessions`}</Text> 
                                <Tooltip placement="top" title={'The change in metric is compared against the change in baseline to identify relevant insights'} trigger="hover">
                                    <Button type={'text'} icon={<SVG name={'infoCircle'} size={16} />} className={'ml-1'} />
                                </Tooltip>
                            </div>
                            }
                        </div>
                    </Col> 
                    <Col span={24}>

                        <div className={'flex items-stretch'}>
                            
                        {data?.insights_type == 'ConvAndDist' ? <>
                            {highlightCard(data?.goal, 'Overall', false)}
                            {highlightCard(data?.base, baseName, true )}
                            {highlightCard(data?.conv, 'Conv. Rate', false, true)}
                            </> : <> {highlightCard(data?.goal, 'Overall')} </>}


                        </div>

                    </Col>
                </Row>
                <Row gutter={[24, 24]}>
                    <Col span={24}>
                        <div className={'mt-4'}>
                        <Collapse 
                            activeKey={defaultActive}
                            expandIconPosition={'right'}
                            className={`fa-insights--panel`}
                            onChange={panelActive}
                        >
                            {data?.actual_metrics && data?.actual_metrics?.map((item, index) => { 
                                return (
                                    <Panel 
                                     className={'fa-insights--panel-item'}
                                     header={genHeader(item,index, data)} key={index} 
                                     > 
                                        {genBody(item)}
                                    </Panel>
                                )
                            })}
                        </Collapse>
                        </div>
                    </Col>
                </Row>
                </>}
            </div>
        </div>
    )
}
const WeeklyInishgts = ({
    insights,
    requestQuery,
    queryType,
    queryTitle,
    fetchWeeklyIngishts,
    activeProject,
    eventPropNames,
    userPropNames,
    updateInsightFeedback
}) => { 
    const [insightsData, setInsightsData] = useState(null); 

    useEffect(()=>{ 
         if(insights){
            setInsightsData(insights)
        }  
    }, [insights]); 

    const renderData = (insightsData) =>{
        if(!insightsData?.active_insight){
            return  <NoData data={'add-to-dashboard'} />
        }
        if(!insightsData?.active_insight?.Enabled){
            return <NoData />
        }
        if((insightsData?.active_insight?.Enabled && _.isEmpty(insightsData?.weekly_insights))){
            return  <NoData data={'not-yet-available'} /> 
        }
        if((insightsData?.active_insight?.Enabled && !_.isEmpty(insightsData?.weekly_insights)))
        {
            return <WeeklyInishgtsResults 
            activeInsight={insightsData?.active_insight}
            data={insightsData?.weekly_insights} 
            requestQuery={requestQuery}
            queryType={queryType}
            queryTitle={queryTitle}
            fetchWeeklyIngishts={fetchWeeklyIngishts}
            activeProject={activeProject}
            eventPropNames={eventPropNames}
            userPropNames={userPropNames}
            updateInsightFeedback={updateInsightFeedback}
             />
        }

    }

    return (
        <>  
            {renderData(insightsData)}
        </>
    )
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    insights: state.insights,
    eventPropNames: state.coreQuery.eventPropNames,
    userPropNames: state.coreQuery.userPropNames

});


export default connect(mapStateToProps, {fetchWeeklyIngishts, updateInsightFeedback})(WeeklyInishgts)