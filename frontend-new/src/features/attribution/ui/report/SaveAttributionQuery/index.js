import React, { useCallback, useState, useContext, useMemo } from 'react';
import { Button, notification } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { BUTTON_TYPES } from 'Constants/buttons.constants';
import SaveQueryModal from './SaveQueryModal';
import ControlledComponent from 'Components/ControlledComponent';
import { SVG } from 'factorsComponents';
import { isStringLengthValid } from 'Utils/global';
import MomentTz from 'Components/MomentTz';
import { apiChartAnnotations } from 'Utils/constants';
import { getChartType } from 'Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import { saveQuery, updateQuery } from 'Reducers/coreQuery/services';
import { get } from 'lodash';
import { EMPTY_ARRAY } from 'Utils/global';
import { saveQueryToDashboard } from 'Reducers/dashboard/services';
import factorsai from 'factorsai';
import { fetchWeeklyIngishtsMetaData } from 'Reducers/insights';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import { QUERY_CREATED, QUERY_UPDATED } from 'Reducers/types';
import { ACTION_TYPES } from './constants';
import FaSelect from 'Components/FaSelect';
import styles from './index.module.scss';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import DeleteQueryModal from 'Components/DeleteQueryModal/index';
import { deleteReport } from 'Reducers/coreQuery/services';
import { getErrorMessage } from 'Utils/dataFormatter';

const SaveAttributionQuery = ({
  requestQuery,
  getCurrentSorter,
  savedQueryId,
  queryType,
  breakdown,
  attributionsState,
  setQuerySaved,
  queryTitle
}) => {
  const [showSaveModal, setShowSaveModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [options, setOptions] = useState(false);
  const [loading, setLoading] = useState(false);
  const [activeAction, setActiveAction] = useState(null);
  const dispatch = useDispatch();
  const { active_project } = useSelector((state) => state.global);
  const { activeDashboard } = useSelector((state) => state.dashboard);
  const savedQueries = useSelector((state) =>
    get(state, 'queries.data', EMPTY_ARRAY)
  );

  const toggleSaveModalVisibility = useCallback(() => {
    setShowSaveModal((flag) => !flag);
  }, []);

  const {
    attributionMetrics,
    coreQueryState: { chartTypes, pivotConfig }
  } = useContext(CoreQueryContext);

  const savedQueryPresentation = useMemo(() => {
    if (!savedQueryId) return null;
    const query = savedQueries.find((elem) => elem.id === savedQueryId);
    return query?.settings?.dashboardPresentation;
  }, [savedQueryId, savedQueries]);

  const handleSave = async ({ title, dashboardPresentation, onSuccess }) => {
    try {
      if (!isStringLengthValid(title)) {
        notification.error({
          message: 'Incorrect Input!',
          description: 'Please Enter query title',
          duration: 5
        });
        return false;
      }

      const startOfWeek = MomentTz().startOf('week').utc().unix();
      const todayNow = MomentTz().utc().unix();
      setLoading(true);
      const query = {
        ...requestQuery,
        query: {
          ...requestQuery.query,
          from: startOfWeek,
          to: todayNow
        }
      };

      const querySettings = {
        ...getCurrentSorter(),
        chart:
          apiChartAnnotations[
            getChartType({
              queryType,
              chartTypes,
              breakdown,
              attributionModels: attributionsState.models,
              campaignGroupBy: null
            })
          ]
      };

      querySettings.attributionMetrics = JSON.stringify(attributionMetrics);

      let queryId;
      let addedToDashboard = false;

      if (activeAction === ACTION_TYPES.SAVE) {
        querySettings.dashboardPresentation = dashboardPresentation;
        const res = await saveQuery(
          active_project.id,
          title,
          query,
          2,
          querySettings
        );
        queryId = res.data.id;

        dispatch({ type: QUERY_CREATED, payload: res.data });

        try {
          const reqBody = {
            query_id: queryId
          };

          await saveQueryToDashboard(
            active_project.id,
            [activeDashboard.id].join(','),
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
      } else {
        const queryGettingUpdated = savedQueries.find(
          (elem) => elem.id === savedQueryId
        );

        const updatedSettings = {
          ...queryGettingUpdated.settings,
          ...querySettings
        };

        updatedSettings.dashboardPresentation = dashboardPresentation;

        const reqBody = {
          title,
          settings: updatedSettings
        };

        await updateQuery(active_project.id, savedQueryId, reqBody);

        dispatch({
          type: QUERY_UPDATED,
          queryId,
          payload: {
            title,
            settings: updatedSettings
          }
        });
        queryId = savedQueryId;
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

      notification.success({
        message: 'Report Saved Successfully',
        duration: 5
      });

      setLoading(false);
      dispatch(fetchWeeklyIngishtsMetaData(active_project.id));
      onSuccess();
    } catch (err) {
      setLoading(false);
      console.log(err);
      console.log(err.response);
      notification.error({
        message: 'Error!',
        description: `${err?.data?.error}`,
        duration: 5
      });
    }
  };

  const handleDelete = useCallback(async () => {
    try {
      setLoading(true);
      await deleteReport({
        project_id: active_project.id,
        queryId: savedQueryId
      });
      setLoading(false);
      toggleDeleteModal();
      setQuerySaved(null);
      dispatch({ type: QUERY_DELETED, payload: savedQueryId });
      notification.success({
        message: 'Report Deleted Successfully',
        duration: 5
      });
      history.push(ATTRIBUTION_ROUTES.reports);
    } catch (err) {
      setLoading(false);
      notification.error({
        message: 'Something went wrong!',
        description: getErrorMessage(err),
        duration: 5
      });
    }
  }, [dispatch, active_project, savedQueryId]);

  const handleSaveClick = () => {
    setShowSaveModal(true);
    setActiveAction(ACTION_TYPES.SAVE);
  };

  const handleEditClick = () => {
    setShowSaveModal(true);
    setActiveAction(ACTION_TYPES.EDIT);
  };

  const setActions = (option) => {
    const val = option[1];
    if (val === 'edit') {
      handleEditClick();
    } else if (val === 'trash') {
      toggleDeleteModal();
    }
    setOptions(false);
  };

  const getActionsMenu = () => {
    return options ? (
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
    ) : null;
  };

  const toggleDeleteModal = () => {
    setShowDeleteModal((val) => !val);
  };

  return (
    <div className='flex gap-x-2 items-center'>
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
      <ControlledComponent controller={!!savedQueryId}>
        <div className={'relative'}>
          <Button
            size='large'
            type='text'
            icon={<SVG name={'threedot'} />}
            onClick={() => setOptions(!options)}
            ƒ
          ></Button>
          {getActionsMenu()}
        </div>
      </ControlledComponent>

      <SaveQueryModal
        toggleSaveModalVisibility={toggleSaveModalVisibility}
        visibility={showSaveModal}
        isLoading={loading}
        activeAction={activeAction}
        onSubmit={handleSave}
        queryTitle={queryTitle}
        savedQueryPresentation={savedQueryPresentation}
      />

      <DeleteQueryModal
        visible={showDeleteModal}
        onDelete={handleDelete}
        toggleModal={toggleDeleteModal}
        isLoading={loading}
      />
    </div>
  );
};

export default SaveAttributionQuery;
