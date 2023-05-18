import React, { useCallback, useEffect, useReducer } from 'react';
import { Button, Divider, notification, Spin, Tooltip } from 'antd';
import { Link, useHistory } from 'react-router-dom';
import { Text, SVG } from 'Components/factorsComponents';
import FaPublicHeader from 'Components/FaPublicHeader';
import style from './index.module.scss';
import { ConnectedProps, connect, useSelector } from 'react-redux';

import SideDrawer from './components/SideDrawer';
import {
  generateEllipsisOption,
  generateUnsavedReportDateRanges,
  getFormattedRange,
  getPublicUrl,
  parseResultGroupResponse,
  parseSavedReportDates
} from '../utils';
// import FaSelect from 'Components/FaSelect';
import FaSelect from 'Components/GenericComponents/FaSelect';
import {
  ShareApiResponse,
  WeekStartEnd,
  ReportApiResponse,
  SavedReportDatesApiResponse,
  PageViewUrlApiResponse
} from '../types';
import ReportTable from './components/ReportTable';
import { ALL_CHANNEL, SHARE_QUERY_PARAMS } from '../const';
import {
  fetchPageViewUrls,
  getSavedReportDates,
  getSixSignalReportData,
  getSixSignalReportPublicData,
  shareSixSignalReport
} from '../state/services';
import useAgentInfo from 'hooks/useAgentInfo';
import ShareModal from './components/ShareModal';
import useQuery from 'hooks/useQuery';
import logger from 'Utils/logger';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { OptionType } from 'Components/GenericComponents/FaSelect/types';
import { setShowAnalyticsResult } from 'Reducers/coreQuery/actions';
import { bindActionCreators } from 'redux';
import {
  VisitorReportActions,
  initialState,
  visitorReportReducer
} from './localStateReducer';

