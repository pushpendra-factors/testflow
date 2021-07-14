import React, { useCallback, useEffect, useState, useContext } from 'react';
import { SVG, Text, FaErrorComp, FaErrorLog, Number } from 'Components/factorsComponents';
import { Button, Tabs, Row, Col, Skeleton, Spin, message, Form, InputNumber, Input } from 'antd';
import { useHistory } from 'react-router-dom';
import { ErrorBoundary } from 'react-error-boundary';
import { fetchTemplateConfig, fetchTemplateInsights } from 'Reducers/templates';
import { connect } from 'react-redux';
import moment from 'moment';

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
      return (
        insights?.map((item, j) => {
          let isIncreased = item?.percentage_change >= 0;
          return (
            <div className={`my-1 py-2 px-4 mx-2 w-full flex items-center justify-between cursor-pointer  border-radius--sm border--thin-2--transparent ${selectedInsight == j ? 'border--thin-2--brand' : ''}`} onClick={() => showSubInsight(item.sub_level_data, j)}>
              <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mr-3 capitalize'}>{item?.name}</Text>
              <div className={'flex items-center'}>
                {item.is_infinity ? <Text type={'title'} level={6} extraClass={'m-0'}>∞</Text> : <>
                  <SVG name={isIncreased ? 'spikeup' : 'spikedown'} color={isIncreased ? 'green' : 'red'} size={18} />
                  <Text type={'title'} level={7} weight={'bold'} color={isIncreased ? 'green' : 'red'} extraClass={'m-0 ml-1'}><Number number={item?.percentage_change} suffix={'%'} /></Text>
                </>
                }
                <Text type={'title'} level={8} color={'grey'} extraClass={'m-0 ml-2'}>{`(`}<Number number={roundNumb(item?.previous_value)} shortHand={true} />{` -> `}<Number number={roundNumb(item?.last_value)} shortHand={true} />{`)`}</Text>
              </div>
            </div>
          )
        })
      )
    }
    else return <NoData />
  }

  const SubInsightItem = () => {
    if (subInsightData) {
      return (
        subInsightData?.map((item, j) => {
          let isIncreased = item?.percentage_change >= 0;
          return (
            <div className={`py-3 px-6 w-full flex items-center justify-between cursor-pointer`}>
              <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mr-3 capitalize'}>{item?.name}</Text>
              <div className={'flex items-center'}>

                {item.is_infinity ? <Text type={'title'} level={6} extraClass={'m-0'}>∞</Text> : <>
                  <SVG name={isIncreased ? 'growthUp' : 'growthDown'} color={isIncreased ? 'green' : 'red'} size={18} />
                  <Text type={'title'} level={7} weight={'bold'} color={isIncreased ? 'green' : 'red'} extraClass={'m-0 ml-1'}><Number number={item?.percentage_change} suffix={'%'} /></Text>
                </>}
                {item?.root_cause_metrics ?
                  <div className={'flex items-center flex-col'}>
                    {/* <Text type={'title'} level={8} color={'grey'} extraClass={'m-0 ml-2'}>Due to</Text> */}
                    {item?.root_cause_metrics?.map((subitem) => {
                      let isIncreased = subitem?.percentage_change >= 0;
                      return <Text type={'title'} level={8} color={'grey'} extraClass={'m-0 ml-2'}>{subitem?.metric} {`${isIncreased ? 'increased' : 'decreased'}`} <Number number={subitem?.percentage_change} suffix={'%'} /></Text>

                    })}
                  </div>
                  : <Text type={'title'} level={8} color={'grey'} extraClass={'m-0 ml-2'}>{`(`}<Number number={roundNumb(item?.previous_value)} shortHand={true} />{` -> `}<Number number={roundNumb(item?.last_value)} shortHand={true} />{`)`}</Text>
                }
              </div>
            </div>
          )
        })
      )
    }
    else return null
  }

  const CardInsights = (data) => {
    const cards = data?.queryResult?.breakdown_analysis?.overall_changes_data;
    if (cards) {
      return (
        <div className={'flex items-center justify-between w-full'}>
          {cards?.map((item, j) => {
            let isIncreased = item?.percentage_change >= 0;
            return (
              <div className={`py-4 px-6 border--thin-2 flex w-full items-center flex-wrap border-radius--sm ${j == 1 ? 'mx-4' : ''}`}>
                <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 mr-3 capitalize'}>{configMatrix.map((element) => { if (element.metric == item.metric) return element.display_name })}</Text>
                {item.is_infinity ? <Text type={'title'} level={5} extraClass={'m-0'}>∞</Text> : <>
                  <SVG name={isIncreased ? 'spikeup' : 'spikedown'} color={isIncreased ? 'green' : 'red'} size={20} />
                  <Text type={'title'} level={6} weight={'bold'} color={isIncreased ? 'green' : 'red'} d extraClass={'m-0 ml-1'}><Number number={item?.percentage_change} suffix={'%'} /></Text>
                </>}
                <Text type={'title'} level={8} color={'grey'} extraClass={'m-0 ml-3'}>{`Overall ${isIncreased ? 'increased' : 'decreased'} from `}<Number number={roundNumb(item?.previous_value)} shortHand={true} />{` to `}<Number number={roundNumb(item?.last_value)} shortHand={true} /></Text>
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
      <img src={'assets/images/add-widget-icon.png'} style={{ maxHeight: '80px' }} />
      <Text type={'title'} level={8} color={'grey'} weight={'thin'} extraClass={'m-0 mr-3'}>{`Select the ${column1} to see the ${column2} data`}</Text>
    </div>)
  }
  const NoData = () => {
    return (<div className={'flex flex-col my-20 py-10 px-5 items-center'}>
      <img src={'assets/images/add-widget-icon.png'} style={{ maxHeight: '80px' }} />
      <Text type={'title'} level={8} color={'grey'} weight={'thin'} extraClass={'m-0 mr-3'}>{`No insights available`}</Text>
    </div>)
  }

  const fetchInsights = (key) => {
    setLoading(true);
    const queryData = {
      "metric": key,
      "from": moment().subtract(1, 'weeks').startOf('week').unix(), //last week timestamp, will calculate previous week in backend
      "to": moment().subtract(1, 'weeks').endOf('week').unix()
    }
    fetchTemplateInsights(activeProject.id, queryData).then(() => {
      setLoading(false);
    }).catch((e) => {
      console.log('fetchTemplateInsights error', e)
      message.error(`Sorry! couldn’t fetch insights for ${key}`)
    });
  }
  useEffect(() => {
    if (templateConfig && configMatrix) {
      // console.log('configMatrix',configMatrix);
      fetchInsights(configMatrix[0].metric);
    }
    else {
      routeChange('/analyse')
    }
  }, []);

  const onTabChange = (key) => {
    setSubInsightData(null)
    setSelectedInsight(null);
    fetchInsights(key);
  }




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
            {`Templates / Template-Name`}
          </Text>
        </div>
      </div>
      <div className='flex items-center'>
      </div>
    </div>

    <div className='mt-24 px-20'>
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
        <Button><SVG name={'calendar'} size={16} extraClass={'mr-1'} />Last Week</Button>
        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mx-2'}>vs</Text>
        <Button><SVG name={'calendar'} size={16} extraClass={'mr-1'} />Prev. Week</Button>
      </div>

        {configMatrix ? <div className='mt-8'>
          <Tabs className={"fa-tabs--dashboard fa-tabs--white capitalize"} defaultActiveKey="1" tabPosition={'top'} onChange={onTabChange} >
            {configMatrix?.map((item, j) => (
              <TabPane tab={`${item.display_name}`} key={item.metric} className={'capitalize'} />
            ))}
          </Tabs>
        </div> : <Skeleton loading={true} active />
        }



        {(queryResult && !loading) ?
          <>
            {queryResult?.queryResult?.breakdown_analysis?.primary_level_data ? <>
            <div className={'my-6 w-full'}>
              <CardInsights queryResult={queryResult} />
            </div>
            <div className='mt-6'>
              <Row gutter={[24, 24]}>
                <Col span={12}>
                  <div className={'pr-4'}>
                    <div className={'py-4 px-6 background-color--brand-color-1 border-radius--sm flex justify-between'}>
                      <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0 capitalize'}>{column1}</Text>
                    </div>
                    <div className={'fa-vertical-scrolling-card'}>
                      <InsightItem queryResult={queryResult} />
                    </div>
                  </div>
                </Col>

                <Col span={12}>
                  <div className={'pl-4'}>
                    <div className={'border--thin-2  border-radius--sm '}>
                      <div className={'py-4 px-6 background-color--brand-color-1 border-radius--sm flex justify-between'}>
                        <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0 capitalize'}>{metaTitle?.sub_level?.column_name}</Text>
                      </div>
                      <div className={'fa-vertical-scrolling-card'}>
                        {subInsightData ? <SubInsightItem queryResult={queryResult} /> : <NoSubInsightsData />}
                      </div>
                    </div>
                  </div>
                </Col>
              </Row>
            </div>
            </>
            : <NoData />
            }
          </>
          : <div className='mt-6 flex justify-center items-center py-10'>
            <Spin />
          </div>
        }

      </ErrorBoundary>
    </div>

  </>
  );
}


const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  templateConfig: state.templates,
});


export default connect(mapStateToProps, { fetchTemplateConfig, fetchTemplateInsights })(TemplateResults);
