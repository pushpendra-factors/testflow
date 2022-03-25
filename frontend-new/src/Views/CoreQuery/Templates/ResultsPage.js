import React, { useCallback, useEffect, useState, useContext } from 'react';
import { SVG, Text, FaErrorComp, FaErrorLog, Number } from 'Components/factorsComponents';
import { Input, Button, Tabs, Row, Col, Skeleton, Spin, message, Form, Tooltip, Popover } from 'antd';
import { useHistory } from 'react-router-dom';
import { ErrorBoundary } from 'react-error-boundary';
import { fetchTemplateConfig, fetchTemplateInsights } from 'Reducers/templates';
import { connect } from 'react-redux';
import FaDatepicker from 'Components/FaDatepicker'; 
import styles from './index.module.scss';
import MomentTz from 'Components/MomentTz'; 

const { TabPane } = Tabs;

function TemplateResults({
  fetchTemplateConfig,
  activeProject,
  templateConfig,
  fetchTemplateInsights
}) {

  const history = useHistory();
  const [selectedInsight, setSelectedInsight] = useState(null);
  const [subInsightData, setSubInsightData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [suffixSymbol, setSuffixSymbol] = useState('%');
  const [selectedTab, setSelectedTab] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [showSearch, setShowSearch] = useState(false);
  const [subSearchTerm, setSubSearchTerm] = useState('');
  const [showSearchSub, setShowSearchSub] = useState(false);
  const [sortInsight, setSortInsight] = useState(false);
  const [sortSubInsight, setSortSubInsight] = useState(false);

  const onInputSearch = (userInput) => {
    setSearchTerm(userInput.currentTarget.value);
  };
  const onSubInputSearch = (userInput) => {
    setSubSearchTerm(userInput.currentTarget.value);
  };

  const [dateRange1, setDateRange1] = useState({
    t1: MomentTz().subtract(2, 'weeks').startOf('week'),
    t2: MomentTz().subtract(2, 'weeks').endOf('week'),
  });
  const [dateRange2, setDateRange2] = useState({
    t1: MomentTz().subtract(1, 'weeks').startOf('week'),
    t2: MomentTz().subtract(1, 'weeks').endOf('week'),
  });

  const addShadowToHeader = useCallback(() => {
    const scrollTop =
      window.pageYOffset !== undefined
        ? window.pageYOffset
        : (
          document.documentElement ||
          document.body.parentNode ||
          document.body
        ).scrollTop;
    if (scrollTop > 0) {
      document.getElementById('app-header').style.filter =
        'drop-shadow(0px 2px 0px rgba(200, 200, 200, 0.25))';
    } else {
      document.getElementById('app-header').style.filter = 'none';
    }
  }, []);

  const routeChange = (url) => {
    history.push(url);
  };

  useEffect(() => {
    document.addEventListener('scroll', addShadowToHeader);
    return () => {
      document.removeEventListener('scroll', addShadowToHeader);
    };
  }, [addShadowToHeader]);


  const roundNumb = (num) => Math.round(num * 100) / 100;


  const configMatrix = templateConfig?.config?.metrics;
  const queryResult = templateConfig?.insight?.result;
  const metaTitle = templateConfig?.insight?.result?.meta;
  const column1 = metaTitle?.primary_level?.column_name;
  const column2 = metaTitle?.sub_level?.column_name;


  const showSubInsight = (data, index) => {
    setSubInsightData(data)
    setSelectedInsight(index);
  }

  const InsightItem = (data) => {
    const insights = data?.queryResult?.breakdown_analysis?.primary_level_data;
    if (insights) {
      insights.sort(function (a, b) {
        return sortInsight ? (a.last_value - a.previous_value) - (b.last_value - b.previous_value) : (b.last_value - b.previous_value) - (a.last_value - a.previous_value);
      });
      return (
        insights?.map((item, j) => {
          if (item?.name.toLowerCase().includes(searchTerm.toLowerCase())) {
            let isIncreased = item?.percentage_change >= 0;
            return (
              <div className={`my-1 py-2 px-4 mx-2 flex items-center justify-between cursor-pointer  border-radius--sm border--thin-2--transparent ${selectedInsight == j ? 'border--thin-2--brand' : ''}`} onClick={() => showSubInsight(item.sub_level_data, j)}>
                <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mr-3 capitalize'}>{item?.name}</Text>
                <div className={'flex items-end flex-col'}>
                  <div className={'flex items-center'}>
                    {item.is_infinity ? <Text type={'title'} level={6} extraClass={'m-0'}>∞</Text> : <>
                      <SVG name={isIncreased ? 'spikeup' : 'spikedown'} color={isIncreased ? 'green' : 'red'} size={18} />
                      <Text type={'title'} level={7} weight={'bold'} color={isIncreased ? 'green' : 'red'} extraClass={'m-0 ml-1'}><Number number={item?.percentage_change} suffix={'%'} /></Text>
                    </>
                    }
                  </div>
                  <Text type={'title'} level={8} color={'grey'} extraClass={'m-0 ml-2'}>{`(`}<Number suffix={suffixSymbol} number={roundNumb(item?.previous_value)} shortHand={true} />{` -> `}<Number suffix={suffixSymbol} number={roundNumb(item?.last_value)} shortHand={true} />{`)`}</Text>
                </div>
              </div>
            )
          }
        })
      )
    }
    else return <NoData />
  }


  const metricDisplayName = (item) => {
    let findItem = configMatrix.find((element) => { if (element.metric == item) return element.display_name })
    return findItem ? findItem.display_name : item
  }
  
  const SubInsightItem = () => {
    if (subInsightData) {
      subInsightData.sort(function (a, b) {
        return sortSubInsight ? roundNumb(a.last_value - a.previous_value) - roundNumb(b.last_value - b.previous_value) : roundNumb(b.last_value - b.previous_value) - roundNumb(a.last_value - a.previous_value);
      });
      return (
        subInsightData?.map((item, j) => {
          if (item?.name.toLowerCase().includes(subSearchTerm.toLowerCase())) {
            let isIncreased = item?.percentage_change >= 0;
            return (
              <div className={`py-3 px-6  flex items-center justify-between`}>
                <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mr-3 capitalize'}>{item?.name}</Text>
                <div className={'flex items-center'}>

                  {item.is_infinity ? <Text type={'title'} level={6} extraClass={'m-0'}>∞</Text> : <>
                    <SVG name={isIncreased ? 'spikeup' : 'spikedown'} color={isIncreased ? 'green' : 'red'} size={18} />
                    <Text type={'title'} level={7} weight={'bold'} color={isIncreased ? 'green' : 'red'} extraClass={'m-0 ml-1'}><Number number={item?.percentage_change} suffix={'%'} /></Text>
                  </>}

                  <Text type={'title'} level={8} color={'grey'} extraClass={'m-0 ml-2'}>{`(`}<Number suffix={suffixSymbol} number={roundNumb(item?.previous_value)} shortHand={true} />{` -> `}<Number suffix={suffixSymbol} number={roundNumb(item?.last_value)} shortHand={true} />{`)`}</Text>

                  {item?.root_cause_metrics && <div className={'flex items-center '}>
                    {/* <Text type={'title'} level={8} color={'grey'} extraClass={'m-0 ml-2'}>Due to</Text> */}
                    <Popover placement="top" content={
                      item?.root_cause_metrics?.map((subitem) => {
                        let isIncreased = subitem?.percentage_change >= 0;
                        return (
                          <div className={'flex items-center'}>
                            <Text type={'title'} level={8} color={'grey'} extraClass={'m-0'}>
                              {metricDisplayName(subitem?.metric)}
                            </Text>
                            <Text type={'title'} level={8} color={'grey'} extraClass={'m-0 mx-1'}>
                              {`${isIncreased ? 'increased' : 'decreased'}`}
                            </Text>
                            {subitem.is_infinity ? <Text type={'title'} level={6} extraClass={'m-0'}>∞</Text> :
                              <Number number={subitem?.percentage_change} suffix={'%'} />}
                          </div>
                        )
                      })
                    } trigger="hover">
                      <Button type={'text'} icon={<SVG name={'infoCircle'} size={16} />} className={'ml-1'} />
                    </Popover>
                  </div>
                  }
                </div>
              </div>
            )
          }
        })
      )
    }
    else return null
  }

  const CardInsights = (data) => {
    const cards = data?.queryResult?.breakdown_analysis?.overall_changes_data;
    if (cards) {
      return (
        <div className={'flex items-center w-full'}>
          {cards?.map((item, j) => {
            let isIncreased = item?.percentage_change >= 0;
            let symbolPecent = item.metric == 'search_impression_share' || item.metric == 'click_through_rate' || item.metric == 'conversion_rate';
            return (
              <div className={`py-4 px-6 border--thin-2 flex flex-col w-full  border-radius--sm ${j == 1 ? 'mx-4' : ''}`} style={{ maxWidth: '380px' }}>
                <div className={`flex  items-center`}>
                  <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 mr-3 capitalize'}>{metricDisplayName(item.metric)}</Text>
                  {item.is_infinity ? <Text type={'title'} level={5} extraClass={'m-0'}>∞</Text> : <>
                    <SVG name={isIncreased ? 'spikeup' : 'spikedown'} color={isIncreased ? 'green' : 'red'} size={20} />
                    <Text type={'title'} level={6} weight={'bold'} color={isIncreased ? 'green' : 'red'} d extraClass={'m-0 ml-1'}><Number number={item?.percentage_change} suffix={'%'} /></Text>
                  </>}
                </div>
                <Text type={'title'} level={8} color={'grey'} extraClass={'m-0'}>{`Overall ${isIncreased ? 'increased' : 'decreased'} from `}<Number suffix={symbolPecent ? "%" : ''} number={roundNumb(item?.previous_value)} shortHand={true} />{` to `}<Number suffix={symbolPecent ? "%" : ''} number={roundNumb(item?.last_value)} shortHand={true} /></Text>
              </div>
            )
          })}
        </div>
      )
    }
    else return null
  }

  const NoSubInsightsData = () => {
    return (<div className={'flex flex-col my-20 py-10 px-5 items-center'}>
      <img src={'https://s3.amazonaws.com/www.factors.ai/assets/img/product/add-widget-icon.png'} style={{ maxHeight: '80px' }} />
      <Text type={'title'} level={8} color={'grey'} weight={'thin'} extraClass={'m-0 mr-3'}>{`Select the ${column1} to see the ${column2} data`}</Text>
    </div>)
  }
  const NoData = () => {
    return (<div className={'flex flex-col my-20 py-10 px-5 items-center'}>
      <img src={'https://s3.amazonaws.com/www.factors.ai/assets/img/product/add-widget-icon.png'} style={{ maxHeight: '80px' }} />
      <Text type={'title'} level={8} color={'grey'} weight={'thin'} extraClass={'m-0 mr-3'}>{`No insights available`}</Text>
    </div>)
  }

  const fetchInsights = (key) => {
    setLoading(true);
    const queryData = {
      "metric": key, 
      'prev_from': dateRange1 ? MomentTz(dateRange1.t1).unix() : MomentTz().subtract(2, 'weeks').startOf('week').unix(),
      'prev_to': dateRange1 ? MomentTz(dateRange1.t2).unix() : MomentTz().subtract(2, 'weeks').endOf('week').unix(),
      "from": dateRange2 ? MomentTz(dateRange2.t1).unix() : MomentTz().subtract(1, 'weeks').startOf('week').unix(),
      "to": dateRange2 ? MomentTz(dateRange2.t2).unix() : MomentTz().subtract(1, 'weeks').endOf('week').unix(),
      thresholds: {
        percentage_change: 10,
        absolute_change: 0,
      },
      time_zone: localStorage.getItem('project_timeZone') || 'Asia/Kolkata'
    }
    fetchTemplateInsights(activeProject.id, queryData).then(() => {
      setLoading(false);
    }).catch((e) => {
      setLoading(false);
      console.log('fetchTemplateInsights error', e)
      message.error(`Sorry! couldn’t fetch insights for ${key}`)
    });
  }
  useEffect(() => {
    if (templateConfig && configMatrix) { 
      setSelectedTab(configMatrix[0].metric); 
    }
    else {
      routeChange('/analyse')
    }
  }, []);

  const onTabChange = (key) => {
    setSearchTerm('');
    setShowSearch(false);
    setSubSearchTerm('');
    setShowSearchSub(false);
    setSortInsight(false);
    setSortSubInsight(false);
    setSubInsightData(null);
    setSelectedInsight(null);
    setSelectedTab(key);
    // fetchInsights(key);
    if (key == 'search_impression_share' || key == 'click_through_rate' || key == 'conversion_rate') {
      setSuffixSymbol("%")
    }
    else {
      setSuffixSymbol("")
    }
  }


  const dateChange1 = (ranges) => {
    let timestamps = {
      t1: MomentTz(ranges.startDate),
      t2: MomentTz(ranges.endDate),
    }
    setDateRange1(timestamps);
  }

  const dateChange2 = (ranges) => {
    let timestamps = {
      t1: MomentTz(ranges.startDate),
      t2: MomentTz(ranges.endDate),
    }
    setDateRange2(timestamps);
  }

  useEffect(() => { 
    if (selectedTab) {
      fetchInsights(selectedTab);
    }
  }, [dateRange1, dateRange2, selectedTab]);

 
  return (<>
    <div
      id='app-header'
      className={`bg-white z-50	flex fixed items-center justify-between py-3 px-8 w-full top-0`}
    >
      <div
        className='flex items-center cursor-pointer'
      >
        <Button
          size={'large'}
          type='text'
          icon={<SVG size={32} name='Brand' />}
          className={'mr-2'}
          onClick={() => { routeChange('/') }}
        />
        <div>
          <Text
            type={'title'}
            level={7}
            extraClass={'m-0 mt-1'}
            color={'grey'}
            lineHeight={'small'}
            onClick={() => { routeChange('/analyse') }}
          >
            {`Templates / Google Search Ads Anomaly`}
          </Text>
        </div>
      </div>
      <div className='flex items-center'>
      </div>
    </div>

    <div className='mt-24 px-20'>
      <div className='fa-container'>
        <ErrorBoundary
          fallback={
            <FaErrorComp
              size={'medium'}
              title={'Analyse Results Error'}
              subtitle={
                'We are facing trouble loading Analyse results. Drop us a message on the in-app chat.'
              }
            />
          }
          onError={FaErrorLog}
        >

          <div className={'flex items-center'}>
            {/* <Tooltip placement="top" title={PrevWeekDateString}>
              <Button><SVG name={'calendar'} size={16} extraClass={'mr-1'} />Prev. Week</Button>
            </Tooltip>
            <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mx-2'}>vs</Text>
            <Tooltip placement="top" title={LastWeekDateString}>
              <Button><SVG name={'calendar'} size={16} extraClass={'mr-1'} />Last Week</Button>
            </Tooltip> */}


            <Tooltip placement="top" title={"Base Timeframe"}>  
            <div>
              <FaDatepicker
                customPicker
                presetRange
                monthPicker
                range={{
                  startDate: dateRange1 ? dateRange1.t1 : null,
                  endDate: dateRange1 ? dateRange1.t2 : null,
                }}
                onSelect={dateChange1}
              />
            </div>
            </Tooltip> 
            <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mx-2'}>vs</Text>
            <Tooltip placement="top" title={"Comparison Timeframe"}>  
            <div> 
              <FaDatepicker
                customPicker
                presetRange
                monthPicker
                range={{
                  startDate: dateRange2 ? dateRange2.t1 : null,
                  endDate: dateRange2 ? dateRange2.t2 : null,
                }}
                onSelect={dateChange2}
              />
            </div>
            </Tooltip>

          </div>

          {configMatrix ? <div className='mt-8'>
            <Tabs className={"fa-tabs--dashboard fa-tabs--white capitalize fa-tabs--no-padding"} defaultActiveKey="1" tabPosition={'top'} onChange={onTabChange} >
              {configMatrix?.map((item, j) => (
                <TabPane tab={`${item.display_name}`} key={item.metric} className={'capitalize'} />
              ))}
            </Tabs>
          </div> : <Skeleton loading={true} active />
          }



          {(!loading) ?
            <> 
              {queryResult && queryResult?.breakdown_analysis?.primary_level_data ? <>
                <div className={'my-6 w-full'}>
                  <CardInsights queryResult={queryResult} />
                </div>
                <div className='mt-6'>
                  <Row gutter={[24, 24]}>
                    <Col span={12}>
                      <div className={'pr-4'}>
                        <div className={'border--thin-2  border-radius--sm '}>
                          <div className={'py-4 px-6 background-color--brand-color-1 border-radius--sm flex justify-between'}>
                            <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0 capitalize'}>{column1}</Text>
                            <div className={'flex justify-between'}>
                              {showSearch ? <Input
                                className={styles.input}
                                onChange={onInputSearch}
                                prefix={(<SVG name="search" size={16} color={'grey'} />)}
                              /> : null}
                              <Button
                                type='text'
                                className={styles.btn}
                                onClick={() => { setShowSearch(!showSearch); if(showSearch){setSearchTerm('')} }}
                              >
                                <SVG name={!showSearch ? 'search' : 'close'} size={20} color={'grey'} />
                              </Button>
                              <Button
                                type='text'
                                className={styles.btn}
                                onClick={() => { setSortInsight(!sortInsight) }}
                              >
                                <SVG name={!sortInsight ? 'sortdown' : 'sortup'} size={20} color={'grey'} />
                              </Button>
                            </div>
                          </div>
                          <div className={'fa-vertical-scrolling-card px-4'}>
                            <InsightItem queryResult={queryResult} />
                          </div>
                        </div>
                      </div>
                    </Col>

                    <Col span={12}>
                      <div className={'pl-4'}>
                        <div className={'border--thin-2  border-radius--sm '}>
                          <div className={'py-4 px-6 background-color--brand-color-1 border-radius--sm flex justify-between'}>
                            <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0 capitalize'}>{metaTitle?.sub_level?.column_name}</Text>
                            <div className={'flex justify-between'}>
                              {showSearchSub ? <Input
                                className={styles.input}
                                onChange={onSubInputSearch}
                                prefix={(<SVG name="search" size={16} color={'grey'} />)}
                              /> : null}
                              <Button
                                type='text'
                                className={styles.btn}
                                onClick={() => { setShowSearchSub(!showSearchSub); if(showSearchSub){setSubSearchTerm('')} }}
                              >
                                <SVG name={!showSearchSub ? 'search' : 'close'} size={20} color={'grey'} />
                              </Button>
                              <Button
                                type='text'
                                className={styles.btn}
                                onClick={() => { setSortSubInsight(!sortSubInsight) }}
                              >
                                <SVG name={!sortSubInsight ? 'sortdown' : 'sortup'} size={20} color={'grey'} />
                              </Button>
                            </div>
                          </div>
                          <div className={'fa-vertical-scrolling-card px-4'}>
                            {subInsightData ? <SubInsightItem queryResult={queryResult} /> : <NoSubInsightsData />}
                          </div>
                        </div>
                      </div>
                    </Col>
                  </Row>
                </div> </>  : <NoData />
              }  
            </>
            : <div className='mt-6 flex justify-center items-center py-10'>
              <Spin />
            </div>
          }

        </ErrorBoundary>
      </div>
    </div>

  </>
  );
}


const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  templateConfig: state.templates,
});


export default connect(mapStateToProps, { fetchTemplateConfig, fetchTemplateInsights })(TemplateResults);
