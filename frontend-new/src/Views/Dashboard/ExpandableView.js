import React, { useEffect, useCallback, useState } from "react";
import moment from "moment";
import { Modal, Spin } from "antd";
import styles from "./index.module.scss";
import ActiveUnitContent from "./ActiveUnitContent";
import {
  initialState,
  formatApiData,
  DefaultDateRangeFormat,
} from "../CoreQuery/utils";
import { useSelector } from "react-redux";
import { getDataFromServer } from "./utils";
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN
} from "../../utils/constants";

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
          if (newDurationObj.frequency === "hour") {
            refresh = true;
          }
        } else if (unit.query.query.cl && unit.query.query.cl === QUERY_TYPE_ATTRIBUTION) {
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
          });
        } else if (queryType === QUERY_TYPE_ATTRIBUTION) {
          setResultState({
            ...initialState,
            data: res.data,
          });
        } else {
          if (refresh) {
            setResultState({
              ...initialState,
              data: formatApiData(
                res.data.result_group[0],
                res.data.result_group[1]
              ),
            });
          } else {
            setResultState({
              ...initialState,
              data: formatApiData(
                res.data.result.result_group[0],
                res.data.result.result_group[1]
              ),
            });
          }
        }
      } catch (err) {
        console.log(err);
        console.log(err.response);
        setResultState({
          ...initialState,
          error: true,
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
      if (dates && dates.selected) {
        let frequency = "date";
        if (
          moment(dates.selected.endDate).diff(
            dates.selected.startDate,
            "hours"
          ) <= 24
        ) {
          frequency = "hour";
        }
        const newDurationObj = {
          ...duration,
          from: dates.selected.startDate,
          to: dates.selected.endDate,
          frequency,
        };
        setDuration(newDurationObj);
        getData(newDurationObj);
      }
    },
    [duration, getData]
  );

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
