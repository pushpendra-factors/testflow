import React, { useState } from 'react';
import styles from './index.module.scss';

import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import {
  setShowCriteria,
  setPerformanceCriteria
} from '../../../reducers/analyticsQuery';

import { Text } from '../../factorsComponents';
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  TOTAL_USERS_CRITERIA
} from '../../../utils/constants';
import FaSelect from '../../FaSelect';

import { Button } from 'antd';
import FunnelsConversionDurationBlock from '../FunnelsConversionDurationBlock/FunnelsConversionDurationBlock';

const CriteriaSection = ({
  queryType,
  crit_show,
  crit_perf,
  groupByState,
  setPerformanceCriteria
}) => {
  const [critPerfSelect, setCritPerfSelect] = useState(false);

  const CRITERIA_PERF_OPTIONS = [
    ['Any Event', null, 'any'],
    ['Each Event', null, 'each'],
    ['All Events', null, 'all']
  ];

  const isEventGroupSelected = () => {
    if (groupByState.event?.length) return true;
    return false;
  };

  const renderCritPerf = () => {
    if (!crit_show || crit_show !== TOTAL_USERS_CRITERIA) return null;

    return (
      <div className={`flex items-center`}>
        <Text type={'title'} level={7} extraClass={'m-0 mr-2 inline'}>
          Who performed
        </Text>

        <div className={`relative fa-button--truncate`}>
          <Button
            type='link'
            onClick={() => setCritPerfSelect(!critPerfSelect)}
          >
            {crit_perf
              ? CRITERIA_PERF_OPTIONS.filter((op) => op[2] === crit_perf)[0][0]
              : 'Select'}
          </Button>

          {critPerfSelect && (
            <FaSelect
              options={CRITERIA_PERF_OPTIONS.filter((op, i) => {
                if (isEventGroupSelected() && i === 0) return false;
                return true;
              })}
              optionClick={(op) => {
                setPerformanceCriteria(op[2]);
                setCritPerfSelect(false);
              }}
              onClickOutside={() => setCritPerfSelect(false)}
            ></FaSelect>
          )}
        </div>
      </div>
    );
  };

  if (queryType === QUERY_TYPE_EVENT) {
    return <div className={`${styles.criteria} mt-2`}>{renderCritPerf()}</div>;
  }

  if (queryType === QUERY_TYPE_FUNNEL) {
    return (
      <div className={`${styles.criteria} mt-2`}>
        <Text
          color='grey-2'
          type={'title'}
          level={7}
          weight='medium'
          extraClass={'m-0 mr-2 inline'}
        >
          Conversion within
        </Text>
        <FunnelsConversionDurationBlock />
      </div>
    );
  }

  return null;
};

const mapStateToProps = (state) => ({
  crit_show: state.analyticsQuery.show_criteria,
  crit_perf: state.analyticsQuery.performance_criteria,
  groupByState: state.coreQuery.groupBy
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setShowCriteria,
      setPerformanceCriteria
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(CriteriaSection);
