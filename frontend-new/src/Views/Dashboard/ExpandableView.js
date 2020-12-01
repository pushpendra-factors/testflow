import React, { useEffect, useCallback, useState } from 'react';
import { Modal, Spin } from 'antd';
import styles from './index.module.scss';
import ActiveUnitContent from './ActiveUnitContent';
import { initialState, formatApiData } from '../CoreQuery/utils';
import { useSelector } from 'react-redux';
import { getDataFromServer } from './utils';

function ExpandableView({ widgetModal, setwidgetModal, widgetModalLoading, durationObj }) {

  const [duration, setDuration] = useState({
    from: '',
    to: '',
    frequency: ''
  });

  const { active_project } = useSelector(state => state.global);
  const [resultState, setResultState] = useState(initialState);
  const [unit, setUnit] = useState(null);

  const getData = useCallback(async (newDurationObj) => {
    try {
      setResultState({
        ...initialState,
        loading: true
      });

      const res = await getDataFromServer(unit.query, unit.id, unit.dashboard_id, newDurationObj, false, active_project.id);
      let queryType;

      if (unit.query.query.query_group) {
        queryType = 'event';
      } else {
        queryType = 'funnel';
      }

      if (queryType === 'funnel') {
        setResultState({
          ...initialState,
          data: res.data.result
        });
      } else {
        setResultState({
          ...initialState,
          data: formatApiData(res.data.result.result_group[0], res.data.result.result_group[1])
        });
      }
    } catch (err) {
      console.log(err);
      console.log(err.response);
      setResultState({
        ...initialState,
        error: true
      });
    }
  }, [active_project.id, unit]);

  useEffect(() => {
    if (widgetModal && widgetModal.data) {
      setDuration({ ...durationObj });
      setResultState({
        ...initialState,
        data: widgetModal.data
      });
      setUnit({ ...widgetModal.unit });
    }
  }, [widgetModal, durationObj]);

  const handleDurationChange = useCallback((dates) => {
    if (dates && dates.selected) {
      const newDurationObj = {
        ...duration,
        from: dates.selected.startDate,
        to: dates.selected.endDate
      }
      setDuration(newDurationObj);
      getData(newDurationObj);
    }
  }, [duration, getData]);

  let content = null;

  if (widgetModalLoading) {
    content = (
      <div className="flex justify-center items-center w-full min-h-screen">
        <Spin size="small" />
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
      mask={false}
      closable={false}
      onCancel={() => setwidgetModal(false)}
      className={`w-full inset-0 ${styles.fullModal}`}
    >
      {content}
    </Modal>
  );
}

export default ExpandableView;
