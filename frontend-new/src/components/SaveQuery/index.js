import React, {
  useCallback,
  useContext,
  useEffect,
  useReducer,
  useState
} from 'react';
import { Button, Col, message, Modal, notification, Row } from 'antd';
import { saveQuery, updateQuery } from 'Reducers/coreQuery/services';
import { useSelector, useDispatch, connect } from 'react-redux';
import { isStringLengthValid } from 'Utils/global';
import { QUERY_CREATED, QUERY_UPDATED } from 'Reducers/types';
import { saveQueryToDashboard } from 'Reducers/dashboard/services';
import { fetchWeeklyIngishtsMetaData } from 'Reducers/insights';
import { QUERY_TYPE_ATTRIBUTION } from 'Utils/constants';
import { EMPTY_ARRAY } from 'Utils/global';
import { CoreQueryContext } from '../../contexts/CoreQueryContext';
import SaveQueryModal from './saveQueryModal';
import {
  ACTION_TYPES,
  SAVE_QUERY_INITIAL_STATE,
  TOGGLE_APIS_CALLED,
  TOGGLE_MODAL_VISIBILITY,
  SET_ACTIVE_ACTION,
  TOGGLE_ADD_TO_DASHBOARD_MODAL,
  TOGGLE_DELETE_MODAL
} from './saveQuery.constants';
import SaveQueryReducer from './saveQuery.reducer';

import factorsai from 'factorsai';
import QueryActions from './QueryActions';
import { getQuery } from './saveQuery.helpers';
import AddToDashboardModal from './AddToDashboardModal';
import { QUERY_DELETED } from '../../reducers/types';
import DeleteQueryModal from '../DeleteQueryModal';
import { getErrorMessage } from '../../utils/dataFormatter';
import { deleteReport } from '../../reducers/coreQuery/services';
import { getChartType } from '../../Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import { apiChartAnnotations } from '../../utils/constants';
import { isPivotSupported } from '../../utils/chart.helpers';
import ShareToEmailModal from '../ShareToEmailModal';
import ShareToSlackModal from '../ShareToSlackModal';
import {
  createAlert,
  sendAlertNow,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration
} from '../../reducers/global';
import AppModal from '../AppModal';
import { SVG, Text } from '../factorsComponents';
import { useHistory } from 'react-router-dom';
import _ from 'lodash';

