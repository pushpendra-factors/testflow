import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Spin, Tag, Input
} from 'antd';
import { fetchFactorsGoals, fetchFactorsModels, fetchGoalInsights, saveGoalInsightRules, fetchFactorsTrackedEvents, fetchFactorsTrackedUserProperties } from 'Reducers/factors';
import { fetchEventNames, getUserProperties } from 'Reducers/coreQuery/middleware';
import { connect } from 'react-redux';
import { fetchProjectAgents } from 'Reducers/agentActions';
import _, { isEmpty } from 'lodash';
import { useHistory } from 'react-router-dom';
import { Text, SVG, FaErrorComp, FaErrorLog, Number } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import L2Modal from './ModalL2';
import CardInsight from './CardInsight';
import ExplainData from './Data';

const CardInsightWrapper = ({ data }) => {
  return (
    <>
      <div className={'flex items-center flex-col'}>
        <div className={'flex items-center justify-center explain-insight--wrapper'}>
          <CardInsight
            title={data?.goal?.st_en ? data?.goal?.st_en : "All Visitors"}
            count={data?.total_users_count}
            arrow={true}
            tagTitle={`A`}
          />
          <CardInsight
            title={data?.goal?.en_en}
            count={data?.goal_user_count}
            arrow={false}
            tagTitle={`B`}
          />
        </div>
        <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0 mt-4 flex items-center'}>
          <Number suffix={'%'} number={data?.overall_percentage} className={'mr-1'} />
          {` Conversions from A to B`}
        </Text>
      </div>
    </>
  )
}


const InsightColumnTitle = ({ title, isIncreased, resultTitle }) => {
  return (
    <div className={`flex items-center justify-between explain-table--row`}>
      <div className={`py-2 px-4 flex items-center`}>
        <SVG name={isIncreased ? 'spikeup' : 'spikedown'} color={isIncreased ? 'green' : 'red'} size={22} />
        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 ml-2 capitalize'}>{title}</Text>

      </div>
      <div className={'py-2 px-4 flex column_right'}>
        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0 capitalize'}>{resultTitle}</Text>
      </div>
    </div>

  )
}

const InsightItem = ({ data, sort, setModalL2, showModalL2, showIncrease = false, isAttribute, setModalData, searchTerm='' }) => {

  let dataSet = sort ? data?.insights?.sort() : data?.insights?.reverse();

  return dataSet?.map((item) => {
    if ((item?.factors_insights_type == "journey" || item?.factors_insights_type == "campaign") && !isAttribute) {
      if (item?.factors_multiplier_increase_flag == showIncrease && item?.factors_insights_key.toLowerCase().includes(searchTerm) ) {
        return (
          <div className={`flex items-center justify-between cursor-pointer explain-table--row`} onClick={() => {
            setModalL2(true)
            setModalData(item)
          }}>
            <div className={`py-2 px-4 flex items-center `}>
              <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mr-3'}>{item?.factors_insights_type == "journey" ? `Users who visit` : `Users from`}</Text>
              <Tag className={'m-0 mx-2'} className={'fa-tag--regular fa-tag--highlight truncate'} style={{ maxWidth: '350px' }}>{item?.factors_insights_key}</Tag>

            </div>
            <div className={'py-2 px-4 flex justify-end column_right'}>
              <Tag color={item?.factors_multiplier_increase_flag ? 'green' : "red"} className={`m-0 mx-1 ${item?.factors_multiplier_increase_flag ? 'fa-tag--green' : "fa-tag--red"}`}>
                <Number suffix={'%'} number={item?.factors_insights_percentage} />
              </Tag>
            </div>
          </div>

        )
      }
      else return null
    }
    if (item?.factors_insights_type == "attribute" && isAttribute) {
      if (item?.factors_multiplier_increase_flag == showIncrease && item?.factors_insights_attribute[0]?.factors_attribute_key.toLowerCase().includes(searchTerm)) {
        return (
          <div className={`flex items-center justify-between cursor-pointer explain-table--row`} onClick={() => {
            setModalL2(true)
            setModalData(item)
          }}>
            <div className={`py-2 px-4 flex items-center `}>
              <Tag className={'m-0 mx-2'} className={'fa-tag--regular fa-tag--highlight truncate'} style={{ maxWidth: '350px' }}>
                {`${item?.factors_insights_attribute[0]?.factors_attribute_key}`}
              </Tag>
              <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mr-3'}>
                {`= ${item?.factors_insights_attribute[0]?.factors_attribute_value}`}
              </Text>
            </div>
            <div className={'py-2 px-4 flex justify-end column_right'}>
              <Tag color={item?.factors_multiplier_increase_flag ? 'green' : "red"} className={`m-0 mx-1 ${item?.factors_multiplier_increase_flag ? 'fa-tag--green' : "fa-tag--red"}`}>
                <Number suffix={'%'} number={item?.factors_insights_percentage} />
              </Tag>
            </div>
          </div>

        )
      }
      else return null
    }
    else return null
  }
  )
}

