import React, { useState, useEffect, memo, useMemo } from 'react';
import { connect, useSelector } from 'react-redux';
import { useParams } from 'react-router-dom';
import { bindActionCreators } from 'redux';
import { Button } from 'antd';
import { fetchEventNames } from 'Reducers/coreQuery/middleware';
import { resetGroupByAction } from 'Reducers/coreQuery/actions';
import { areKpisInSameGroup } from 'Utils/kpiQueryComposer.helpers';
import _, { isEqual } from 'lodash';
import { getKPIPropertyMappings } from 'Reducers/kpi';
import { ReactSortable } from 'react-sortablejs';
import { SVG } from '../factorsComponents';
import styles from './index.module.scss';
import QueryBlock from './QueryBlock';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_KPI,
  INITIAL_SESSION_ANALYTICS_SEQ,
  QUERY_OPTIONS_DEFAULT_VALUE
} from '../../utils/constants';
import FaDatepicker from '../FaDatepicker';
import ComposerBlock from '../QueryCommons/ComposerBlock';
import CriteriaSection from './CriteriaSection';
import { getValidGranularityOptions } from '../../utils/dataFormatter';
import { DefaultDateRangeFormat } from '../../Views/CoreQuery/utils';
import GlobalFilterBlock from './GlobalFilterBlock';
import GroupByBlock from './GroupByBlock';