const SixSignalReport = ({
  setShowAnalyticsResult
}: VisitorIdentificationComponentProps) => {
  const [state, localDispatch] = useReducer(visitorReportReducer, initialState);
  const reportData = state.reportData.data;
  const reportDataLoading = state.reportData.loading;

  const { isLoggedIn, email } = useAgentInfo();
  const {
    active_project,
    currentProjectSettings,
    currentProjectSettingsLoading
  } = useSelector((state: any) => state.global);
  const routerQuery = useQuery();
  const history = useHistory();
  const paramQueryId = routerQuery.get(SHARE_QUERY_PARAMS.queryId);
  const paramProjectId = routerQuery.get(SHARE_QUERY_PARAMS.projectId);
  const showShareButton = state.reportData?.data
    ? state.reportData.data.is_shareable && isLoggedIn
    : false;

  const isSixSignalActivated = currentProjectSettings
    ? !currentProjectSettings?.int_factors_six_signal_key &&
      !currentProjectSettings?.int_client_six_signal_key
      ? false
      : true
    : true;

  const showDrawer = () => {
    localDispatch({
      type: VisitorReportActions.SET_DRAWER_VISIBILITY,
      payload: true
    });
  };

  const hideDrawer = () => {
    localDispatch({
      type: VisitorReportActions.SET_DRAWER_VISIBILITY,
      payload: false
    });
  };

  const getOptions = (values: string[], selectedOptions?: string[]) => {
    if (!values || !Array.isArray(values)) return [];
    return values.map((value: string) => {
      return {
        value,
        label: value,
        isSelected: selectedOptions
          ? selectedOptions.indexOf(value) > -1
          : undefined
      };
    });
  };

  const getDateOptions = () => {
    return state.dateValues.map((date) => {
      return {
        label: date.formattedRangeOption,
        value: date.formattedRange
      };
    });
  };

  const handleCampaignApplyClick = (
    _options: OptionType[],
    selectedOption: string[]
  ) => {
    localDispatch({
      type: VisitorReportActions.SET_SELECTED_CAMPAIGNS,
      payload: selectedOption
    });
  };

  const handlePageViewsApplyClick = (
    _options: OptionType[],
    selectedOptions: string[]
  ) => {
    localDispatch({
      type: VisitorReportActions.SET_SELECTED_PAGE_VIEWS,
      payload: selectedOptions
    });
  };

  const getDateObjFromSelectedDate = useCallback(() => {
    const dateObj: WeekStartEnd | undefined = state.dateValues.find(
      (date) => date.formattedRange === state.selectedDate
    );
    return dateObj;
  }, [state.selectedDate, state.dateValues]);

  const handleShareClick = async () => {
    try {
      //checking if share data is already fetched for the dates
      if (state.selectedDate === state.shareData?.data?.dateSelected) {
        localDispatch({
          type: VisitorReportActions.SET_SHARE_MODAL_VISIBILITY,
          payload: true
        });
        return;
      }
      localDispatch({ type: VisitorReportActions.SET_SHARE_DATA_LOADING });
      const dateObj = getDateObjFromSelectedDate();
      if (!dateObj) {
        notification.error({
          message: 'Error',
          description: 'Please select date range to share report',
          duration: 5
        });
        return;
      }

      const res = (await shareSixSignalReport(
        active_project.id,
        dateObj?.from,
        dateObj?.to,
        active_project?.time_zone || 'Asia/Kolkata'
      )) as ShareApiResponse;
      if (res.data) {
        localDispatch({
          type: VisitorReportActions.SET_SHARE_DATA,
          payload: {
            ...res?.data,
            dateSelected: state.selectedDate,
            publicUrl: getPublicUrl(res.data, active_project?.id),
            from: dateObj.from,
            to: dateObj.to,
            timezone: active_project?.time_zone || 'Asia/Kolkata',
            domain: active_project?.name,
            projectId: active_project.id
          }
        });

        localDispatch({
          type: VisitorReportActions.SET_SHARE_MODAL_VISIBILITY,
          payload: true
        });
      } else {
        logger.error('No data found to share', res?.data);
      }
    } catch (error) {
      logger.error('Error in sharing report', error);
      notification.error({
        message: 'Error',
        description: error?.data?.error || 'Something went wrong',
        duration: 5
      });
      localDispatch({ type: VisitorReportActions.SET_SHARE_DATA_FAILED });
    }
  };

  const handleShareModalCancel = () => {
    localDispatch({
      type: VisitorReportActions.SET_SHARE_MODAL_VISIBILITY,
      payload: false
    });
  };

  const handleDateChange = (option: OptionType) => {
    if (state.selectedDate === option.value) {
      localDispatch({
        type: VisitorReportActions.SET_DATE_SELECTION_VISIBILITY,
        payload: false
      });
      return;
    }
    localDispatch({
      type: VisitorReportActions.SET_SELECTED_DATE,
      payload: option.value
    });

    //resetting filters
    localDispatch({ type: VisitorReportActions.RESET_FILTERS });
  };

  //Effect for hiding the side panel and menu
  useEffect(() => {
    let hideSidePanel = false;
    if (!isLoggedIn) {
      setShowAnalyticsResult(true);
      hideSidePanel = true;
    }

    return () => {
      if (hideSidePanel) setShowAnalyticsResult(false);
    };
  }, [isLoggedIn, setShowAnalyticsResult]);

  //Effect for fetching page view Urls
  useEffect(() => {
    const fetchPageUrls = async (projectId: string, queryId?: string) => {
      try {
        localDispatch({ type: VisitorReportActions.SET_PAGE_URL_DATA_LOADING });
        const res = (await fetchPageViewUrls(
          projectId,
          queryId
        )) as PageViewUrlApiResponse;
        if (res?.data) {
          localDispatch({
            type: VisitorReportActions.SET_PAGE_URL_DATA,
            payload: res.data
          });
        }
      } catch (error) {
        logger.error('Error in fetching page urls', error);
        localDispatch({ type: VisitorReportActions.SET_PAGE_URL_DATA_ERROR });
      }
    };
    //remove the below condition when page view is successfully tested
    if (email !== 'solutions@factors.ai') return;
    if (isLoggedIn && active_project?.id) fetchPageUrls(active_project?.id);
    else if (!isLoggedIn && paramProjectId && paramQueryId)
      fetchPageUrls(paramProjectId, paramQueryId);
  }, [isLoggedIn, paramProjectId, paramQueryId, active_project?.id, email]);

  //Effect for fetching dates
  useEffect(() => {
    if (!active_project?.id || !isSixSignalActivated) return;
    const getSavedReports = async () => {
      try {
        const res = (await getSavedReportDates(
          active_project.id
        )) as SavedReportDatesApiResponse;
        if (res?.data) {
          const resDates = [
            ...generateUnsavedReportDateRanges(),
            ...parseSavedReportDates(res.data)
          ];
          localDispatch({
            type: VisitorReportActions.SET_DATE_VALUES,
            payload: resDates
          });
          localDispatch({
            type: VisitorReportActions.SET_SELECTED_DATE,
            payload: resDates?.[0]?.formattedRange
          });
        }
      } catch (error) {
        logger.error('Error in fetching dates', error);
      }
    };
    getSavedReports();
  }, [active_project?.id, isSixSignalActivated]);

  //Effect for fetching the visitor identification public data
  useEffect(() => {
    const fetchPublicData = async () => {
      try {
        if (!isLoggedIn)
          localDispatch({
            type: VisitorReportActions.SET_PAGE_MODE,
            payload: 'public'
          });
        localDispatch({ type: VisitorReportActions.REPORT_DATA_LOADING });
        localDispatch({
          type: VisitorReportActions.SET_SELECTED_DATE,
          payload: ''
        });
        if (paramQueryId && paramProjectId) {
          const res = (await getSixSignalReportPublicData(
            paramProjectId,
            paramQueryId,
            state.selectedPageViews
          )) as ReportApiResponse;

          if (res?.data?.[1]?.result_group) {
            localDispatch({
              type: VisitorReportActions.REPORT_DATA_LOADED,
              payload: res?.data?.[1]
            });
            const _query = res?.data?.[1].query?.six_signal_query_group?.[0];
            if (_query.fr && _query.to) {
              localDispatch({
                type: VisitorReportActions.SET_SELECTED_DATE,
                payload: getFormattedRange(_query.fr, _query.to, _query.tz)
              });
            }
          }
        }
      } catch (error) {
        logger.error('Error in fetching public data', error);
        notification.error({
          message: 'Error',
          description: error?.data?.error || 'Something went wrong',
          duration: 5
        });
        localDispatch({
          type: VisitorReportActions.REPORT_DATA_ERROR
        });
      }
    };
    if (!isLoggedIn && paramQueryId && paramProjectId) fetchPublicData();
    if (!isLoggedIn && (!paramProjectId || !paramQueryId)) {
      history.push('/login');
    }
  }, [
    isLoggedIn,
    paramProjectId,
    paramQueryId,
    history,
    state.selectedPageViews
  ]);

  //Effect for fetching the visitor identification logged in data
  useEffect(() => {
    const fetchDataForLoggedInUser = async () => {
      console.log('fetch data for logged in user called-----', Date.now());
      try {
        localDispatch({ type: VisitorReportActions.REPORT_DATA_LOADING });
        if (state.selectedDate) {
          const dateObj = getDateObjFromSelectedDate();
          if (!dateObj) return;
          const res = (await getSixSignalReportData(
            active_project.id,
            dateObj?.from,
            dateObj?.to,
            active_project?.time_zone || 'Asia/Kolkata',
            dateObj.isSaved,
            state.selectedPageViews
          )) as ReportApiResponse;

          if (res?.data?.[1]?.result_group) {
            localDispatch({
              type: VisitorReportActions.REPORT_DATA_LOADED,
              payload: res?.data?.[1]
            });
          }
        }
      } catch (error) {
        logger.error('Error in fetching data', error);
        localDispatch({
          type: VisitorReportActions.REPORT_DATA_ERROR
        });
      }
    };

    if (
      isLoggedIn &&
      active_project?.id &&
      state.selectedDate &&
      isSixSignalActivated
    )
      fetchDataForLoggedInUser();
  }, [
    isLoggedIn,
    active_project,
    state.selectedDate,
    state.selectedPageViews,
    getDateObjFromSelectedDate,
    isSixSignalActivated
  ]);

  //Effect for formatting data when api data is available.
  useEffect(() => {
    if (reportData) {
      const value = parseResultGroupResponse(reportData.result_group[0]);
      localDispatch({
        type: VisitorReportActions.SET_PARSED_VALUES,
        payload: {
          campaigns: value.campaigns,
          channels:
            value.channels?.length > 0 ? [ALL_CHANNEL, ...value.channels] : []
        }
      });
    }
  }, [reportData]);

  return (
    <div className='flex flex-col'>
      {!isLoggedIn && (
        <FaPublicHeader
          showDrawer={showDrawer}
          handleShareClick={handleShareClick}
          showShareButton={showShareButton}
        />
      )}

      <div className={`${isLoggedIn ? 'px-20' : 'px-24'} pt-16 mt-12`}>
        <div className='flex justify-between align-middle'>
          <div className='flex align-middle gap-6'>
            <div className={style.mixChartContainer}>
              <SVG name={'MixChart'} color='#5ACA89' size={24} />
            </div>
            <div>
              <div>
                <div className='flex'>
                  <Text
                    type={'title'}
                    level={4}
                    weight={'bold'}
                    color='grey-1'
                    extraClass='mb-1'
                  >
                    Top accounts that visited your website{' '}
                  </Text>
                </div>
                <div className='flex items-center flex-wrap gap-1'>
                  <Text type={'paragraph'} mini color='grey'>
                    See which key accounts are engaging with your marketing.
                    Take action and close more deals.
                  </Text>
                  {/* To do: uncomment the below line when learn more link is available */}
                  {/* <Link
                    className='flex items-center font-semibold gap-2'
                    style={{ color: `#1d89ff` }}
                    target='_blank'
                    to={{
                      pathname:
                        'https://www.factors.ai/blog/attribution-reporting-what-you-can-learn-from-marketing-attribution-reports'
                    }}
                  >
                    <Text
                      type={'paragraph'}
                      level={7}
                      weight={'bold'}
                      color='brand-color-6'
                    >
                      Learn more
                    </Text>
                  </Link> */}
                </div>
              </div>
            </div>
          </div>
          <div>
            {/* match account */}
            <ControlledComponent controller={isLoggedIn}>
              <Tooltip
                placement='bottom'
                title={`${
                  showShareButton
                    ? 'Share'
                    : 'Only weekly visitor reports can be shared for easy access'
                }`}
              >
                <Button
                  onClick={handleShareClick}
                  size='large'
                  type='primary'
                  icon={
                    <SVG
                      name={'link'}
                      color={`${showShareButton ? '#fff' : '#b8b8b8'}`}
                    />
                  }
                  disabled={!showShareButton}
                >
                  Share
                </Button>
              </Tooltip>
            </ControlledComponent>
          </div>
        </div>
        <Divider />
        <div className='flex justify-between items-center '>
          <div className='flex items-center gap-4'>
            <div className={style.filter}>
              <Button
                onClick={() =>
                  localDispatch({
                    type: VisitorReportActions.SET_CHANNEL_SELECTION_VISIBILITY,
                    payload: true
                  })
                }
                className={`${style.customButton} flex items-center gap-1`}
              >
                <Text type='title' level={7} extraClass='m-0'>
                  {state.selectedChannel ? state.selectedChannel : 'Channel'}
                </Text>
                <SVG size={14} name='chevronDown' />
              </Button>
              {state.channelSelectionVisibility && (
                <FaSelect
                  options={getOptions(state.channels)}
                  onClickOutside={() =>
                    localDispatch({
                      type: VisitorReportActions.SET_CHANNEL_SELECTION_VISIBILITY,
                      payload: false
                    })
                  }
                  optionClickCallback={(option: OptionType) => {
                    localDispatch({
                      type: VisitorReportActions.SET_SELECTED_CHANNELS,
                      payload: option.value
                    });
                    localDispatch({
                      type: VisitorReportActions.SET_CHANNEL_SELECTION_VISIBILITY,
                      payload: false
                    });
                  }}
                  loadingState={reportDataLoading}
                >
                  {!state.channels?.length ? (
                    <div className='px-2'>
                      <Text type={'title'} level={7} extraClass={'m-0'}>
                        No Channels Found!
                      </Text>
                    </div>
                  ) : null}
                </FaSelect>
              )}
            </div>
            <div className={style.filter}>
              <div>
                <Button
                  className={`${style.customButton} flex items-center gap-1`}
                  onClick={() =>
                    localDispatch({
                      type: VisitorReportActions.SET_CAMPAIGN_SELECT_VISIBILITY,
                      payload: true
                    })
                  }
                >
                  <Text type='title' level={7} extraClass='m-0'>
                    {!state.selectedCampaigns?.length
                      ? 'Campaign'
                      : generateEllipsisOption(state.selectedCampaigns)}
                  </Text>
                  <SVG size={14} name='chevronDown' />
                </Button>
              </div>

              {state.campaignSelectionVisibility && (
                <FaSelect
                  options={getOptions(state.campaigns, state.selectedCampaigns)}
                  onClickOutside={() =>
                    localDispatch({
                      type: VisitorReportActions.SET_CAMPAIGN_SELECT_VISIBILITY,
                      payload: false
                    })
                  }
                  applyClickCallback={handleCampaignApplyClick}
                  allowSearch={state.campaigns?.length > 0}
                  variant='Multi'
                  loadingState={reportDataLoading}
                  allowSearchTextSelection={false}
                >
                  {!state.campaigns?.length ? (
                    <div className='px-2'>
                      <Text type={'title'} level={7} extraClass={'m-0'}>
                        No Campaigns Found!
                      </Text>
                    </div>
                  ) : null}
                </FaSelect>
              )}
            </div>
            <ControlledComponent controller={email === 'solutions@factors.ai'}>
              <div className={style.filter}>
                <div>
                  <Button
                    className={`${style.customButton} flex items-center gap-1`}
                    onClick={() =>
                      localDispatch({
                        type: VisitorReportActions.SET_PAGE_VIEW_SELECTION_VISIBILITY,
                        payload: true
                      })
                    }
                  >
                    <Text type='title' level={7} extraClass='m-0'>
                      {!state.selectedPageViews?.length
                        ? 'Page Viewed'
                        : generateEllipsisOption(state.selectedPageViews)}
                    </Text>
                    <SVG size={14} name='chevronDown' />
                  </Button>
                </div>

                {state.pageViewSelectionVisibility && (
                  <FaSelect
                    options={getOptions(
                      state.pageViewUrls?.data || [],
                      state.selectedPageViews
                    )}
                    onClickOutside={() =>
                      localDispatch({
                        type: VisitorReportActions.SET_PAGE_VIEW_SELECTION_VISIBILITY,
                        payload: false
                      })
                    }
                    applyClickCallback={handlePageViewsApplyClick}
                    allowSearch={
                      state.pageViewUrls?.data
                        ? state.pageViewUrls?.data?.length > 0
                        : false
                    }
                    variant='Multi'
                    loadingState={state.pageViewUrls.loading}
                    maxAllowedSelection={5}
                  >
                    {!state.pageViewUrls?.data?.length ? (
                      <div className='px-2'>
                        <Text type={'title'} level={7} extraClass={'m-0'}>
                          No Page Views!
                        </Text>
                      </div>
                    ) : null}
                  </FaSelect>
                )}
              </div>
            </ControlledComponent>
          </div>

          <div className={style.filter}>
            {state.pageMode === 'public' ? (
              <div className='flex items-center gap-2'>
                {state.selectedDate && (
                  <>
                    <SVG name={'calendar'} color='#8692A3' size={16} />
                    <Text type={'paragraph'} mini extraClass={'m-0'}>
                      {state.selectedDate}
                    </Text>
                  </>
                )}
              </div>
            ) : (
              <Button
                onClick={() =>
                  localDispatch({
                    type: VisitorReportActions.SET_DATE_SELECTION_VISIBILITY,
                    payload: true
                  })
                }
                icon={<SVG name={'calendar'} color='#8692A3' size={16} />}
                className={style.customButton}
                disabled={!isSixSignalActivated}
              >
                {state?.selectedDate ? state.selectedDate : 'Select Report'}
              </Button>
            )}

            {state.dateSelectionVisibility && (
              <FaSelect
                options={getDateOptions()}
                onClickOutside={() =>
                  localDispatch({
                    type: VisitorReportActions.SET_DATE_SELECTION_VISIBILITY,
                    payload: false
                  })
                }
                allowSearch
                optionClickCallback={handleDateChange}
                placement='BottomRight'
                loadingState={reportDataLoading}
                allowSearchTextSelection={false}
              />
            )}
          </div>
        </div>
        <div className='mt-6'>
          {reportDataLoading ||
          state.shareData.loading ||
          currentProjectSettingsLoading ? (
            <div className='w-full h-full flex items-center justify-center'>
              <div className='w-full h-64 flex items-center justify-center'>
                <Spin size='large' />
              </div>
            </div>
          ) : (
            <>
              <ReportTable
                data={reportData?.result_group[0]}
                selectedChannel={state.selectedChannel}
                selectedCampaigns={state.selectedCampaigns}
                isSixSignalActivated={isSixSignalActivated}
                dataSelected={state.selectedDate}
              />
              {!!reportData && reportData.result_group?.[0]?.rows?.length > 0 && (
                <div className='text-right font-size--small'>
                  Logos provided by{' '}
                  <a
                    className='font-size--small'
                    href='https://www.uplead.com'
                    target='_blank'
                    rel='noreferrer'
                  >
                    UpLead
                  </a>
                </div>
              )}
            </>
          )}
        </div>
      </div>
      <SideDrawer drawerVisible={state.drawerVisible} hideDrawer={hideDrawer} />
      {state.shareModalVisibility && (
        <ShareModal
          visible={state.shareModalVisibility}
          onCancel={handleShareModalCancel}
          shareData={state.shareData.data ? state.shareData.data : undefined}
        />
      )}
    </div>
  );
};

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setShowAnalyticsResult
    },
    dispatch
  );
const connector = connect(null, mapDispatchToProps);
type VisitorIdentificationComponentProps = ConnectedProps<typeof connector>;

export default connector(SixSignalReport);
