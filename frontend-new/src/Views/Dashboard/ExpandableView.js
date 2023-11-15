import React, { useEffect, useCallback, useState } from 'react';
import MomentTz from 'Components/MomentTz';
import { Modal, Spin } from 'antd';
import ActiveUnitContent from './ActiveUnitContent';
import {
  initialState,
  formatApiData,
  DefaultDateRangeFormat,
} from '../CoreQuery/utils';
import { useSelector } from 'react-redux';
import { getDataFromServer } from './utils';
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
} from '../../utils/constants';

function ExpandableView({
  widgetModal,
  setwidgetModal,
  widgetModalLoading,
  durationObj,
}) {
  const [duration, setDuration] = useState({ ...DefaultDateRangeFormat });

  const { active_project } = useSelector((state) => state.global);
  const [resultState, setResultState] = useState(initialState);
  const [unit, setUnit] = useState(null);

  const getData = useCallback(
    async (newDurationObj) => {
      try {
        setResultState({
          ...initialState,
          loading: true,
        });

        let queryType;
        let refresh = false;

        if (unit.query.query.query_group) {
          queryType = QUERY_TYPE_EVENT;
          if (
            unit.query.query.cl &&
            unit.query.query.cl === QUERY_TYPE_CAMPAIGN
          ) {
            queryType = QUERY_TYPE_CAMPAIGN;
          } else {
            queryType = QUERY_TYPE_EVENT;
          }
        } else if (
          unit.query.query.cl &&
          unit.query.query.cl === QUERY_TYPE_ATTRIBUTION
        ) {
          queryType = QUERY_TYPE_ATTRIBUTION;
        } else {
          queryType = QUERY_TYPE_FUNNEL;
        }

        const res = await getDataFromServer(
          unit.query,
          unit.id,
          unit.dashboard_id,
          newDurationObj,
          refresh,
          active_project.id
        );

        if (queryType === QUERY_TYPE_FUNNEL) {
          setResultState({
            ...initialState,
            data: res.data.result,
            status:res.status
          });
        } else if (queryType === QUERY_TYPE_ATTRIBUTION) {
          setResultState({
            ...initialState,
            data: res.data.result,
            status:res.status
          });
        } else if (queryType === QUERY_TYPE_CAMPAIGN) {
          setResultState({
            ...initialState,
            data: res.data.result,
            status:res.status
          });
        } else {
          setResultState({
            ...initialState,
            data: formatApiData(
              res.data.result.result_group[0],
              res.data.result.result_group[1]
            ),
            status:res.status
          });
        }
      } catch (err) {
        console.log(err);
        console.log(err.response);
        setResultState({
          ...initialState,
          error: true,
          status:err.status
        });
      }
    },
    [active_project.id, unit]
  );

  useEffect(() => {
    if (widgetModal && widgetModal.data) {
      setDuration({ ...durationObj });
      setResultState({
        ...initialState,
        data: widgetModal.data,
      });
      setUnit({ ...widgetModal.unit });
    }
  }, [widgetModal, durationObj]);

  const handleDurationChange = useCallback(
    (dates) => {
      let from,
        to,
        frequency = 'date';
      if (Array.isArray(dates.startDate)) {
        from = dates.startDate[0];
        to = dates.startDate[1];
      } else {
        from = dates.startDate;
        to = dates.endDate;
      }
      if (MomentTz(to).diff(from, 'hours') < 24) {
        frequency = 'hour';
      }
      const newDurationObj = {
        ...duration,
        from,
        to,
        frequency,
      };
      setDuration(newDurationObj);
      getData(newDurationObj);
    },
    [duration, getData]
  );

  let content = null;

  if (widgetModalLoading) {
    content = (
      <div className='flex justify-center items-center w-full min-h-screen'>
        <Spin />
      </div>
    );
  } else if (unit) {
    content = (
      <ActiveUnitContent
        unit={unit}
        resultState={resultState}
        setwidgetModal={setwidgetModal}
        durationObj={duration}
        handleDurationChange={handleDurationChange}
      />
    );
  }

  return (
    <Modal
      title={null}
      visible={widgetModal}
      footer={null}
      centered={false}
      zIndex={1015}
      mask={true}
      closable={false}
      onCancel={() => setwidgetModal(false)}
      // className={`w-full inset-0 ${styles.fullModal}`}
      className={`fa-modal--regular fa-modal--quick-view fa-modal--slideInDown`}
      transitionName=''
      maskTransitionName=''
    >
      {content}
    </Modal>
  );
}

export default ExpandableView;
