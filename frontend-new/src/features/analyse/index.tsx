import React, { useCallback, useEffect, useMemo, useReducer, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { useParams } from "react-router-dom";
import get from 'lodash/get';

import { fetchQueries, getEventsData } from "Reducers/coreQuery/services";
import { EMPTY_ARRAY, EMPTY_OBJECT } from "Utils/global";
import AnalysisHeader from "Views/CoreQuery/AnalysisResultsPage/AnalysisHeader";
import { EACH_USER_TYPE, QUERY_TYPE_EVENT, QUERY_TYPE_FUNNEL, REPORT_SECTION, TOTAL_EVENTS_CRITERIA, TOTAL_USERS_CRITERIA } from "Utils/constants";
import { CORE_QUERY_INITIAL_STATE, DEFAULT_PIVOT_CONFIG, INITIAL_STATE as INITIAL_RESULT_STATE, SET_COMPARISON_SUPPORTED, SET_SAVED_QUERY_SETTINGS, UPDATE_PIVOT_CONFIG } from "Views/CoreQuery/constants";
import { formatApiData, getStateQueryFromRequestQuery, isComparisonEnabled } from "Views/CoreQuery/utils";
import { getQueryOptionsFromEquivalentQuery } from "./utils";
import { CoreQueryState, QueryParams } from "./types";
import { SHOW_ANALYTICS_RESULT } from "Reducers/types";
import CoreQueryReducer from "Views/CoreQuery/CoreQueryReducer";
import { INITIALIZE_GROUPBY } from "Reducers/coreQuery/actions";
import { ErrorBoundary } from "react-error-boundary";
import { FaErrorComp, FaErrorLog, SVG } from "Components/factorsComponents";
import QueryComposer from "Components/QueryComposer";
import { Button } from "antd";
import logger from "Utils/logger";
import ReportContent from "Views/CoreQuery/AnalysisResultsPage/ReportContent";


const CoreQuery = () => {

    // Query params
    const { query_id, query_type } = useParams<QueryParams>();

    const [activeTab, setActiveTab] = useState(1);

    // Redux States
    const { active_project } = useSelector((state: any) => state.global);
    const { show_criteria: result_criteria, user_type } = useSelector((state: any) => state.analyticsQuery);
    const { models, eventNames } = useSelector((state: any) => state.coreQuery);
    const savedQueries = useSelector((state: any) =>
        get(state, 'queries.data', EMPTY_ARRAY)
    );

    // Local states
    const [coreQueryState, setCoreQueryState] = useState<CoreQueryState>(new CoreQueryState());
    const [queryOpen, setQueryOpen] = useState(true);

    const dispatch = useDispatch();
    const [coreQueryReducerState, localDispatch] = useReducer(
        CoreQueryReducer,
        CORE_QUERY_INITIAL_STATE
    );
    const renderedCompRef = useRef<any>(null);

    const getCurrentSorter = useCallback(() => {
        if (renderedCompRef.current && renderedCompRef.current.currentSorter) {
            return renderedCompRef.current.currentSorter;
        }
        return [];
    }, []);


    // Use Effects
    useEffect(() => {
        if (!savedQueries || !savedQueries.length) {
            fetchQueries(active_project.id);
        }
    }, [savedQueries])

    useEffect(() => {
        if (query_id && query_id != '' && query_type && savedQueries?.length) {
            runEventsQueryFromUrl();
        }
    }, [query_id, query_type, savedQueries])

    const getQueryFromHashId = () => savedQueries?.find((quer: any) => quer.id_text === query_id);

    const runEventsQueryFromUrl = () => {
        const queryToAdd = getQueryFromHashId();
        if (queryToAdd) {
            // updateResultState({ ...initialState, loading: true });
            getEventsData(active_project.id, null, null, false, query_id).then(
                (res) => {
                    const equivalentQuery = getStateQueryFromRequestQuery(
                        queryToAdd?.query?.query_group[0]
                    );
                    const queryState = new CoreQueryState();
                    queryState.queryType = QUERY_TYPE_EVENT;
                    queryState.querySaved = { name: queryToAdd.title, id: queryToAdd.id };
                    queryState.requestQuery = queryToAdd?.query?.query_group;
                    queryState.showResult = true;
                    queryState.loading = false;
                    queryState.queries = equivalentQuery.events;
                    queryState.appliedQueries = equivalentQuery.events.map((elem: any) =>
                        elem.alias ? elem.alias : elem.label
                    );
                    queryState.queryOptions = getQueryOptionsFromEquivalentQuery(queryState.queryOptions, equivalentQuery)

                    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
                    dispatch({
                        type: SET_COMPARISON_SUPPORTED,
                        payload: isComparisonEnabled(
                            queryState.queryType,
                            equivalentQuery.events,
                            equivalentQuery.breakdown,
                            models
                        )
                    });

                    dispatch({
                        type: INITIALIZE_GROUPBY,
                        payload: equivalentQuery.breakdown
                    });

                    const newAppliedBreakdown = [
                        ...equivalentQuery.breakdown.event,
                        ...equivalentQuery.breakdown.global
                    ];
                    queryState.appliedBreakdown = newAppliedBreakdown;

                    // updateAppliedBreakdown();
                    dispatch({ type: UPDATE_PIVOT_CONFIG, payload: { ...DEFAULT_PIVOT_CONFIG } });
                    dispatch({ type: SET_SAVED_QUERY_SETTINGS, payload: EMPTY_OBJECT })

                    updateResultFromSavedQuery(res, queryState);
                    setCoreQueryState(queryState);
                },
                (err) => {
                    logger.error(err);
                }
            );
        }
    };

    const arrayMapper = coreQueryState.appliedQueries.map((q, index) => ({
            eventName: q,
            index,
            mapper: `event${index + 1}`,
            displayName: eventNames[q] ? eventNames[q] : q
          }))

    const updateResultFromSavedQuery = (res: any, qState: CoreQueryState) => {
        const data = res.data.result || res.data;
        let resultSt;
        if (result_criteria === TOTAL_EVENTS_CRITERIA) {
            resultSt = {
                ...INITIAL_RESULT_STATE,
                data: formatApiData(data.result_group[0], data.result_group[1]),
                apiCallStatus: res.status
            }
        } else if (result_criteria === TOTAL_USERS_CRITERIA) {
            if (user_type === EACH_USER_TYPE) {
                resultSt = {
                    ...INITIAL_RESULT_STATE,
                    data: formatApiData(data.result_group[0], data.result_group[1]),
                    apiCallStatus: res.status
                }
            } else {
                resultSt = {
                    ...INITIAL_RESULT_STATE,
                    data: data.result_group[0],
                    apiCallStatus: res.status
                }
            }
        }
        qState.setItem('resultState', resultSt)
    };

    const renderQueryComposerNew = () => {
        return (
            <div
                className={`query_card_cont ${queryOpen ? `query_card_open` : `query_card_close`
                    }`}
                onClick={() => !queryOpen && setQueryOpen(true)}
            >
                <div className="query_composer">{renderComposer()}</div>
                <Button size="large" className="query_card_expand">
                    <SVG name="expand" size={20} />
                    Expand
                </Button>
            </div>
        );
    }

    const renderComposer = () => {
        if (coreQueryState.queryType === QUERY_TYPE_FUNNEL || coreQueryState.queryType === QUERY_TYPE_EVENT) {
            return (
                <QueryComposer
                    queries={coreQueryState.queries}
                    setQueries={() => { }}
                    runQuery={() => { }}
                    eventChange={() => { }}
                    queryType={coreQueryState.queryType}
                    queryOptions={coreQueryState.queryOptions}
                    setQueryOptions={() => { }}
                    runFunnelQuery={() => { }}
                    collapse={coreQueryState.showResult}
                    setCollapse={() => setQueryOpen(false)}
                />
            );
        }
    }

    const renderMain = () => {
        if (coreQueryState.showResult && !coreQueryState.resultState.loading) {
            return (<>
                <AnalysisHeader
                    isFromAnalysisPage={false}
                    requestQuery={coreQueryState.requestQuery}
                    onBreadCrumbClick={() => { console.log("breadcrumb click") }}
                    queryType={coreQueryState.queryType}
                    queryTitle={coreQueryState.querySaved ? coreQueryState.querySaved?.name : null}
                    setQuerySaved={(v: any) => coreQueryState.setItem('querySaved', v)}
                    breakdownType={EACH_USER_TYPE}
                    changeTab={(v: any) => coreQueryState.setItem('activeTab', v)}
                    activeTab={coreQueryState.activeTab}
                    getCurrentSorter={getCurrentSorter}
                    savedQueryId={coreQueryState.querySaved ? coreQueryState.querySaved.id : null}
                    breakdown={coreQueryState.appliedBreakdown}
                    dateFromTo={{ from: coreQueryState.requestQuery.fr, to: coreQueryState.requestQuery.to }}
                />
                    <div className="mt-24 px-8">
                        <ErrorBoundary
                            fallback={
                                <FaErrorComp
                                    size="medium"
                                    title="Analyse Results Error"
                                    subtitle="We are facing trouble loading Analyse results. Drop us a message on the in-app chat." className={undefined} type={undefined}                            />
                            }
                            onError={FaErrorLog}
                        >
                            {Number(coreQueryState.activeTab) === 1 && (
                                <>
                                    {renderQueryComposerNew()}
                                    {coreQueryState.requestQuery && (
                                        <ReportContent
                                            breakdownType={EACH_USER_TYPE}
                                            queryType={coreQueryState.queryType}
                                            renderedCompRef={renderedCompRef}
                                            breakdown={coreQueryState.appliedBreakdown}
                                            handleChartTypeChange={()=>{}}
                                            queryOptions={coreQueryState.queryOptions}
                                            arrayMapper={arrayMapper}
                                            resultState={coreQueryState.resultState}
                                            queryTitle={coreQueryState.querySaved.name}
                                            section={REPORT_SECTION}
                                            eventPage={result_criteria}
                                            onReportClose={()=>{}}
                                            handleGranularityChange={()=>{}}
                                            setDrawerVisible={()=>{}}
                                        />
                                    )}
                                </>
                            )}
                        </ErrorBoundary>
                    </div>
                </>)
        } else {
            return null;
        }
    }

    return renderMain();
}

export default CoreQuery;