function SaveQuery({
  requestQuery,
  queryType,
  setQuerySaved,
  getCurrentSorter,
  savedQueryId,
  queryTitle,
  breakdown,
  attributionsState,
  campaignState,
  fetchProjectSettingsV1,
  fetchSlackChannels,
  enableSlackIntegration,
  createAlert,
  sendAlertNow,
  dateFromTo,
  showSaveQueryModal,
  setShowSaveQueryModal,
  showUpdateQuery
}) {
  const dispatch = useDispatch();

  const history = useHistory();

  const [showShareToEmailModal, setShowShareToEmailModal] = useState(false);
  const [showShareToSlackModal, setShowShareToSlackModal] = useState(false);
  const [channelOpts, setChannelOpts] = useState([]);
  const [allChannels, setAllChannels] = useState([]);
  const [overrideDate, setOverrideDate] = useState(false);

  const { performance_criteria: user_type } = useSelector(
    (state) => state.analyticsQuery
  );

  const savedQueries = useSelector((state) =>
    _.get(state, 'queries.data', EMPTY_ARRAY)
  );
  const { active_project } = useSelector((state) => state.global);

  const { slack } = useSelector((state) => state.global);
  const { projectSettingsV1 } = useSelector((state) => state.global);
  const { agent_details } = useSelector((state) => state.agent);

  const {
    attributionMetrics,
    setNavigatedFromAnalyse,
    coreQueryState: { chartTypes, pivotConfig, navigatedFromDashboard, navigatedFromAnalyse }
  } = useContext(CoreQueryContext);

  const [saveQueryState, localDispatch] = useReducer(
    SaveQueryReducer,
    SAVE_QUERY_INITIAL_STATE
  );

  useEffect(() => {
    if (dateFromTo?.to === undefined || dateFromTo?.to === '') {
      setOverrideDate(false);
    } else {
      setOverrideDate(true);
    }
  }, [dateFromTo]);

  useEffect(() => {
    if (showSaveQueryModal) {
      handleSaveClick();
    }
  }, [showSaveQueryModal]);

  useEffect(() => {
    if (showUpdateQuery) {
      handleUpdateClick();
    }
  }, [showUpdateQuery]);

  const {
    activeAction,
    apisCalled,
    showSaveModal,
    showDeleteModal,
    showAddToDashModal
  } = saveQueryState;

  const updateLocalReducer = useCallback(({ type, payload }) => {
    localDispatch({ type, payload });
  }, []);

  const toggleModal = useCallback(() => {
    updateLocalReducer({ type: TOGGLE_MODAL_VISIBILITY });
  }, [updateLocalReducer]);

  const toggleAddToDashModal = useCallback(() => {
    updateLocalReducer({ type: TOGGLE_ADD_TO_DASHBOARD_MODAL });
  }, [updateLocalReducer]);

  const toggleDeleteModal = useCallback(() => {
    updateLocalReducer({ type: TOGGLE_DELETE_MODAL });
  }, [updateLocalReducer]);

  const handleSaveClick = useCallback(() => {
    toggleModal();
    updateLocalReducer({ type: SET_ACTIVE_ACTION, payload: ACTION_TYPES.SAVE });
    setShowSaveQueryModal(false);
  }, [updateLocalReducer, toggleModal]);

  const handleEditClick = useCallback(() => {
    toggleModal();
    updateLocalReducer({ type: SET_ACTIVE_ACTION, payload: ACTION_TYPES.EDIT });
  }, [updateLocalReducer, toggleModal]);

  const handleDelete = useCallback(async () => {
    try {
      updateLocalReducer({ type: TOGGLE_APIS_CALLED });
      await deleteReport({
        project_id: active_project.id,
        queryId: savedQueryId
      });
      updateLocalReducer({ type: TOGGLE_APIS_CALLED });
      toggleDeleteModal();
      setQuerySaved(null);
      dispatch({ type: QUERY_DELETED, payload: savedQueryId });
      notification.success({
        message: 'Report Deleted Successfully',
        duration: 5
      });
      history.push('/');
    } catch (err) {
      updateLocalReducer({ type: TOGGLE_APIS_CALLED });
      notification.error({
        message: 'Something went wrong!',
        description: getErrorMessage(err),
        duration: 5
      });
    }
  }, [dispatch, active_project, savedQueryId]);

  const handleAddToDashboard = useCallback(
    async ({ selectedDashboards, dashboardPresentation, onSuccess }) => {
      try {
        if (!selectedDashboards.length) {
          notification.error({
            message: 'Incorrect Input!',
            description: 'Please select atleast one dashboard',
            duration: 5
          });
          return false;
        }
        updateLocalReducer({ type: TOGGLE_APIS_CALLED });

        const queryGettingUpdated = savedQueries.find(
          (elem) => elem.id === savedQueryId
        );

        const querySettings = {
          ...queryGettingUpdated.settings,
          dashboardPresentation
        };

        const updateReqBody = {
          settings: querySettings,
          type: 1,
          title: queryTitle
        };

        await updateQuery(active_project.id, savedQueryId, updateReqBody);

        dispatch({
          type: QUERY_UPDATED,
          queryId: savedQueryId,
          payload: {
            title: queryTitle,
            settings: querySettings
          }
        });

        const reqBody = {
          query_id: savedQueryId
        };

        await saveQueryToDashboard(
          active_project.id,
          selectedDashboards.join(','),
          reqBody
        );

        dispatch({
          type: QUERY_UPDATED,
          queryId: savedQueryId,
          payload: {
            is_dashboard_query: true
          }
        });

        notification.success({
          message: 'Report added to dashboard Successfully',
          duration: 5
        });

        updateLocalReducer({ type: TOGGLE_APIS_CALLED });
        onSuccess();
      } catch (err) {
        updateLocalReducer({ type: TOGGLE_APIS_CALLED });
        console.log(err);
        console.log(err.response);
        notification.error({
          message: 'Error!',
          description: 'Something went wrong.',
          duration: 5
        });
      }
    },
    [savedQueryId, active_project, updateLocalReducer, queryTitle, savedQueries]
  );

  const handleSave = useCallback(
    async ({
      title,
      addToDashboard,
      selectedDashboards,
      dashboardPresentation,
      onSuccess
    }) => {
      try {
        if (!isStringLengthValid(title)) {
          notification.error({
            message: 'Incorrect Input!',
            description: 'Please Enter query title',
            duration: 5
          });
          return false;
        }

        if (addToDashboard && !selectedDashboards.length) {
          notification.error({
            message: 'Incorrect Input!',
            description: 'Please select atleast one dashboard',
            duration: 5
          });
          return false;
        }

        updateLocalReducer({ type: TOGGLE_APIS_CALLED });
        const query = getQuery({ queryType, requestQuery, user_type });

        const querySettings = {
          ...getCurrentSorter(),
          chart:
            apiChartAnnotations[
              getChartType({
                queryType,
                chartTypes,
                breakdown,
                attributionModels: attributionsState.models,
                campaignGroupBy: campaignState.group_by
              })
            ]
        };

        if (isPivotSupported({ queryType })) {
          querySettings.pivotConfig = JSON.stringify(pivotConfig);
        }

        if (queryType === QUERY_TYPE_ATTRIBUTION) {
          querySettings.attributionMetrics = JSON.stringify(attributionMetrics);
        }

        let queryId;
        let addedToDashboard = false;

        if (activeAction === ACTION_TYPES.SAVE) {
          const type = 2;
          if (addToDashboard) {
            querySettings.dashboardPresentation = dashboardPresentation;
          }
          const res = await saveQuery(
            active_project.id,
            title,
            query,
            type,
            querySettings
          );
          queryId = res.data.id;

          dispatch({ type: QUERY_CREATED, payload: res.data });
          setNavigatedFromAnalyse(res?.data);
          // setQuerySaved({ name: title, id: res.data.id });

          // if(queryType === QUERY_TYPE_EVENT && res?.data?.id_text) {
          //   history.replace('/analyse/event/' + res.data.id_text);
          // }

          // if(queryType === QUERY_TYPE_FUNNEL && res?.data?.id_text) {
          //   history.replace('/analyse/funnel/' + res.data.id_text);
          // }
        } else {
          const queryGettingUpdated = savedQueries.find(
            (elem) => elem.id === savedQueryId
          );

          const updatedSettings = {
            ...queryGettingUpdated.settings,
            ...querySettings
          };

          if (addToDashboard) {
            updatedSettings.dashboardPresentation = dashboardPresentation;
          }

          const reqBody = {
            title,
            settings: updatedSettings
          };

          await updateQuery(active_project.id, savedQueryId, reqBody);

          dispatch({
            type: QUERY_UPDATED,
            queryId: savedQueryId,
            payload: {
              title,
              settings: updatedSettings
            }
          });
          // setQuerySaved({ name: title, id: savedQueryId });
          queryId = savedQueryId;
        }

        if (addToDashboard) {
          try {
            const reqBody = {
              query_id: queryId
            };

            await saveQueryToDashboard(
              active_project.id,
              selectedDashboards.join(','),
              reqBody
            );
            addedToDashboard = true;
            dispatch({
              type: QUERY_UPDATED,
              queryId,
              payload: {
                is_dashboard_query: true
              }
            });
          } catch (error) {
            console.error('Error in adding to dashboard', error);
          }
        }

        setQuerySaved({ name: title, id: queryId });
        // Factors SAVE_QUERY EDIT_QUERY tracking
        factorsai.track(activeAction, {
          query_type: queryType,
          saved_query_id: savedQueryId,
          query_title: title,
          project_id: active_project.id,
          project_name: active_project.name
        });

        if (!addToDashboard) {
          notification.success({
            message: 'Report Saved Successfully',
            duration: 5
          });
        } else if (addedToDashboard) {
          notification.success({
            message: 'Saved and added to dashboard',
            duration: 5
          });
        } else {
          notification.warning({
            message:
              'Report saved, but couldn’t add it to a dashboard. Try again?',
            duration: 5
          });
        }
        updateLocalReducer({ type: TOGGLE_APIS_CALLED });
        dispatch(fetchWeeklyIngishtsMetaData(active_project.id));
        onSuccess();
      } catch (err) {
        updateLocalReducer({ type: TOGGLE_APIS_CALLED });
        console.log(err);
        console.log(err.response);
        notification.error({
          message: 'Error!',
          description: `${err?.data?.error}`,
          duration: 5
        });
      }
    },
    [
      active_project.id,
      requestQuery,
      dispatch,
      queryType,
      setQuerySaved,
      attributionMetrics,
      getCurrentSorter,
      savedQueryId,
      updateLocalReducer,
      activeAction,
      chartTypes,
      breakdown,
      attributionsState,
      campaignState,
      user_type
    ]
  );

  const handleUpdateClick = useCallback(async () => {
    try {
      let navigatedData;
      if(navigatedFromDashboard) {
        navigatedData = navigatedFromDashboard;
      }
      if(navigatedFromAnalyse){
        navigatedData = navigatedFromAnalyse;
      }
      const query = getQuery({ queryType, requestQuery, user_type });

      const querySettings = {
        ...getCurrentSorter(),
        chart:
          apiChartAnnotations[
            getChartType({
              queryType,
              chartTypes,
              breakdown,
              attributionModels: attributionsState.models,
              campaignGroupBy: campaignState.group_by
            })
          ]
      };

      if (isPivotSupported({ queryType })) {
        querySettings.pivotConfig = JSON.stringify(pivotConfig);
      }

      if (queryType === QUERY_TYPE_ATTRIBUTION) {
        querySettings.attributionMetrics = JSON.stringify(attributionMetrics);
      }

      const queryGettingUpdated = savedQueries.find(
        (elem) => elem.id === (navigatedData?.query_id || navigatedData?.key || navigatedData?.id)
      );

      const updatedSettings = {
        ...queryGettingUpdated.settings,
        ...querySettings
      };

      const reqBody = {
        title: (queryGettingUpdated?.query?.title || queryGettingUpdated?.title),
        query: query,
        settings: updatedSettings
      };

      await updateQuery(active_project.id, (navigatedData?.query_id || navigatedData?.key || navigatedData?.id), reqBody);

      dispatch({
        type: QUERY_UPDATED,
        queryId: (navigatedData?.query_id || navigatedData?.key || navigatedData?.id),
        payload: {
          title: (queryGettingUpdated?.query?.title || queryGettingUpdated?.title),
          query,
          settings: updatedSettings
        }
      });
      setQuerySaved({ name: (queryGettingUpdated?.query?.title || queryGettingUpdated?.title), id: (navigatedData?.query_id || navigatedData?.key || navigatedData?.id) });

      notification.success({
        message: 'Report Saved Successfully',
        duration: 5
      });
      dispatch(fetchWeeklyIngishtsMetaData(active_project.id));
    } catch (err) {
      console.log(err);
      console.log(err.response);
      notification.error({
        message: 'Error!',
        description: `${err?.data?.error}`,
        duration: 5
      });
    }
  }, [
    active_project.id,
    requestQuery,
    dispatch,
    queryType,
    setQuerySaved,
    attributionMetrics,
    getCurrentSorter,
    navigatedFromDashboard?.id,
    navigatedFromAnalyse?.key,
    activeAction,
    chartTypes,
    breakdown,
    attributionsState,
    campaignState,
    user_type
  ]);

  const onConnectSlack = () => {
    enableSlackIntegration(active_project.id)
      .then((r) => {
        if (r.status == 200) {
          window.open(r.data.redirectURL, '_blank');
          setShowShareToSlackModal(false);
        }
        if (r.status >= 400) {
          message.error('Error fetching slack redirect url');
        }
      })
      .catch((err) => {
        console.log('Slack error-->', err);
      });
  };

  useEffect(() => {
    fetchProjectSettingsV1(active_project.id);
    if (projectSettingsV1?.int_slack) {
      fetchSlackChannels(active_project.id);
    }
  }, [active_project, projectSettingsV1?.int_slack, showShareToSlackModal]);

  useEffect(() => {
    if (slack?.length > 0) {
      let tempArr = [];
      let allArr = [];
      for (let i = 0; i < slack.length; i++) {
        tempArr.push({ label: '#' + slack[i].name, value: slack[i].id });
        allArr.push({
          name: slack[i].name,
          id: slack[i].id,
          is_private: slack[i].is_private
        });
      }
      setChannelOpts(tempArr);
      setAllChannels(allArr);
    }
  }, [active_project, agent_details, slack]);

  const handleEmailClick = ({ data, frequency, onSuccess }) => {
    updateLocalReducer({ type: TOGGLE_APIS_CALLED });

    let queryData = undefined;
    if (savedQueryId) {
      queryData = savedQueries.find((elem) => elem.id === savedQueryId);
    }

    let emails = [];
    if (data?.emails) {
      emails = data.emails.map((item) => {
        return item.email;
      });
    }
    if (data.email) {
      emails.push(data.email);
    }

    let payload = {
      alert_name: queryData?.title || data?.subject,
      alert_type: 3,
      // "query_id": savedQueryId,
      alert_description: {
        message: data?.message,
        date_range: frequency == 'send_now' ? '' : frequency,
        subject: data?.subject
      },
      alert_configuration: {
        email_enabled: true,
        slack_enabled: false,
        emails: emails,
        slack_channels_and_user_groups: {}
      }
    };

    if (frequency === 'send_now') {
      sendAlertNow(
        active_project.id,
        payload,
        savedQueryId,
        dateFromTo,
        overrideDate
      )
        .then((r) => {
          notification.success({
            message: 'Report Sent Successfully',
            description: 'Report has been sent to the selected emails',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    } else {
      createAlert(active_project.id, payload, savedQueryId)
        .then((r) => {
          notification.success({
            message: 'Report Saved Successfully',
            description: 'Report will be sent on the specified date.',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    }
    updateLocalReducer({ type: TOGGLE_APIS_CALLED });
    onSuccess();
  };

  const handleSlackClick = ({ data, frequency, onSuccess }) => {
    updateLocalReducer({ type: TOGGLE_APIS_CALLED });

    let queryData = undefined;
    if (savedQueryId) {
      queryData = savedQueries.find((elem) => elem.id === savedQueryId);
    }

    let slackChannels = {};
    const selected = allChannels.filter((c) => c.id === data.channel);
    const map = new Map();
    map.set(agent_details.uuid, selected);
    for (const [key, value] of map) {
      slackChannels = { ...slackChannels, [key]: value };
    }

    let payload = {
      alert_name: queryData?.title || data?.subject,
      alert_type: 3,
      // "query_id": savedQueryId,
      alert_description: {
        message: data?.message,
        date_range: frequency == 'send_now' ? '' : frequency,
        subject: data?.subject
      },
      alert_configuration: {
        email_enabled: false,
        slack_enabled: true,
        emails: [],
        slack_channels_and_user_groups: slackChannels
      }
    };

    if (frequency === 'send_now') {
      sendAlertNow(
        active_project.id,
        payload,
        savedQueryId,
        dateFromTo,
        overrideDate
      )
        .then((r) => {
          notification.success({
            message: 'Report Sent Successfully',
            description: 'Report has been sent to the selected Slack channels',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    } else {
      createAlert(active_project.id, payload, savedQueryId)
        .then((r) => {
          notification.success({
            message: 'Report Saved Successfully',
            description: 'Report will be sent on the specified date.',
            duration: 5
          });
        })
        .catch((err) => {
          message.error(err?.data?.error);
        });
    }
    updateLocalReducer({ type: TOGGLE_APIS_CALLED });
    onSuccess();
  };

  return (
    <>
      <QueryActions
        queryType={queryType}
        chartTypes={chartTypes}
        breakdown={breakdown}
        savedQueryId={savedQueryId}
        handleSaveClick={handleSaveClick}
        handleEditClick={handleEditClick}
        handleUpdateClick={handleUpdateClick}
        handleDeleteClick={toggleDeleteModal}
        toggleAddToDashboardModal={toggleAddToDashModal}
        setShowShareToEmailModal={setShowShareToEmailModal}
        setShowShareToSlackModal={setShowShareToSlackModal}
        attributionModels={attributionsState.models}
      />

      <SaveQueryModal
        visible={showSaveModal}
        isLoading={apisCalled}
        modalTitle={
          activeAction === ACTION_TYPES.SAVE
            ? 'Create New Report'
            : 'Edit Report Details'
        }
        queryType={queryType}
        requestQuery={requestQuery}
        onSubmit={handleSave}
        toggleModalVisibility={toggleModal}
        activeAction={activeAction}
        queryTitle={queryTitle}
      />

      <AddToDashboardModal
        toggleModalVisibility={toggleAddToDashModal}
        visible={showAddToDashModal}
        isLoading={apisCalled}
        onSubmit={handleAddToDashboard}
      />

      <DeleteQueryModal
        visible={showDeleteModal}
        onDelete={handleDelete}
        toggleModal={toggleDeleteModal}
        isLoading={apisCalled}
      />

      <ShareToEmailModal
        visible={showShareToEmailModal}
        onSubmit={handleEmailClick}
        isLoading={apisCalled}
        setShowShareToEmailModal={setShowShareToEmailModal}
        queryTitle={queryTitle}
      />

      {projectSettingsV1?.int_slack ? (
        <ShareToSlackModal
          visible={showShareToSlackModal}
          onSubmit={handleSlackClick}
          channelOpts={channelOpts}
          isLoading={apisCalled}
          setShowShareToSlackModal={setShowShareToSlackModal}
          queryTitle={queryTitle}
        />
      ) : (
        <AppModal
          title={null}
          visible={showShareToSlackModal}
          footer={null}
          centered={true}
          mask={true}
          maskClosable={false}
          maskStyle={{ backgroundColor: 'rgb(0 0 0 / 70%)' }}
          closable={true}
          isLoading={apisCalled}
          onCancel={() => setShowShareToSlackModal(false)}
          className={`fa-modal--regular`}
          width={'470px'}
        >
          <div className={'m-0 mb-2'}>
            <Row className={'m-0'}>
              <Col>
                <SVG
                  name={'Slack'}
                  size={25}
                  extraClass={'inline mr-2 -mt-2'}
                />
                <Text
                  type={'title'}
                  level={5}
                  weight={'bold'}
                  extraClass={'inline m-0'}
                >
                  Slack Integration
                </Text>
              </Col>
            </Row>
            <Row className={'m-0 mt-4'}>
              <Col>
                <Text
                  type={'title'}
                  level={6}
                  color={'grey-2'}
                  weight={'regular'}
                  extraClass={'m-0'}
                >
                  Slack is not integrated, Do you want to integrate with your
                  slack account now?
                </Text>
              </Col>
            </Row>
            <Col>
              <Row justify='end' className={'w-full mb-1 mt-4'}>
                <Col className={'mr-2'}>
                  <Button
                    type={'default'}
                    onClick={() => setShowShareToSlackModal(false)}
                  >
                    Cancel
                  </Button>
                </Col>
                <Col className={'mr-2'}>
                  <Button type={'primary'} onClick={onConnectSlack}>
                    Connect to slack
                  </Button>
                </Col>
              </Row>
            </Col>
          </div>
        </AppModal>
      )}
    </>
  );
}

export default connect(null, {
  createAlert,
  sendAlertNow,
  fetchSlackChannels,
  fetchProjectSettingsV1,
  enableSlackIntegration
})(SaveQuery);
