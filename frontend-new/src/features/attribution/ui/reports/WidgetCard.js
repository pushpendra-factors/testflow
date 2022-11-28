import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState
} from 'react';
import { useSelector } from 'react-redux';
import { SVG } from 'Components/factorsComponents';
import _ from 'lodash';
import {
  ATTRIBUTION_METRICS,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_WEB
} from 'Utils/constants';
import CardContent from './CardContent';
import styles from './index.module.scss';
import {
  getDataFromServer,
  getSavedAttributionMetrics,
  getValidGranularityForSavedQueryWithSavedGranularity
} from 'Views/Dashboard/utils';
import { initialState } from 'Views/CoreQuery/utils';
import { shouldDataFetch } from 'Utils/dataFormatter';
import { useHistory } from 'react-router-dom';
import { ATTRIBUTION_ROUTES } from '../../utils/constants';

function WidgetCard({ unit, onDrop, durationObj }) {
  const { data: savedQueries } = useSelector((state) => state.queries);
  const history = useHistory();

  const { active_project: activeProject } = useSelector(
    (state) => state.global
  );
  const savedQuery = useMemo(() => {
    return _.find(savedQueries, (sq) => sq.id === unit.query_id);
  }, [savedQueries]);

  const durationWithSavedFrequency = useMemo(() => {
    if (_.get(savedQuery, 'query.query_group', null)) {
      const savedFrequency = _.get(
        savedQuery,
        'query.query_group.0.gbt',
        'date'
      );
      const frequency = getValidGranularityForSavedQueryWithSavedGranularity({
        durationObj,
        savedFrequency
      });
      return {
        ...durationObj,
        frequency
      };
    } else if (_.get(savedQuery, 'query.cl', null) === QUERY_TYPE_KPI) {
      const savedFrequency = _.get(savedQuery, 'query.qG.1.gbt', 'date');
      const frequency = getValidGranularityForSavedQueryWithSavedGranularity({
        durationObj,
        savedFrequency
      });
      return {
        ...durationObj,
        frequency
      };
    }
    return durationObj;
  }, [durationObj, savedQuery]);

  const [attributionMetrics, setAttributionMetrics] = useState([
    ...ATTRIBUTION_METRICS
  ]);

  useEffect(() => {
    if (
      unit.query &&
      unit.query.settings &&
      unit.query.settings.attributionMetrics
    ) {
      setAttributionMetrics(
        getSavedAttributionMetrics(
          JSON.parse(unit.query.settings.attributionMetrics)
        )
      );
    }
  }, [unit.query.settings]);

  const handleReportClick = () => {
    if (unit?.query_id) {
      history.push({
        pathname: ATTRIBUTION_ROUTES.report,
        search: `?${new URLSearchParams({ queryId: unit.query_id }).toString()}`
      });
    }
  };

  //this part needs for for setting the resultState state variable for now which is used as dummy import in next component card content.
  const [resultState, setResultState] = useState(initialState);
  const hasComponentUnmounted = useRef(false);

  //the getData function needs work used for setting the resultState below is the getDate function and the useeffect

  // const getData = useCallback(
  //   async (refresh = false) => {
  //     try {
  //       hasComponentUnmounted.current = false;
  //       setResultState({
  //         ...initialState,
  //         loading: true
  //       });

  //       let queryType;
  //       let apiCallStatus = {
  //         required: true,
  //         message: null
  //       };

  //       if (unit.query.query.query_group) {
  //         if (
  //           unit.query.query.cl &&
  //           unit.query.query.cl === QUERY_TYPE_CAMPAIGN
  //         ) {
  //           queryType = QUERY_TYPE_CAMPAIGN;
  //         } else {
  //           queryType = QUERY_TYPE_EVENT;
  //         }
  //       } else if (unit.query.query.cl === QUERY_TYPE_KPI) {
  //         queryType = QUERY_TYPE_KPI;
  //       } else if (
  //         unit.query.query.cl &&
  //         unit.query.query.cl === QUERY_TYPE_ATTRIBUTION
  //       ) {
  //         apiCallStatus = shouldDataFetch(durationWithSavedFrequency);
  //         queryType = QUERY_TYPE_ATTRIBUTION;
  //       } else if (
  //         unit.query.query.cl &&
  //         unit.query.query.cl === QUERY_TYPE_WEB
  //       ) {
  //         queryType = QUERY_TYPE_WEB;
  //       } else if (
  //         unit.query.query.cl &&
  //         unit.query.query.cl === QUERY_TYPE_PROFILE
  //       ) {
  //         queryType = QUERY_TYPE_PROFILE;
  //       } else {
  //         queryType = QUERY_TYPE_FUNNEL;
  //       }

  //       let lastRefreshedAt = null;
  //       if (apiCallStatus.required) {
  //         const res = await getDataFromServer(
  //           unit.query,
  //           unit.id,
  //           unit.dashboard_id,
  //           durationWithSavedFrequency,
  //           refresh,
  //           activeProject.id
  //         );
  //         if (!hasComponentUnmounted.current) {
  //           onDataLoadSuccess({ unitId: unit.id });
  //         }
  //         if (
  //           queryType === QUERY_TYPE_FUNNEL &&
  //           !hasComponentUnmounted.current
  //         ) {
  //           lastRefreshedAt = _.get(
  //             res,
  //             'data.cache_meta.last_computed_at',
  //             null
  //           );
  //           setResultState({
  //             ...initialState,
  //             data: res.data.result
  //           });
  //         } else if (
  //           queryType === QUERY_TYPE_PROFILE &&
  //           !hasComponentUnmounted.current
  //         ) {
  //           lastRefreshedAt = _.get(
  //             res,
  //             'data.cache_meta.last_computed_at',
  //             null
  //           );
  //           setResultState({
  //             ...initialState,
  //             data: res.data.result
  //           });
  //         } else if (
  //           queryType === QUERY_TYPE_ATTRIBUTION &&
  //           !hasComponentUnmounted.current
  //         ) {
  //           lastRefreshedAt = _.get(
  //             res,
  //             'data.cache_meta.last_computed_at',
  //             null
  //           );
  //           setResultState({
  //             ...initialState,
  //             data: res.data.result,
  //             apiCallStatus
  //           });
  //         } else if (
  //           queryType === QUERY_TYPE_CAMPAIGN &&
  //           !hasComponentUnmounted.current
  //         ) {
  //           lastRefreshedAt = _.get(
  //             res,
  //             'data.cache_meta.last_computed_at',
  //             null
  //           );
  //           setResultState({
  //             ...initialState,
  //             data: res.data.result
  //           });
  //         } else if (
  //           queryType === QUERY_TYPE_KPI &&
  //           !hasComponentUnmounted.current
  //         ) {
  //           lastRefreshedAt = _.get(
  //             res,
  //             'data.cache_meta.last_computed_at',
  //             null
  //           );
  //           setResultState({
  //             ...initialState,
  //             data: res.data.result || res.data
  //           });
  //         } else {
  //           if (!hasComponentUnmounted.current) {
  //             lastRefreshedAt = _.get(
  //               res,
  //               'data.cache_meta.last_computed_at',
  //               null
  //             );
  //             const resultGroup = res.data.result.result_group;
  //             const equivalentQuery = getStateQueryFromRequestQuery(
  //               unit.query.query.query_group[0]
  //             );
  //             const appliedBreakdown = [
  //               ...equivalentQuery.breakdown.event,
  //               ...equivalentQuery.breakdown.global
  //             ];

  //             if (unit.query.query.query_group.length === 1) {
  //               setResultState({
  //                 ...initialState,
  //                 data: resultGroup[0]
  //               });
  //             } else if (unit.query.query.query_group.length === 3) {
  //               const userData = formatApiData(resultGroup[0], resultGroup[1]);
  //               const sessionsData = resultGroup[2];
  //               const activeUsersData = calculateActiveUsersData(
  //                 userData,
  //                 sessionsData,
  //                 appliedBreakdown
  //               );
  //               setResultState({
  //                 ...initialState,
  //                 data: activeUsersData
  //               });
  //             } else if (unit.query.query.query_group.length === 4) {
  //               const eventsData = formatApiData(
  //                 resultGroup[0],
  //                 resultGroup[1]
  //               );
  //               const userData = formatApiData(resultGroup[2], resultGroup[3]);
  //               const frequencyData = calculateFrequencyData(
  //                 eventsData,
  //                 userData,
  //                 appliedBreakdown
  //               );
  //               setResultState({
  //                 ...initialState,
  //                 data: frequencyData
  //               });
  //             } else {
  //               setResultState({
  //                 ...initialState,
  //                 data: formatApiData(resultGroup[0], resultGroup[1])
  //               });
  //             }
  //           }
  //         }
  //         if (lastRefreshedAt != null && !hasComponentUnmounted.current) {
  //           setOldestRefreshTime((currValue) => {
  //             if (currValue == null || lastRefreshedAt < currValue) {
  //               return lastRefreshedAt;
  //             }
  //             return currValue;
  //           });
  //         }
  //       } else {
  //         setResultState({
  //           ...initialState,
  //           apiCallStatus
  //         });
  //       }
  //     } catch (err) {
  //       console.log(err);
  //       console.log(err.response);
  //       if (!hasComponentUnmounted.current) {
  //         onDataLoadSuccess({ unitId: unit.id });
  //       }
  //       setResultState({
  //         ...initialState,
  //         error: true
  //       });
  //     }
  //   },
  //   [
  //     activeProject.id,
  //     unit.query,
  //     unit.id,
  //     unit.dashboard_id,
  //     durationWithSavedFrequency,
  //     onDataLoadSuccess
  //   ]
  // );

  // useEffect(() => {
  //   getData();
  //   return () => {
  //     hasComponentUnmounted.current = true;
  //   };
  // }, [getData, durationWithSavedFrequency]);

  return (
    <div
      className={`py-3 flex ${styles.widgetCard_h} mx-auto`}
      onClick={handleReportClick}
    >
      <div
        // id={`card-${unit.id}`}
        // ref={cardRef}
        className={`fa-dashboard--widget-card h-full w-full flex ${styles.widgetCard_max_w}`}
      >
        <div className={'flex justify-between items-start w-full'}>
          <div className={'w-full flex flex-1 flex-col h-full justify-between'}>
            <div className={` flex items-center justify-between px-4`}>
              {/* <div
                className="widget-card--title-container py-3 flex truncate cursor-pointer items-center w-full mr-2"
                // onClick={handleEditQuery}
              >
                <div className="flex  items-center">
                  <Tooltip title={unit?.query?.title} mouseEnterDelay={0.2}>
                    <Text
                      ellipsis
                      type={'title'}
                      level={6}
                      weight={'bold'}
                      extraClass={`widget-card--title m-0 mr-1 flex`}
                    >
                      {unit?.query?.title}
                    </Text>
                  </Tooltip>
                </div>
                <SVG
                  extraClass={`widget-card--expand-icon ml-1`}
                  size={20}
                  color={'grey'}
                  name="arrowright"
                />
              </div> */}
              {/* <div className="flex items-center">
                {resultState.apiCallStatus &&
                resultState.apiCallStatus.required &&
                resultState.apiCallStatus.message ? (
                  <Tooltip
                    mouseEnterDelay={0.2}
                    title={resultState.apiCallStatus.message}
                  >
                    <div className="cursor-pointer">
                      <SVG color="#dea069" name={'warning'} />
                    </div>
                  </Tooltip>
                ) : null}
                <Dropdown
                  placement="bottomRight"
                  overlay={getMenu()}
                  trigger={['hover']}
                >
                  <Button
                    type="text"
                    icon={<SVG size={20} name={'threedot'} color={'grey'} />}
                  />
                </Dropdown>
              </div> */}
            </div>
            {/* <DashboardContext.Provider
              value={{
                attributionMetrics,
                setAttributionMetrics,
                handleEditQuery
              }}
            > */}
            <CardContent
              durationObj={durationWithSavedFrequency}
              unit={unit}
              attributionMetrics={attributionMetrics}
              setAttributionMetrics={setAttributionMetrics}
              // resultState={resultState}
            />
            {/* </DashboardContext.Provider> */}
          </div>
        </div>
      </div>
      {/* <div
        id={`resize-${unit.id}`}
        className={'fa-widget-card--resize-container'}
      >
        <span className={'fa-widget-card--resize-contents'}>
          {unit.cardSize === 0 ? (
            <>
              <a href="#!" onClick={changeCardSize.bind(this, 1)}>
                <RightOutlined />
              </a>
              <a href="#!" onClick={changeCardSize.bind(this, 2)}>
                <LeftOutlined />
              </a>
            </>
          ) : null}
          {unit.cardSize === 1 ? (
            <a href="#!" onClick={changeCardSize.bind(this, 0)}>
              <LeftOutlined />
            </a>
          ) : null}
          {unit.cardSize === 2 ? (
            <a href="#!" onClick={changeCardSize.bind(this, 0)}>
              <RightOutlined />
            </a>
          ) : null}
        </span>
      </div> */}
    </div>
  );
}

export default WidgetCard;