function KPIComposer({
  queries,
  setQueries,
  eventChange,
  queryType,
  activeProject,
  queryOptions,
  setQueryOptions,
  handleRunQuery,
  collapse = false,
  setCollapse,
  selectedMainCategory,
  setSelectedMainCategory,
  KPIConfigProps,
  resetGroupByAction,
  getKPIPropertyMappings,
  propertyMaps
}) {
  const [criteriaBlockOpen, setCriteriaBlockOpen] = useState(true);
  const [eventBlockOpen, setEventBlockOpen] = useState(true);
  const { groupBy } = useSelector((state) => state.coreQuery);
  const [sameGrpState, setSameGrpState] = useState(null);
  const { query_id } = useParams();

  const DefaultQueryOptsVal = {
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    session_analytics_seq: INITIAL_SESSION_ANALYTICS_SEQ,
    date_range: { ...DefaultDateRangeFormat }
  };

  useEffect(() => {
    setSelectedMainCategory(queries[0]);
  }, [queries, setSelectedMainCategory]);

  const handleEventChange = (...props) => {
    eventChange(...props);
  };

  const queryList = () => {
    const blockList = [];

    queries.forEach((event, index) => {
      blockList.push(
        <div key={index} className={styles.composer_body__query_block}>
          <QueryBlock
            index={index + 1}
            queryType={queryType}
            event={event}
            queries={queries}
            eventChange={handleEventChange}
            selectedMainCategory={selectedMainCategory}
            setSelectedMainCategory={setSelectedMainCategory}
            KPIConfigProps={KPIConfigProps}
          />
        </div>
      );
    });
    return blockList;
  };

  const isSameKPIGrp = useMemo(
    () => areKpisInSameGroup({ kpis: queries }),
    [queries]
  );

  useEffect(() => {
    // Filters gets reset when ever query groups are changed
    setSameGrpState(isSameKPIGrp);
    if (!isSameKPIGrp && !_.isEmpty(queries)) {
      // fetch common properties if different group
      if (queries.length > 1 && !isSameKPIGrp) {
        const payload = queries?.map((item) => ({
          name: item?.group,
          derived_kpi: item?.qt == 'derived'
        }));
        getKPIPropertyMappings(activeProject?.id, payload).catch((err) => {
          console.log('getKPIPropertyMappings err', err);
          return null;
        });
      }
    }
    // if(_.isEmpty(queries)){
    //   setQueryOptions((currState) => {
    //     return {
    //       ...currState,
    //       globalFilters: DefaultQueryOptsVal.globalFilters
    //     };
    //   });
    //   resetGroupByAction();
    // }
    if (!_.isNull(sameGrpState)) {
      if (sameGrpState != isSameKPIGrp) {
        setQueryOptions((currState) => ({
          ...currState,
          globalFilters: DefaultQueryOptsVal.globalFilters
        }));
        resetGroupByAction();
      }
    }
  }, [isSameKPIGrp]);

  const setGlobalFiltersOption = (filters) => {
    const opts = { ...queryOptions };
    opts.globalFilters = filters;
    setQueryOptions(opts);
  };

  const setDateRange = (dates) => {
    const queryOptionsState = { ...queryOptions };
    if (dates && dates.startDate && dates.endDate) {
      if (Array.isArray(dates.startDate)) {
        queryOptionsState.date_range.from = dates.startDate[0];
        queryOptionsState.date_range.to = dates.startDate[1];
      } else {
        queryOptionsState.date_range.from = dates.startDate;
        queryOptionsState.date_range.to = dates.endDate;
      }
      const frequency = getValidGranularityOptions({
        from: queryOptionsState.date_range.from,
        to: queryOptionsState.date_range.to
      })[0];
      queryOptionsState.date_range.frequency = frequency;
      setQueryOptions(queryOptionsState);
    }
  };

  const handleRunQueryCamp = () => {
    handleRunQuery();
  };

  const footer = () => {
    try {
      if (queryType === QUERY_TYPE_KPI && queries.length === 0) {
        return null;
      }
      return (
        <div
          className={
            !collapse ? styles.composer_footer : styles.composer_footer_right
          }
        >
          {!collapse ? (
            !query_id && (
              <FaDatepicker
                customPicker
                presetRange
                monthPicker
                quarterPicker
                yearPicker
                placement={
                  areKpisInSameGroup({ kpis: queries })
                    ? 'topRight'
                    : 'bottomRight'
                }
                buttonSize='large'
                range={{
                  startDate: queryOptions.date_range.from,
                  endDate: queryOptions.date_range.to
                }}
                onSelect={setDateRange}
              />
            )
          ) : (
            <Button
              className='mr-2'
              size='large'
              type='default'
              onClick={() => setCollapse(false)}
            >
              <SVG name='arrowUp' size={20} extraClass='mr-1' />
              Collapse all
            </Button>
          )}
          <Button
            className='ml-2'
            size='large'
            type='primary'
            onClick={handleRunQueryCamp}
          >
            Run Analysis
          </Button>
        </div>
      );
    } catch (err) {
      console.log(err);
    }
  };

  const renderEACrit = () => (
    <div>
      <CriteriaSection
        queryCount={queries.length}
        queryType={QUERY_TYPE_EVENT}
      />
    </div>
  );

  const renderCriteria = () => {
    try {
      if (queryType === QUERY_TYPE_EVENT) {
        if (queries.length <= 0) return null;

        return (
          <ComposerBlock
            blockTitle='CRITERIA'
            isOpen={criteriaBlockOpen}
            showIcon
            onClick={() => {
              setCriteriaBlockOpen(!criteriaBlockOpen);
            }}
            extraClass='no-padding-l'
          >
            <div className={styles.criteria}>{renderEACrit()}</div>
          </ComposerBlock>
        );
      }
      if (queryType === QUERY_TYPE_FUNNEL) {
        return null;
      }
    } catch (err) {
      console.log(err);
    }
  };

  const renderQueryList = () => {
    try {
      return (
        <ComposerBlock
          blockTitle='KPI TO ANALYSE'
          isOpen={eventBlockOpen}
          showIcon
          onClick={() => setEventBlockOpen(!eventBlockOpen)}
          extraClass='no-padding-l'
        >
          <ReactSortable
            list={queries}
            setList={(newQueriesState) => {
              if (!isEqual(queries, newQueriesState)) {
                setQueries(newQueriesState);
              }
            }}
          >
            {queryList()}
          </ReactSortable>
          {queries.length < 10 && (
            <div key='init' className={styles.composer_body__query_block}>
              <QueryBlock
                queryType={queryType}
                index={queries.length + 1}
                queries={queries}
                eventChange={handleEventChange}
                groupBy={queryOptions.groupBy}
                selectedMainCategory={selectedMainCategory}
                setSelectedMainCategory={setSelectedMainCategory}
                KPIConfigProps={KPIConfigProps}
              />
            </div>
          )}
        </ComposerBlock>
      );
    } catch (err) {
      console.log(err);
    }
  };

  return (
    <div className={styles.composer_body}>
      {renderQueryList()}
      <GlobalFilterBlock
        queryType={queryType}
        queries={queries}
        queryOptions={queryOptions}
        setGlobalFiltersOption={setGlobalFiltersOption}
        activeProject={activeProject}
        selectedMainCategory={selectedMainCategory}
        KPIConfigProps={KPIConfigProps}
        setQueryOptions={setQueryOptions}
        DefaultQueryOptsVal={DefaultQueryOptsVal}
      />
      <GroupByBlock
        queryType={queryType}
        queries={queries}
        selectedMainCategory={selectedMainCategory}
        KPIConfigProps={KPIConfigProps}
        groupBy={groupBy}
        resetGroupByAction={resetGroupByAction}
      />
      {renderCriteria()}
      {footer()}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  propertyMaps: state.kpi.kpi_property_mapping
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      fetchEventNames,
      resetGroupByAction,
      getKPIPropertyMappings
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(memo(KPIComposer));
