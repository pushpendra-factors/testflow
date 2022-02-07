import React, { useState, useEffect, useCallback } from 'react';
import { connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, Collapse, Select, DatePicker } from 'antd';
import { SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';
import ProfileBlock from './ProfileBlock';
import GroupBlock from './GroupBlock';
import { QUERY_TYPE_PROFILE } from '../../utils/constants';
import ComposerBlock from '../QueryCommons/ComposerBlock';
import {
  fetchEventNames,
  getUserProperties,
} from 'Reducers/coreQuery/middleware';
import GLobalFilter from './GlobalFilter';
import MomentTz from 'Components/MomentTz';
import FaSelect from '../FaSelect';
import {
  ProfileGroupMapper,
  revProfileGroupMapper,
} from '../../utils/constants';

function ProfileComposer({
  queries,
  setQueries,
  runProfileQuery,
  eventChange,
  queryType,
  fetchEventNames,
  getUserProperties,
  activeProject,
  queryOptions,
  setQueryOptions,
  collapse = false,
  setCollapse,
}) {
  const [isDDVisible, setDDVisible] = useState(false);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [showDatePickerStr, setShowDatePickerStr] = useState('Select Date');

  const groupOptions = [
    ['Users'],
    ['Salesforce Opportunity'],
    ['Salesforce Accounts'],
    ['Hubspot Deals'],
    ['Hubspot Companies'],
  ];

  useEffect(() => {
    if (activeProject && activeProject.id) {
      fetchEventNames(activeProject.id);
      getUserProperties(activeProject.id, queryType);
    }
  }, [activeProject, fetchEventNames]);

  const setGroupAnalysis = (group) => {
    const opts = Object.assign({}, queryOptions);
    opts.group_analysis = ProfileGroupMapper[group]
      ? ProfileGroupMapper[group]
      : group;
    setQueryOptions(opts);
  };

  const onChange = (value) => {
    setGroupAnalysis(value);
    setDDVisible(false);
    setQueries([]);
  };

  const triggerDropDown = () => {
    setDDVisible(true);
  };

  const selectGroup = () => {
    return (
      <div className={`${styles.groupsection_dropdown}`}>
        {isDDVisible ? (
          <FaSelect
            extraClass={`${styles.groupsection_dropdown_menu}`}
            options={groupOptions}
            onClickOutside={() => setDDVisible(false)}
            optionClick={(val) => onChange(val[0])}
          ></FaSelect>
        ) : null}
      </div>
    );
  };

  const renderGroupSection = () => {
    try {
      return (
        <div className={`flex items-center pt-6`}>
          <Text
            type={'title'}
            level={6}
            weight={'normal'}
            extraClass={`m-0 mr-3`}
          >
            Analyse
          </Text>{' '}
          <div className={`${styles.groupsection}`}>
            <Button
              className={`${styles.groupsection_button}`}
              type='text'
              onClick={triggerDropDown}
            >
              <div className={`flex items-center`}>
                <Text
                  type={'title'}
                  level={6}
                  weight={'bold'}
                  extraClass={`m-0 mr-1`}
                >
                  {queryOptions?.group_analysis
                    ? revProfileGroupMapper[queryOptions.group_analysis]
                    : 'Select Group'}
                </Text>
                <SVG name='caretDown' />
              </div>
            </Button>
            {selectGroup()}
          </div>
        </div>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const queryList = () => {
    const blockList = [];

    queries.forEach((event, index) => {
      blockList.push(
        <div key={index} className={styles.composer_body__query_block}>
          <ProfileBlock
            index={index + 1}
            queryType={queryType}
            event={event}
            queries={queries}
            eventChange={eventChange}
            groupAnalysis={queryOptions.group_analysis}
          ></ProfileBlock>
        </div>
      );
    });

    if (queries.length < 6) {
      blockList.push(
        <div key={'init'} className={styles.composer_body__query_block}>
          <ProfileBlock
            queryType={queryType}
            index={queries.length + 1}
            queries={queries}
            eventChange={eventChange}
            groupBy={queryOptions.groupBy}
            groupAnalysis={queryOptions.group_analysis}
          ></ProfileBlock>
        </div>
      );
    }
    return blockList;
  };

  const renderProfileQueryList = () => {
    const [profileBlockOpen, setProfileBlockOpen] = useState(true);
    try {
      return (
        <ComposerBlock
          blockTitle={'PROFILES TO ANALYSE'}
          isOpen={profileBlockOpen}
          showIcon={true}
          onClick={() => setProfileBlockOpen(!profileBlockOpen)}
          extraClass={`pt-2 no-padding-l no-padding-r`}
        >
          {queryList()}
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const setGlobalFiltersOption = (filters) => {
    const opts = Object.assign({}, queryOptions);
    opts.globalFilters = filters;
    setQueryOptions(opts);
  };

  const renderGlobalFilterBlock = () => {
    const [filterBlockOpen, setFilterBlockOpen] = useState(true);
    try {
      if (queryType === QUERY_TYPE_PROFILE && queries.length < 1) {
        return null;
      }
      return (
        <ComposerBlock
          blockTitle={'FILTER BY'}
          isOpen={filterBlockOpen}
          showIcon={true}
          onClick={() => setFilterBlockOpen(!filterBlockOpen)}
          extraClass={`no-padding-l no-padding-r`}
        >
          <div key={0} className={'fa--query_block borderless no-padding '}>
            <GLobalFilter
              filters={queryOptions.globalFilters}
              setGlobalFilters={setGlobalFiltersOption}
              onFiltersLoad={[
                () => {
                  getUserProperties(activeProject.id, queryType);
                },
              ]}
            ></GLobalFilter>
          </div>
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const groupByBlock = () => {
    const [groupBlockOpen, setGroupBlockOpen] = useState(true);
    try {
      if (queryType === QUERY_TYPE_PROFILE && queries.length < 1) {
        return null;
      }
      return (
        <ComposerBlock
          blockTitle={'BREAKDOWN'}
          isOpen={groupBlockOpen}
          showIcon={true}
          onClick={() => setGroupBlockOpen(!groupBlockOpen)}
          extraClass={`no-padding-l no-padding-r`}
        >
          <div key={0} className={'fa--query_block borderless no-padding '}>
            <GroupBlock queryType={queryType} events={queries}></GroupBlock>
          </div>
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const setDateSince = (val) => {
    let dateT;
    let dateValue = {};
    const queryOptionsState = Object.assign({}, queryOptions);
    dateT = MomentTz(val).startOf('day');
    dateValue['from'] = dateT.toDate().getTime();
    queryOptionsState.date_range.from = dateValue.from;
    setShowDatePickerStr(MomentTz(val).format('MMM DD, YYYY'));
    setQueryOptions(queryOptionsState);
    setShowDatePicker(false);
  };

  const handleRunQuery = useCallback(() => {
    if (queryType === QUERY_TYPE_PROFILE) {
      runProfileQuery(false);
    }
  }, [runProfileQuery, queryType]);

  const renderFooter = () => {
    try {
      if (queryType === QUERY_TYPE_PROFILE && queries.length < 1) {
        return null;
      } else {
        return (
          <div
            className={
              !collapse ? styles.composer_footer : styles.composer_footer_right
            }
          >
            {!collapse ? (
              <div className={'flex items-center'}>
                <Text
                  type={'title'}
                  level={7}
                  weight={'bold'}
                  extraClass={'m-0 mr-2'}
                >
                  Created Since
                </Text>
                <div className={`fa-custom-datepicker`}>
                  {!showDatePicker ? (
                    <Button
                      onClick={() => {
                        setShowDatePicker(true);
                      }}
                    >
                      <SVG name={'calendar'} size={16} extraClass={'mr-1'} />
                      {showDatePickerStr}
                    </Button>
                  ) : (
                    <Button>
                      <SVG name={'calendar'} size={16} extraClass={'mr-1'} />
                      <DatePicker
                        format={'MMM DD YYYY'}
                        style={{ width: '96px' }}
                        disabledDate={(d) => !d || d.isAfter(MomentTz())}
                        dropdownClassName={'fa-custom-datepicker--datepicker'}
                        size={'small'}
                        suffixIcon={null}
                        showToday={false}
                        bordered={false}
                        autoFocus={true}
                        allowClear={false}
                        open={true}
                        onOpenChange={() => {
                          setShowDatePicker(false);
                        }}
                        onChange={setDateSince}
                      />
                    </Button>
                  )}
                </div>
              </div>
            ) : (
              <Button
                className={`mr-2`}
                size={'large'}
                type={'default'}
                onClick={() => setCollapse(false)}
              >
                <SVG name={`arrowUp`} size={20} extraClass={`mr-1`}></SVG>
                Collapse all
              </Button>
            )}
            <Button
              className={`ml-2`}
              size={'large'}
              type='primary'
              onClick={handleRunQuery}
            >
              Run Analysis
            </Button>
          </div>
        );
      }
    } catch (err) {
      console.log(err);
    }
  };

  return (
    <div className={styles.composer_body}>
      {renderGroupSection()}
      {renderProfileQueryList()}
      {renderGlobalFilterBlock()}
      {groupByBlock()}
      {renderFooter()}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchEventNames,
      getUserProperties,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(ProfileComposer);
