import React, { memo, useContext, useEffect, useState } from 'react';
import _ from 'lodash';
import PropTypes from 'prop-types';
import { SVG } from 'factorsComponents';
import { Button, Tooltip } from 'antd';
import { BUTTON_TYPES } from '../../constants/buttons.constants';
import ControlledComponent from '../ControlledComponent';
import styles from './index.module.scss';
import {QuestionCircleOutlined} from "@ant-design/icons"
import {
  CHART_TYPE_SPARKLINES,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_FUNNEL
} from '../../utils/constants';
import FaSelect from 'Components/FaSelect';
import { getChartType } from '../../Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import { CoreQueryContext } from '../../contexts/CoreQueryContext';
import userflow from 'userflow.js';
import {USERFLOW_CONFIG_ID} from 'Utils/userflowConfig'


const QueryActionsComponent = ({
  queryType,
  breakdown,
  chartTypes,
  savedQueryId,
  attributionModels,
  handleSaveClick,
  handleEditClick,
  handleDeleteClick,
  toggleAddToDashboardModal,
  setShowShareToEmailModal,
  setShowShareToSlackModal
}) => {

  
  const [hideIntercomState, setHideIntercomState] = useState(true);
  let [helpMenu, setHelpMenu] = useState(false)

  useEffect(() => {
    if (window.Intercom) {
      window.Intercom('update', { hide_default_launcher: true });
    }
    return () => {
      if (window.Intercom) {
        window.Intercom('update', { hide_default_launcher: false });
      }
    };
  }, []);
  

  let handleIntercomHelp = ()=>{
      const w = window;
      const ic = w.Intercom;
      if (typeof ic === 'function') {
        setHideIntercomState(!hideIntercomState);
        ic('update', { hide_default_launcher: !hideIntercomState });
        ic(!hideIntercomState === true ? 'hide' : 'show');
      }

  }
  const [options, setOptions] = useState(false);
  const [chart, setChart] = useState(null);
  useEffect(() => {
    setChart(
      getChartType({ queryType, chartTypes, breakdown, attributionModels })
    );
  }, [queryType, chartTypes, breakdown, attributionModels]);
  const setActions = (opt) => {
    if (opt[1] === 'envelope' && opt[2] !== 'disabled') {
      setShowShareToEmailModal(true);
    } else if (opt[1] === 'SlackStroke' && opt[2] !== 'disabled') {
      setShowShareToSlackModal(true);
    } else if (opt[1] === 'edit') {
      handleEditClick();
    } else if (opt[1] === 'trash') {
      handleDeleteClick();
    }else if(opt[1] === 'intercom_help'){
      handleIntercomHelp();
    }else if(opt[1] === 'help_doc'){
      window.open('https://help.factors.ai/','_blank')
    }
    setOptions(false);
  };
  const getActionsMenu = () => {
    return options ? (
      queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_KPI ? (
        <FaSelect
          extraClass={styles.additionalops}
          options={[
            chart === CHART_TYPE_SPARKLINES &&
            (queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_KPI)
              ? ['Email this report', 'envelope']
              : ['Email this report', 'envelope', 'disabled'],
            chart === CHART_TYPE_SPARKLINES &&
            (queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_KPI)
              ? ['Share to slack', 'SlackStroke']
              : ['Share to slack', 'SlackStroke', 'disabled'],
            ['Edit Details', 'edit'],
            ['Delete', 'trash']
          ]}
          optionClick={(val) => setActions(val)}
          onClickOutside={() => setOptions(false)}
          posRight={true}
        ></FaSelect>
      ) : (
        <FaSelect
          extraClass={styles.additionalops}
          options={[
            ['Edit Details', 'edit'],
            ['Delete', 'trash']
          ]}
          optionClick={(val) => setActions(val)}
          onClickOutside={() => setOptions(false)}
          posRight={true}
        ></FaSelect>
      )
    ) : null;
  };
  const getHelpMenu = () => {
    return helpMenu === false ? '' : <FaSelect
    extraClass={styles.additionalops}
    options={[
      ['Help and Support', 'help_doc'],
      ['Talk to us', 'intercom_help']
    ]}
    optionClick={(val) => setActions(val)}
    onClickOutside={() => setHelpMenu(false)}
    posRight={true}
  ></FaSelect>;
    
  };

  const triggerUserFlow  = () =>{

    if(queryType == QUERY_TYPE_ATTRIBUTION || queryType == QUERY_TYPE_FUNNEL || queryType == QUERY_TYPE_KPI){

      let flowID = "";
      if(queryType == QUERY_TYPE_ATTRIBUTION) {flowID = USERFLOW_CONFIG_ID?.AttributionQueryBuilder};
      if(queryType == QUERY_TYPE_FUNNEL){flowID = USERFLOW_CONFIG_ID?.FunnelSQueryBuilder};
      if(queryType == QUERY_TYPE_KPI){flowID = USERFLOW_CONFIG_ID?.KPIQueryBuilder};

      userflow.start(flowID)
  }
  }

  return (
    <div className="flex gap-x-2 items-center">
      <ControlledComponent controller={!savedQueryId}>
        <Button
          onClick={handleSaveClick}
          type={BUTTON_TYPES.PRIMARY}
          size={'large'}
          icon={<SVG name={'save'} size={20} color={'white'} />}
        >
          {'Save'}
        </Button>
      </ControlledComponent>

      {(queryType == QUERY_TYPE_ATTRIBUTION || queryType == QUERY_TYPE_FUNNEL || queryType == QUERY_TYPE_KPI) && <>
        <Tooltip placement="bottom" title="Walk me through">
          <Button
            onClick={triggerUserFlow} 
            size="large"
            type="text"
            icon={<SVG name={'Handshake'} size={24} color={'grey'} />}
          /> 
        </Tooltip>
        </>
        }

      <ControlledComponent controller={!!savedQueryId}>
        <Tooltip placement="bottom" title="Save as New">
          <Button
            onClick={handleSaveClick}
            size="large"
            type="text"
            icon={<SVG name={'pluscopy'} />}
          ></Button>
        </Tooltip>
        <ControlledComponent controller={queryType !== QUERY_TYPE_PROFILE}>
          <Tooltip placement="bottom" title="Add to Dashboard">
            <Button
              onClick={toggleAddToDashboardModal}
              size="large"
              type="text"
              icon={<SVG name={'addtodash'} />}
            ></Button>
          </Tooltip>
        </ControlledComponent>
        <div className={'relative'}>
          <Button
            size="large"
            type="text"
            icon={<SVG name={'threedot'} />}
            onClick={() => setOptions(!options)}ƒ
          ></Button>
          {getActionsMenu()}
        </div>
      </ControlledComponent>

      <div className={'relative'}>
          <Button
            size="large"
            type="text"
            icon={<QuestionCircleOutlined />}
            onClick={()=>setHelpMenu(!helpMenu)}
          ></Button>
          {getHelpMenu()}
        </div>
    </div>
  );
};
const QueryActionsMemoized = memo(QueryActionsComponent);
const QueryActions = (props) => {
  const {
    coreQueryState: { chartTypes }
  } = useContext(CoreQueryContext);
  return <QueryActionsMemoized chartTypes={chartTypes} {...props} />;
};
export default QueryActions;
QueryActions.propTypes = {
  queryType: PropTypes.string.isRequired,
  savedQueryId: PropTypes.string,
  breakdown: PropTypes.array,
  attributionModels: PropTypes.array,
  handleSaveClick: PropTypes.func,
  handleEditClick: PropTypes.func,
  handleDeleteClick: PropTypes.func,
  toggleAddToDashboardModal: PropTypes.func,
  setShowShareToEmailModal: PropTypes.func,
  setShowShareToSlackModal: PropTypes.func
};
QueryActions.defaultProps = {
  savedQueryId: null,
  breakdown: [],
  attributionModels: [],
  handleSaveClick: _.noop,
  handleEditClick: _.noop,
  handleDeleteClick: _.noop,
  toggleAddToDashboardModal: _.noop,
  setShowShareToEmailModal: _.noop,
  setShowShareToSlackModal: _.noop
};
