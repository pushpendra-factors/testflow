import React, { useState } from "react";
import styles from "./index.module.scss";

import { connect } from "react-redux";
import { bindActionCreators } from "redux";

import {
  setShowCriteria,
  setPerformanceCriteria,
} from "../../../reducers/analyticsQuery";

import { SVG, Text } from "../../factorsComponents";
import { QUERY_TYPE_EVENT, QUERY_TYPE_FUNNEL, TOTAL_EVENTS_CRITERIA, TOTAL_USERS_CRITERIA, ACTIVE_USERS_CRITERIA, FREQUENCY_CRITERIA } from '../../../utils/constants';
import FaSelect from '../../FaSelect';

import {Button} from 'antd';

const CriteriaSection = ({queryType, queryCount = 0, crit_show, crit_perf, setShowCriteria, setPerformanceCriteria}) => {

    const [critShowSelect, setCritShowSelect] = useState(false);
    const [critPerfSelect, setCritPerfSelect] = useState(false);

    const CRITERIA_SHOW_OPTIONS = [
        ['Total Events', null, TOTAL_EVENTS_CRITERIA],
        ['Total Users', null, TOTAL_USERS_CRITERIA], 
        ['Active Users', null, ACTIVE_USERS_CRITERIA],
        ['Frequency', null, FREQUENCY_CRITERIA],
    ];

    const CRITERIA_PERF_OPTIONS = [
        ['Any Event', null, 'any'], 
        ['Each Event', null, 'each'],
        ['All Events', null, 'all']
    ];

    const renderCritPerf = () => {
        if(!crit_show || crit_show !== TOTAL_USERS_CRITERIA || queryCount <=1 ) return null;

        return (
            <div className={`flex items-center`}>

                <Text type={"title"} level={7} extraClass={"m-0 mr-2 inline"}>
                  who performed
                </Text>

                <div className={``}>
                    <Button 
                            size={'large'} 
                            type="link" 
                            onClick={() => setCritPerfSelect(!critPerfSelect)}>
                                {crit_perf? CRITERIA_PERF_OPTIONS.filter((op) => op[2] === crit_perf)[0][0] : 'Select'}</Button>

                    {critPerfSelect && <FaSelect 
                            options={CRITERIA_PERF_OPTIONS}
                            optionClick={(op) => {
                                setPerformanceCriteria(op[2])
                                setCritPerfSelect(false);
                            }}
                            onClickOutside={() => setCritPerfSelect(false)}
                        >

                        </FaSelect>}

                </div>

                
            </div>
        )
    }

    const renderCritShow = () => {
        return (<div className={`mr-2 items-center`}>
            <Button 
                    size={'large'} 
                    type="link" 
                    onClick={() => setCritShowSelect(!critShowSelect)}>
                        {crit_show? CRITERIA_SHOW_OPTIONS.filter((op) => op[2] === crit_show)[0][0]: crit_show}
            </Button>

            {critShowSelect && 
                <FaSelect 
                    options={CRITERIA_SHOW_OPTIONS}
                    optionClick={(op) => {
                        setShowCriteria(op[2]);
                        setCritShowSelect(false);
                    }}
                    onClickOutside={() => setCritShowSelect(false)}
                >

                </FaSelect>
            }

        </div>);
    }

    if(queryType == QUERY_TYPE_EVENT) {
        return (<div className={styles.criteria}>
            <Text type={"title"} level={7} extraClass={"m-0 mr-2 inline"}>
                  Show
            </Text>

            {renderCritShow()}

            {renderCritPerf()}
            
        </div>)
    } else {
        return null;
    }

}

const mapStateToProps = (state) => ({
  crit_show: state.analyticsQuery.show_criteria,
  crit_perf: state.analyticsQuery.performance_criteria,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setShowCriteria,
      setPerformanceCriteria,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(CriteriaSection);
