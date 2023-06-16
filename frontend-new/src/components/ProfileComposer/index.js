import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { connect, useDispatch } from 'react-redux';
import { bindActionCreators } from 'redux';
import { Button, DatePicker, Tooltip } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import ProfileBlock from './ProfileBlock';
import GroupBlock from './GroupBlock';
import { QUERY_TYPE_PROFILE } from '../../utils/constants';
import ComposerBlock from '../QueryCommons/ComposerBlock';
import {
  getUserProperties,
  getGroupProperties
} from 'Reducers/coreQuery/middleware';
import MomentTz from 'Components/MomentTz';
import FaSelect from '../FaSelect';
import { INITIALIZE_GROUPBY } from '../../reducers/coreQuery/actions';
import { TOOLTIP_CONSTANTS } from '../../constants/tooltips.constans';
import { fetchGroups } from 'Reducers/coreQuery/services';
import GlobalFilter from 'Components/GlobalFilter';

function ProfileComposer({
  queries,
  setQueries,
  runProfileQuery,
  eventChange,
  queryType,
  fetchGroups,
  getUserProperties,
  getGroupProperties,
  activeProject,
  groupOpts,
  queryOptions,
  setQueryOptions,
  collapse = false,
  setCollapse
}) {
  const [isDDVisible, setDDVisible] = useState(false);
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [profileBlockOpen, setProfileBlockOpen] = useState(true);
  const [filterBlockOpen, setFilterBlockOpen] = useState(true);
  const [groupBlockOpen, setGroupBlockOpen] = useState(true);
  const dispatch = useDispatch();

  useEffect(() => {
    fetchGroups(activeProject?.id);
  }, [activeProject?.id]);

  const groupsList = useMemo(() => {
    let groups = [['Users', 'users']];
    Object.entries(groupOpts || {}).forEach(([group_name, display_name]) => {
      groups.push([display_name, group_name]);
    });
    return groups;
  }, [groupOpts]);

  useEffect(() => {
    if (activeProject && activeProject.id) {
      getUserProperties(activeProject.id, queryType);
    }
  }, [activeProject.id]);

  useEffect(() => {
    if (queryOptions.group_analysis === 'users') return;
    getGroupProperties(activeProject.id, queryOptions.group_analysis);
  }, [activeProject.id, queryOptions.group_analysis]);

  const setGroupAnalysis = (group) => {
    if (group !== 'users') {
      getGroupProperties(activeProject.id, group);
    }
    const opts = Object.assign({}, queryOptions);
    opts.group_analysis = group;
    opts.globalFilters = [];
    dispatch({
      type: INITIALIZE_GROUPBY,
      payload: {
        global: [],
        event: []
      }
    });
    setQueryOptions(opts);
  };

  const resetLabel = (group) => {
    const labelMap = ['salesforce', 'hubspot', '6signal', 'linkedin_company', 'g2'];
    const label =
      labelMap.find((key) => group.toLowerCase().includes(key)) || 'web';
    const query = { ...queries, label, alias: '', filters: [] };
    setQueries([query]);
  };

  const onChange = (value) => {
    if (value !== queryOptions.group_analysis) {
      setGroupAnalysis(value);
      resetLabel(value);
    }
    setDDVisible(false);
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
            options={groupsList}
            onClickOutside={() => setDDVisible(false)}
            optionClick={(val) => onChange(val[1])}
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
          </Text>
          <div className={`${styles.groupsection}`}>
            <Tooltip
              title='Select profile type to analyse'
              color={TOOLTIP_CONSTANTS.DARK}
            >
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
                    {
                      groupsList?.find(
                        ([_, groupName]) =>
                          groupName === queryOptions?.group_analysis
                      )?.[0]
                    }
                  </Text>
                  <SVG name='caretDown' />
                </div>
              </Button>
            </Tooltip>

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
            queryOptions={queryOptions}
            setQueryOptions={setQueryOptions}
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
            queryOptions={queryOptions}
            setQueryOptions={setQueryOptions}
          ></ProfileBlock>
        </div>
      );
    }
    return blockList;
  };

  const renderProfileQueryList = () => {
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
            <GlobalFilter
              filters={queryOptions.globalFilters}
              setGlobalFilters={setGlobalFiltersOption}
              groupName={queryOptions.group_analysis}
            />
          </div>
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const groupByBlock = () => {
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
            <GroupBlock groupName={queryOptions.group_analysis} />
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
                      {MomentTz(queryOptions?.date_range?.from).format(
                        'MMM DD, YYYY'
                      )}
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
  groupOpts: state.groups.data
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchGroups,
      getUserProperties,
      getGroupProperties
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(ProfileComposer);
