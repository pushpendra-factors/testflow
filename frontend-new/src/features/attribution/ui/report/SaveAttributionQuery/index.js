import React, {
  useCallback,
  useState,
  useContext,
  useMemo,
  useEffect
} from 'react';
import { Button, notification, Tooltip } from 'antd';
import { useDispatch, useSelector } from 'react-redux';
import { BUTTON_TYPES } from 'Constants/buttons.constants';
import SaveQueryModal from './SaveQueryModal';
import ControlledComponent from 'Components/ControlledComponent';
import { SVG } from 'factorsComponents';
import { isStringLengthValid } from 'Utils/global';
import MomentTz from 'Components/MomentTz';
import { apiChartAnnotations } from 'Utils/constants';
import { getChartType } from 'Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import { updateQuery } from 'Reducers/coreQuery/services';
import { get } from 'lodash';
import { EMPTY_ARRAY } from 'Utils/global';
import factorsai from 'factorsai';
import { fetchWeeklyIngishtsMetaData } from 'Reducers/insights';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import { ACTION_TYPES } from './constants';
import FaSelect from 'Components/FaSelect';
import styles from './index.module.scss';
import { ATTRIBUTION_ROUTES } from 'Attribution/utils/constants';
import DeleteQueryModal from 'Components/DeleteQueryModal/index';
import { deleteReport } from 'Reducers/coreQuery/services';
import { getErrorMessage } from 'Utils/dataFormatter';
import { saveAttributionQuery } from 'Attribution/state/services';
import useQuery from 'hooks/useQuery';
import {
  ATTRIBUTION_DASHBOARD_UNIT_ADDED,
  ATTRIBUTION_QUERY_CREATED,
  ATTRIBUTION_QUERY_DELETED,
  ATTRIBUTION_QUERY_UPDATED
} from 'Attribution/state/action.constants';

const SaveAttributionQuery = ({
  requestQuery,
  getCurrentSorter,
  savedQueryId,
  queryType,
  breakdown,
  attributionsState,
  setQuerySaved,
  queryTitle,
  showSaveOrUpdateModal
}) => {
  const [showSaveModal, setShowSaveModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [options, setOptions] = useState(false);
  const [loading, setLoading] = useState(false);
  const [activeAction, setActiveAction] = useState(null);
  const routerQuery = useQuery();
  const paramQueryId = routerQuery.get('queryId');
  const dispatch = useDispatch();
  const { active_project } = useSelector((state) => state.global);
  const savedQueries = useSelector((state) =>
    get(state, 'attributionDashboard.attributionQueries.data', EMPTY_ARRAY)
  );
  const { agent_details } = useSelector((state) => state.agent);

  const toggleSaveModalVisibility = useCallback(() => {
    setShowSaveModal((flag) => !flag);
  }, []);

  const {
    attributionMetrics,
    coreQueryState: { chartTypes }
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

      if (activeAction === ACTION_TYPES.SAVE) {
        querySettings.dashboardPresentation = dashboardPresentation;
        const res = await saveAttributionQuery(
          active_project.id,
          title,
          query,
          3,
          querySettings
        );
        queryId = res.data.query.id;
        //updating locally saved attribution queries
        dispatch({ type: ATTRIBUTION_QUERY_CREATED, payload: res.data.query });
        //updating locally saved attribution dashboard units
        dispatch({
          type: ATTRIBUTION_DASHBOARD_UNIT_ADDED,
          payload: res.data.dashboard_unit
        });
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
        queryId = savedQueryId;

        dispatch({
          type: ATTRIBUTION_QUERY_UPDATED,
          queryId,
          payload: {
            title,
            settings: updatedSettings
          }
        });
      }

      setQuerySaved({ name: title, id: queryId });
      // Factors SAVE_QUERY EDIT_QUERY tracking
      factorsai.track(activeAction, {
        email_id: agent_details?.email,
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
      setQuerySaved(false);
      dispatch({ type: ATTRIBUTION_QUERY_DELETED, payload: savedQueryId });
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

  const handleUpdateClick = async () => {
    try {
      const oldQuery = savedQueries.find((elem) => elem.id === paramQueryId);
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

      const updatedSettings = {
        ...oldQuery.settings,
        ...querySettings
      };

      const reqBody = {
        title: oldQuery.title,
        query,
        settings: updatedSettings
      };

      await updateQuery(active_project.id, paramQueryId, reqBody);
      dispatch({
        type: ATTRIBUTION_QUERY_UPDATED,
        queryId: paramQueryId,
        payload: {
          title: oldQuery.title,
          settings: updatedSettings,
          query
        }
      });
      setQuerySaved({ name: oldQuery.title, id: paramQueryId, isSaved: true });
      notification.success({
        message: 'Report Saved Successfully',
        duration: 5
      });
    } catch (error) {
      console.error('Error in updating report ---', error);
    }
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

  useEffect(() => {
    if (showSaveOrUpdateModal) {
      if (showSaveOrUpdateModal?.save) {
        handleSaveClick();
      } else if (showSaveOrUpdateModal?.update) {
        handleUpdateClick();
      }
    }
  }, [showSaveOrUpdateModal]);

  return (
    <div className='flex gap-x-2 items-center'>
      {paramQueryId && !savedQueryId && (
        <Tooltip placement='bottom' title='Save as New'>
          <Button
            onClick={handleSaveClick}
            size='large'
            type='text'
            icon={<SVG name={'pluscopy'} />}
          ></Button>
        </Tooltip>
      )}
      {savedQueryId ? (
        <Tooltip placement='bottom' title={'No changes to be saved'}>
          <Button
            onClick={handleSaveClick}
            disabled={savedQueryId}
            type={BUTTON_TYPES.PRIMARY}
            size={'large'}
            icon={<SVG name={'save'} size={20} color={'white'} />}
          >
            {'Save'}
          </Button>
        </Tooltip>
      ) : (
        <Button
          onClick={paramQueryId ? handleUpdateClick : handleSaveClick}
          type={BUTTON_TYPES.PRIMARY}
          size={'large'}
          disabled={savedQueryId}
          icon={<SVG name={'save'} size={20} color={'white'} />}
        >
          {'Save'}
        </Button>
      )}
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
