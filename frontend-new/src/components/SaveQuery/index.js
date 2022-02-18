import React, { useCallback, useContext, useReducer } from 'react';
import { notification } from 'antd';
import { saveQuery, updateQuery } from 'Reducers/coreQuery/services';
import { useSelector, useDispatch } from 'react-redux';
import { isStringLengthValid } from 'Utils/global';
import { QUERY_CREATED, QUERY_UPDATED } from 'Reducers/types';
import { saveQueryToDashboard } from 'Reducers/dashboard/services';
import { fetchWeeklyIngishtsMetaData } from 'Reducers/insights';
import { QUERY_TYPE_ATTRIBUTION } from 'Utils/constants';
import { CoreQueryContext } from '../../contexts/CoreQueryContext';
import SaveQueryModal from './SaveQueryModal';
import {
  ACTION_TYPES,
  SAVE_QUERY_INITIAL_STATE,
  TOGGLE_APIS_CALLED,
  TOGGLE_MODAL_VISIBILITY,
  SET_ACTIVE_ACTION,
  TOGGLE_ADD_TO_DASHBOARD_MODAL,
} from './saveQuery.constants';
import SaveQueryReducer from './saveQuery.reducer';

import factorsai from 'factorsai';
import QueryActions from './QueryActions';
import { deleteQuery } from '../../reducers/coreQuery/services';
import { getQuery } from './saveQuery.helpers';
import AddToDashboardModal from './AddToDashboardModal';

function SaveQuery({
  requestQuery,
  queryType,
  setQuerySaved,
  getCurrentSorter,
  savedQueryId,
  queryTitle,
}) {
  const dispatch = useDispatch();

  const { active_project } = useSelector((state) => state.global);

  const { attributionMetrics } = useContext(CoreQueryContext);

  const [saveQueryState, localDispatch] = useReducer(
    SaveQueryReducer,
    SAVE_QUERY_INITIAL_STATE
  );

  const {
    activeAction,
    apisCalled,
    showSaveModal,
    showAddToDashModal,
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

  const handleSaveClick = useCallback(() => {
    toggleModal();
    updateLocalReducer({ type: SET_ACTIVE_ACTION, payload: ACTION_TYPES.SAVE });
  }, [updateLocalReducer, toggleModal]);

  const handleEditClick = useCallback(() => {
    toggleModal();
    updateLocalReducer({ type: SET_ACTIVE_ACTION, payload: ACTION_TYPES.EDIT });
  }, [updateLocalReducer, toggleModal]);

  const handleDeleteReport = useCallback(() => {
    setQuerySaved(null);
    dispatch(deleteQuery({ project_id: active_project.id, id: savedQueryId }));
  }, [setQuerySaved, savedQueryId, active_project, dispatch]);

  const handleAddToDashboard = useCallback(
    async ({ selectedDashboards, dashboardPresentation, onSuccess }) => {
      try {
        if (!selectedDashboards.length) {
          notification.error({
            message: 'Incorrect Input!',
            description: 'Please select atleast one dashboard',
            duration: 5,
          });
          return false;
        }
        updateLocalReducer({ type: TOGGLE_APIS_CALLED });
        const querySettings = {
          chart: dashboardPresentation,
        };
        const updateReqBody = {
          settings: querySettings,
          type: 1,
          title: queryTitle,
        };
        await updateQuery(active_project.id, savedQueryId, updateReqBody);

        const reqBody = {
          query_id: savedQueryId,
        };

        await saveQueryToDashboard(
          active_project.id,
          selectedDashboards.join(','),
          reqBody
        );

        notification.success({
          message: 'Report added to dashboard Successfully',
          duration: 5,
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
          duration: 5,
        });
      }
    },
    [savedQueryId, active_project, updateLocalReducer, queryTitle]
  );

  const handleSave = useCallback(
    async ({ title, onSuccess }) => {
      try {
        if (!isStringLengthValid(title)) {
          notification.error({
            message: 'Incorrect Input!',
            description: 'Please Enter query title',
            duration: 5,
          });
          return false;
        }

        updateLocalReducer({ type: TOGGLE_APIS_CALLED });
        const query = getQuery({ queryType, requestQuery });

        const querySettings = {
          ...getCurrentSorter(),
        };

        if (queryType === QUERY_TYPE_ATTRIBUTION) {
          querySettings.attributionMetrics = JSON.stringify(attributionMetrics);
        }

        if (activeAction === ACTION_TYPES.SAVE) {
          const type = 2;
          const res = await saveQuery(
            active_project.id,
            title,
            query,
            type,
            querySettings
          );
          dispatch({ type: QUERY_CREATED, payload: res.data });
          setQuerySaved({ name: title, id: res.data.id });
        } else {
          const reqBody = {
            title,
            settings: querySettings,
          };
          await updateQuery(active_project.id, savedQueryId, reqBody);
          dispatch({
            type: QUERY_UPDATED,
            queryId: savedQueryId,
            payload: {
              title,
              settings: querySettings,
            },
          });
          setQuerySaved({ name: title, id: savedQueryId });
        }

        //Factors SAVE_QUERY EDIT_QUERY tracking
        factorsai.track(activeAction, {
          query_type: queryType,
          saved_query_id: savedQueryId,
          query_title: title,
        });

        notification.success({
          message: 'Report Saved Successfully',
          duration: 5,
        });

        updateLocalReducer({ type: TOGGLE_APIS_CALLED });
        dispatch(fetchWeeklyIngishtsMetaData(active_project.id));
        onSuccess();
      } catch (err) {
        updateLocalReducer({ type: TOGGLE_APIS_CALLED });
        console.log(err);
        console.log(err.response);
        notification.error({
          message: 'Error!',
          description: 'Something went wrong.',
          duration: 5,
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
    ]
  );

  return (
    <>
      <QueryActions
        savedQueryId={savedQueryId}
        handleSaveClick={handleSaveClick}
        handleEditClick={handleEditClick}
        handleDeleteReport={handleDeleteReport}
        toggleAddToDashboardModal={toggleAddToDashModal}
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
        queryType={queryType}
        requestQuery={requestQuery}
      />
    </>
  );
}

export default SaveQuery;
``