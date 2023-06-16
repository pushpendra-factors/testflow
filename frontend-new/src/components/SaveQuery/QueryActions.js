import React, { memo, useContext, useEffect, useState } from 'react';
import _ from 'lodash';
import PropTypes from 'prop-types';
import { SVG } from 'factorsComponents';
import { Button, Dropdown, Menu, message, Tooltip } from 'antd';
import { BUTTON_TYPES } from '../../constants/buttons.constants';
import ControlledComponent from '../ControlledComponent';
import styles from './index.module.scss';
import { QuestionCircleOutlined } from '@ant-design/icons';
import {
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_FUNNEL
} from '../../utils/constants';
import { getChartType } from '../../Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import { CoreQueryContext } from '../../contexts/CoreQueryContext';
import userflow from 'userflow.js';
import { USERFLOW_CONFIG_ID } from 'Utils/userflowConfig';

const QueryActionsComponent = ({
  queryType,
  breakdown,
  chartTypes,
  savedQueryId,
  attributionModels,
  handleSaveClick,
  handleEditClick,
  handleDeleteClick,
  handleUpdateClick,
  toggleAddToDashboardModal,
  setShowShareToEmailModal,
  setShowShareToSlackModal
}) => {
  const [hideIntercomState, setHideIntercomState] = useState(true);

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

  const {
    coreQueryState: { navigatedFromDashboard, navigatedFromAnalyse }
  } = useContext(CoreQueryContext);

  const handleIntercomHelp = () => {
    const w = window;
    const ic = w.Intercom;
    if (typeof ic === 'function') {
      setHideIntercomState(!hideIntercomState);
      ic('update', { hide_default_launcher: !hideIntercomState });
      ic(!hideIntercomState === true ? 'hide' : 'show');
    }
  };

  const [chart, setChart] = useState(null);
  useEffect(() => {
    setChart(
      getChartType({ queryType, chartTypes, breakdown, attributionModels })
    );
  }, [queryType, chartTypes, breakdown, attributionModels]);

  const triggerUserFlow = () => {
    if (
      queryType === QUERY_TYPE_ATTRIBUTION ||
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_KPI
    ) {
      let flowID = '';
      if (queryType === QUERY_TYPE_ATTRIBUTION) {
        flowID = USERFLOW_CONFIG_ID?.AttributionQueryBuilder;
      }
      if (queryType === QUERY_TYPE_FUNNEL) {
        flowID = USERFLOW_CONFIG_ID?.FunnelSQueryBuilder;
      }
      if (queryType === QUERY_TYPE_KPI) {
        flowID = USERFLOW_CONFIG_ID?.KPIQueryBuilder;
      }

      userflow.start(flowID);
    }
  };

  const handleMenuClick = (e) => {
    if (e?.key === '1') {
      handleSaveClick();
    } else {
      handleUpdateClick();
    }
  };

  const handleActionMenuClick = (e) => {
    if (e?.key === '1') {
      setShowShareToEmailModal(true);
    } else if (e?.key === '2') {
      setShowShareToSlackModal(true);
    } else if (e?.key === '3') {
      toggleAddToDashboardModal();
    } else if (e?.key === '4') {
      handleEditClick();
    } else if (e?.key === '5') {
      handleDeleteClick();
    } else if (e?.key === '6') {
      handleIntercomHelp();
    } else if (e?.key === '7') {
      window.open('https://help.factors.ai/', '_blank');
    }
  };

  const menuItems = (
    <Menu onClick={handleMenuClick} className={`${styles.antdActionMenu}`}>
      <Menu.Item key='1'>
        <SVG
          name={'pluscopy'}
          size={20}
          color={'grey'}
          extraClass={'inline -mt-1 mr-1'}
        />
        Save as New
      </Menu.Item>
      <Menu.Item key='2'>
        <SVG
          name={'save'}
          size={20}
          color={'grey'}
          extraClass={'inline -mt-1 mr-1'}
        />
        Save
      </Menu.Item>
    </Menu>
  );

  const actionMenu = (
    <Menu
      onClick={handleActionMenuClick}
      className={`${styles.antdActionMenu}`}
    >
      <Menu.Item key='1' disabled={!savedQueryId}>
        <SVG
          name={'envelope'}
          size={18}
          color={`${!savedQueryId ? 'LightGray' : 'grey'}`}
          extraClass={'inline mr-2'}
        />
        Email this report
      </Menu.Item>
      <Menu.Item key='2' disabled={!savedQueryId}>
        <SVG
          name={'SlackStroke'}
          size={18}
          color={`${!savedQueryId ? 'LightGray' : 'grey'}`}
          extraClass={'inline mr-2'}
        />
        Share to slack
      </Menu.Item>
      <Menu.Item key='3' disabled={!savedQueryId}>
        <SVG
          name={'addtodash'}
          size={18}
          color={`${!savedQueryId ? 'LightGray' : 'grey'}`}
          extraClass={'inline mr-2'}
        />
        Add to Dashboard
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item key='4' disabled={!savedQueryId}>
        <SVG
          name={'edit'}
          size={18}
          color={`${!savedQueryId ? 'LightGray' : 'grey'}`}
          extraClass={'inline mr-2'}
        />
        Edit Details
      </Menu.Item>
      <Menu.Item key='5' disabled={!savedQueryId}>
        <SVG
          name={'trash'}
          size={18}
          color={`${!savedQueryId ? 'LightGray' : 'grey'}`}
          extraClass={'inline mr-2'}
        />
        Delete
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item key='6'>
        <SVG
          name={'headset'}
          size={18}
          color={'grey'}
          extraClass={'inline mr-2'}
        />
        Talk to us
      </Menu.Item>
      <Menu.Item key='7'>
        <QuestionCircleOutlined
          style={{ fontSize: '15px', marginRight: '12px' }}
        />
        Help and Support
      </Menu.Item>
    </Menu>
  );

  return (
    <div className='flex gap-x-2 items-center'>
      {(queryType == QUERY_TYPE_ATTRIBUTION ||
        queryType == QUERY_TYPE_FUNNEL ||
        queryType == QUERY_TYPE_KPI) && (
        <>
          <Tooltip placement='bottom' title='Walk me through'>
            <Button
              onClick={triggerUserFlow}
              size='large'
              type='text'
              icon={<SVG name={'Handshake'} size={24} color={'grey'} />}
            />
          </Tooltip>
        </>
      )}

      <ControlledComponent controller={!!savedQueryId}>
        <ControlledComponent controller={queryType !== QUERY_TYPE_PROFILE}>
          <Tooltip placement='bottom' title='Add to Dashboard'>
            <Button
              onClick={toggleAddToDashboardModal}
              size='large'
              type='text'
              icon={<SVG name={'addtodash'} size={20} />}
            ></Button>
          </Tooltip>
        </ControlledComponent>
      </ControlledComponent>

      <div className={'relative gap-x-2 mr-2'}>
        <Dropdown overlay={actionMenu} placement='bottomRight'>
          <Button type='text' icon={<SVG name={'threedot'} size={25} />} />
        </Dropdown>
      </div>

      {!(
        navigatedFromDashboard?.id ||
        navigatedFromAnalyse?.key ||
        navigatedFromAnalyse?.id
      ) &&
        !savedQueryId && (
          <Button
            onClick={handleSaveClick}
            size='large'
            type={BUTTON_TYPES.PRIMARY}
          >
            Save
          </Button>
        )}

      {savedQueryId && (
        <Tooltip placement='bottom' title={'No changes to be saved'}>
          <div className={`${styles.antdIcon}`}>
            <Dropdown.Button
              overlay={menuItems}
              disabled={savedQueryId}
              type={BUTTON_TYPES.PRIMARY}
              size={'large'}
              icon={<SVG name={'CaretDown'} size={20} color={'LightGray'} />}
            >
              Save
            </Dropdown.Button>
          </div>
        </Tooltip>
      )}
      {!savedQueryId &&
        (navigatedFromDashboard?.id ||
          navigatedFromAnalyse?.key ||
          navigatedFromAnalyse?.id) && (
          <div className={`${styles.antdIcon}`}>
            <Dropdown.Button
              overlay={menuItems}
              onClick={handleUpdateClick}
              type={BUTTON_TYPES.PRIMARY}
              size={'large'}
              icon={<SVG name={'CaretDown'} size={20} color={'white'} />}
            >
              Save
            </Dropdown.Button>
          </div>
        )}
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