const InsightTable = ({ 
  data,  
  setModalL2, 
  showModalL2, 
  isAttribute = false, 
  setModalData,
}) => {

  const [tableData, setTableData] = useState(false);
  const [showSearch, setShowSearch] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');
  const [sort, setSort] = useState(true);


  const onInputSearch = (userInput) => {
    let searchWord = userInput.currentTarget.value.toLowerCase();
    setSearchTerm(searchWord);
  };

  const onSortChange = () => { 
      setSort(!sort)
  } 
 
  return (
    <Row gutter={[24, 24]}>
      <Col span={24}>
        <div className={'my-5'}>
          <div className={'border--thin-2  border-radius--sm'}>
            <div className={'py-4 px-6 background-color--brand-color-1 border-radius--sm flex justify-between'}>
              <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0 capitalize'}>{`${isAttribute ? 'Segments (Attributes)' : `Engagements (Journeys + Campaigns)`}`}</Text>
              <div className={'flex justify-between'}>
                {showSearch ? <Input
                  onChange={onInputSearch} 
                  prefix={(<SVG name="search" size={16} color={'grey'} />)}
                /> : null}
                <Button className='fa-button-ghost' type='text' onClick={()=>{
                    if(!showSearch){ 
                      setShowSearch(true) 
                    }else {
                      setShowSearch(false)  
                      setSearchTerm('');
                    }
                  }}> <SVG name={!showSearch ? 'search' : 'close'} size={20} color={'grey'} /> </Button>
                {/* <Button className='fa-button-ghost' type='text' onClick={()=>onSortChange()} > <SVG name={ sort ? 'sortdown' : 'sortup'} size={20} color={'grey'} /> </Button> */}
              </div>
            </div>
            <div className={'explain-insight--container'}>
              <Row gutter={[0, 0]}>
                <Col span={12}> 
                  <div>

                    <InsightColumnTitle
                      title={`What’s working well?`}
                      isIncreased={true}
                      resultTitle={`Conversion`}
                    />
                    <InsightItem
                      setModalL2={setModalL2}
                      showModalL2={showModalL2}
                      showIncrease={true}
                      isAttribute={isAttribute}
                      setModalData={setModalData}
                      data={data}
                      searchTerm={searchTerm}
                      sort={sort}
                    />

                  </div>
                </Col>
                <Col span={12}>
                  <div className='explain-insight--item-right'>
                    <InsightColumnTitle
                      title={`What’s working poorly?`}
                      isIncreased={false}
                      resultTitle={`Conversion`}
                    />
                    <InsightItem
                      setModalL2={setModalL2}
                      showModalL2={showModalL2}
                      showIncrease={false}
                      isAttribute={isAttribute}
                      setModalData={setModalData}
                      data={data}
                      searchTerm={searchTerm}
                      sort={sort}
                    />
                  </div>
                </Col>
              </Row>
            </div>
          </div>
        </div>
      </Col>
    </Row>
  )
}



const ResultsTableL1 = ({ goalInsights }) => {
  const [loading, setLoading] = useState(false);
  const [subInsightData, setSubInsightData] = useState(null);
  const [showSearch, setShowSearch] = useState(false);
  const [showSearchSub, setShowSearchSub] = useState(false); 
  const [sortSubInsight, setSortSubInsight] = useState(false);
  const [showModalL2, setModalL2] = useState(false);
  const [modalData, setModalData] = useState(false);


  const history = useHistory();

  const handleCancel = () => {
    setConfigureDPModal(false);
  };

  return (
    <>
      <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Explain Error '} subtitle={'We are facing trouble loading Explain. Drop us a message on the in-app chat.'} />} onError={FaErrorLog}>


        {(!loading) ?
          <>
            {true ? <>
              <div id={`explain-results`} className={'my-6 w-full flex justify-center mt-10'}>
                <CardInsightWrapper data={goalInsights} />
              </div>

              <L2Modal
                setModalL2={setModalL2}
                showModalL2={showModalL2}
                setModalData={setModalData}
                modalData={modalData}
                data={goalInsights}
              />

              <div className='mt-6'>

                <InsightTable showSearch={showSearch} showModalL2={showModalL2} setModalL2={setModalL2} isAttribute={false} setModalData={setModalData} data={goalInsights} />
                <InsightTable showSearch={showSearch} showModalL2={showModalL2} setModalL2={setModalL2} isAttribute={true} setModalData={setModalData} data={goalInsights} />

              </div> </> : <NoData />
            }
          </>
          : <div className='mt-6 flex justify-center items-center py-10'>
            <Spin />
          </div>
        }

      </ErrorBoundary>
    </>
  );
};
const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project,
    goals: state.factors.goals,
    agents: state.agent.agents,
    factors_models: state.factors.factors_models,
  };
};
export default connect(mapStateToProps, { fetchFactorsGoals, fetchFactorsTrackedEvents, fetchFactorsTrackedUserProperties, fetchProjectAgents, saveGoalInsightRules, fetchGoalInsights, fetchFactorsModels, fetchEventNames, getUserProperties })(ResultsTableL1);
