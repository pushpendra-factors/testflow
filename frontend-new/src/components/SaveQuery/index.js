import React, { useState, useCallback } from "react";
import moment from "moment";
import {
  Button,
  Modal,
  Input,
  Switch,
  Select,
  Radio,
  notification,
} from "antd";
import { SVG, Text } from "../factorsComponents";
import styles from "./index.module.scss";
import { saveQuery } from "../../reducers/coreQuery/services";
import { useSelector, useDispatch } from "react-redux";
import { QUERY_CREATED } from "../../reducers/types";
import { saveQueryToDashboard } from "../../reducers/dashboard/services";
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
} from "../../utils/constants";
import { getSessionsQuery, getFrequencyQuery, getTotalEventsQuery, getTotalUsersQuery } from "../../Views/CoreQuery/utils";

function SaveQuery({
  requestQuery,
  setQuerySaved,
  visible,
  setVisible,
  activeKey,
  breakdownType,
  queryType,
}) {
  const [title, setTitle] = useState("");
  const [addToDashboard, setAddToDashboard] = useState(false);
  const [selectedDashboards, setSelectedDashboards] = useState([]);
  const [dashboardPresentation, setDashboardPresentation] = useState("pt");
  const [apisCalled, setApisCalled] = useState(false);
  const { active_project } = useSelector((state) => state.global);
  const { dashboards } = useSelector((state) => state.dashboard);
  const dispatch = useDispatch();

  const handleTitleChange = useCallback((e) => {
    setTitle(e.target.value);
  }, []);

  const resetModalState = useCallback(() => {
    setTitle("");
    setSelectedDashboards([]);
    setAddToDashboard(false);
    setDashboardPresentation("pb");
    setVisible(false);
  }, [setVisible]);

  const handleSaveCancel = useCallback(() => {
    if (!apisCalled) {
      resetModalState();
    }
  }, [resetModalState, apisCalled]);

  const handleSelectChange = useCallback(
    (value) => {
      const resp = value.map((v) => {
        return dashboards.data.find((d) => d.name === v).id;
      });
      setSelectedDashboards(resp);
    },
    [dashboards.data]
  );

  const handlePresentationChange = useCallback((e) => {
    setDashboardPresentation(e.target.value);
  }, []);

  const toggleAddToDashboard = useCallback(
    (val) => {
      setAddToDashboard(val);
    },
    [setAddToDashboard]
  );

  const getSelectedDashboards = useCallback(() => {
    return selectedDashboards.map((s) => {
      return dashboards.data.find((d) => d.id === s).name;
    });
  }, [dashboards.data, selectedDashboards]);

  const handleSave = useCallback(async () => {
    if (!title.trim().length) {
      notification.error({
        message: "Incorrect Input!",
        description: "Please Enter query title",
        duration: 5,
      });
      return false;
    }
    if (addToDashboard && !selectedDashboards.length) {
      notification.error({
        message: "Incorrect Input!",
        description: "Please select atleast one dashboard",
        duration: 5,
      });
      return false;
    }

    try {
      setApisCalled(true);
      let query;
      if (queryType === QUERY_TYPE_FUNNEL) {
        query = {
          ...requestQuery,
          fr: moment().startOf("week").utc().unix(),
          to: moment().utc().unix(),
        };
      } else if (queryType === QUERY_TYPE_ATTRIBUTION) {
        query = {
          ...requestQuery,
          query: {
            ...requestQuery.query,
            from: moment().startOf("week").utc().unix(),
            to: moment().utc().unix(),
          },
        };
      } else if (queryType === QUERY_TYPE_EVENT) {
        query = {
          query_group: requestQuery.map((q) => {
            return {
              ...q,
              fr: moment().startOf("week").utc().unix(),
              to: moment().utc().unix(),
              gbt: q.gbt ? "date" : "",
            };
          }),
        };
        if(parseInt(activeKey) === 0) {
          query.query_group = getTotalEventsQuery(query);
        }
        if(parseInt(activeKey) === 1) {
          query.query_group = getTotalUsersQuery(query);
        }
        if (parseInt(activeKey) === 2) {
          query.query_group = getSessionsQuery(query);
        }
        if (parseInt(activeKey) === 3) {
          query.query_group = getFrequencyQuery(query);
        }
      } else if (queryType === QUERY_TYPE_CAMPAIGN) {
        query = {
          ...requestQuery,
          query_group: requestQuery.query_group.map((q) => {
            return {
              ...q,
              fr: moment().startOf("week").utc().unix(),
              to: moment().utc().unix(),
              gbt: q.gbt ? "date" : "",
            };
          }),
        };
      }
      const type = addToDashboard ? 1 : 2;
      const res = await saveQuery(active_project.id, title, query, type);
      if (addToDashboard) {
        const settings = {
          chart: dashboardPresentation,
        };
        if (activeKey) {
          settings.activeKey = activeKey;
        }
        if (breakdownType !== "each") {
          settings.breakdownType = breakdownType;
        }
        const reqBody = {
          settings,
          description: "",
          title,
          query_id: res.data.id,
        };
        await saveQueryToDashboard(
          active_project.id,
          selectedDashboards.join(","),
          reqBody
        );
      }
      dispatch({ type: QUERY_CREATED, payload: res.data });
      setQuerySaved(title);
      setApisCalled(false);
      resetModalState();
    } catch (err) {
      setApisCalled(false);
      console.log(err);
      console.log(err.response);
      notification.error({
        message: "Error!",
        description: "Something went wrong.",
        duration: 5,
      });
    }
  }, [
    activeKey,
    breakdownType,
    title,
    active_project.id,
    requestQuery,
    dispatch,
    setQuerySaved,
    resetModalState,
    addToDashboard,
    dashboardPresentation,
    selectedDashboards,
    queryType,
  ]);

  let dashboardHelpText = "Create a dashboard widget for regular monitoring";
  let dashboardOptions = null;

  if (addToDashboard) {
    dashboardHelpText = "This widget will appear on the following dashboards:";

    let firstOption = <Radio value="pb">Display Bar Chart</Radio>;
    let secondOption = null;

    if (queryType === QUERY_TYPE_EVENT) {
      secondOption = <Radio value="pl">Display Line Chart</Radio>;
      if (!requestQuery[0].gbp.length) {
        firstOption = <Radio value="pc">Display Spark Line Chart</Radio>;
      }
    }

    if (queryType === QUERY_TYPE_CAMPAIGN) {
      secondOption = <Radio value="pl">Display Line Chart</Radio>;
      if (!requestQuery.query_group[0].group_by.length) {
        firstOption = <Radio value="pc">Display Spark Line Chart</Radio>;
      }
    }

    dashboardOptions = (
      <>
        <div className="mt-5">
          <Select
            mode="multiple"
            style={{ width: "100%" }}
            placeholder={"Please Select"}
            onChange={handleSelectChange}
            className={styles.selectStyles}
            value={getSelectedDashboards()}
          >
            {dashboards.data.map((d) => {
              return (
                <Select.Option value={d.name} key={d.id}>
                  {d.name}
                </Select.Option>
              );
            })}
          </Select>
        </div>
        <div className="mt-6">
          <Radio.Group
            value={dashboardPresentation}
            onChange={handlePresentationChange}
          >
            {firstOption}
            {secondOption}
            <Radio value="pt">Display Table</Radio>
          </Radio.Group>
        </div>
      </>
    );
  }

  return (
    <>
      <Button
        onClick={setVisible.bind(this, true)}
        style={{ display: "flex" }}
        placeholder={"Select Options"}
        className="items-center"
        type="primary"
        icon={<SVG extraClass="mr-1" name={"save"} size={24} color="#FFFFFF" />}
      >
        Save
      </Button>

      <Modal
        centered={true}
        visible={visible}
        width={700}
        title={null}
        onOk={handleSave}
        onCancel={handleSaveCancel}
        className={"fa-modal--regular p-4 fa-modal--slideInDown"}
        okText={"Save"}
        closable={false}
        confirmLoading={apisCalled}
        transitionName=""
        maskTransitionName=""
      >
        <div className="p-4">
          <Text extraClass="m-0" type={"title"} level={3} weight={"bold"}>
            Save this Query
          </Text>
          <div className="pt-6">
            <Text
              type={"title"}
              level={7}
              extraClass={`m-0 ${styles.inputLabel}`}
            >
              Title
            </Text>
            <Input
              onChange={handleTitleChange}
              value={title}
              className={"fa-input"}
              size={"large"}
            />
          </div>
          {/* <div className={`pt-2 ${styles.linkText}`}>Help others to find this query easily?</div> */}
          <div className={"pt-6 flex items-center"}>
            <Switch
              onChange={toggleAddToDashboard}
              checked={addToDashboard}
              className={styles.switchBtn}
              checkedChildren="On"
              unCheckedChildren="Off"
            />
            <Text extraClass="m-0" type="title" level={6} weight="bold">
              Add to Dashboard
            </Text>
          </div>
          <Text extraClass={`pt-1 ${styles.noteText}`} mini type={"paragraph"}>
            {dashboardHelpText}
          </Text>
          {dashboardOptions}
        </div>
      </Modal>
    </>
  );
}

export default SaveQuery;
