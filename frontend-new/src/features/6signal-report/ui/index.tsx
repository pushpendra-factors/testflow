import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Button, Divider, notification, Spin } from 'antd';
import { Link, useHistory } from 'react-router-dom';
import { Text, SVG } from 'Components/factorsComponents';
import FaPublicHeader from 'Components/FaPublicHeader';
import style from './index.module.scss';
import { useDispatch, useSelector } from 'react-redux';
import { SHOW_ANALYTICS_RESULT } from 'Reducers/types';

import QuickFilter from './components/QuickFilter';
import SideDrawer from './components/SideDrawer';
import {
  generateFirstAndLastDayOfLastWeeks,
  getFormattedRange,
  getPublicUrl,
  parseResultGroupResponse
} from '../utils';
import FaSelect from 'Components/FaSelect';
import {
  ShareApiResponse,
  WeekStartEnd,
  ShareData,
  ReportApiResponse,
  ReportApiResponseData
} from '../types';
import ReportTable from './components/ReportTable';
import { CHANNEL_QUICK_FILTERS, SHARE_QUERY_PARAMS } from '../const';
import {
  getSixSignalReportData,
  getSixSignalReportPublicData,
  shareSixSignalReport
} from '../state/services';
import useAgentInfo from 'hooks/useAgentInfo';
import ShareModal from './components/ShareModal';
import useQuery from 'hooks/useQuery';

const SixSignalReport = () => {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [filterValue, setFilterValue] = useState<string>(
    CHANNEL_QUICK_FILTERS[0].id
  );
  const [data, setData] = useState<ReportApiResponseData | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [campaigns, setCampaigns] = useState<string[]>([]);
  const [isCampaignSelectVisible, setIsCampaignSelectVisible] = useState(false);
  const [seletedCampaigns, setSelectedCampaigns] = useState([]);
  const [dateSelected, setDateSelected] = useState<string>('');
  const [isDateSelectionOpen, setIsDateSelectionOpen] =
    useState<boolean>(false);
  const [pageMode, setPageMode] = useState<'in-app' | 'public'>('in-app');
  const [shareModalVisibility, setShareModalVisibility] =
    useState<boolean>(false);
  const [loadingShareData, setLoadingShareData] = useState<boolean>(false);
  const [shareData, setShareData] = useState<ShareData | null>(null);
  const { isLoggedIn, email } = useAgentInfo();
  const { active_project, currentProjectSettings } = useSelector(
    (state: any) => state.global
  );
  const routerQuery = useQuery();
  const history = useHistory();
  const paramQueryId = routerQuery.get(SHARE_QUERY_PARAMS.queryId);
  const paramProjectId = routerQuery.get(SHARE_QUERY_PARAMS.projectId);

  const isSixSignalActivated = currentProjectSettings
    ? !currentProjectSettings?.int_factors_six_signal_key &&
      !currentProjectSettings?.int_client_six_signal_key
      ? false
      : true
    : true;

  const showDrawer = () => {
    setDrawerVisible(true);
  };

  const hideDrawer = () => {
    setDrawerVisible(false);
  };

  const dateValues = useMemo(() => generateFirstAndLastDayOfLastWeeks(5), []);

  const dispatch = useDispatch();

  const handleQuickFilterChange = (filterId: string) => {
    setFilterValue(filterId);
  };

  const getCampaignOptions = () => {
    return campaigns.map((campaign) => [campaign]);
  };

  const getDateOptions = () => {
    return dateValues.map((date) => [
      date.formattedRangeOption,
      date.formattedRange
    ]);
  };

  const handleApplyClick = (val) => {
    setSelectedCampaigns(val.map((vl) => JSON.parse(vl)[0]));
  };

  const renderCampaignText = () => {
    const text = seletedCampaigns?.join(', ');
    return text?.length > 40 ? `${text.slice(0, 40)}...` : text;
  };

  const getDateObjFromSelectedDate = useCallback(() => {
    const dateObj: WeekStartEnd | undefined = dateValues.find(
      (date) => date.formattedRange === dateSelected
    );
    return dateObj;
  }, [dateSelected, dateValues]);

  const handleShareClick = async () => {
    try {
      //checking if share data is already fetched for the dates
      if (dateSelected === shareData?.dateSelected) {
        setShareModalVisibility(true);
        return;
      }
      setLoadingShareData(true);
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
      setLoadingShareData(false);
      if (res.data) {
        setShareData({
          ...res?.data,
          dateSelected,
          publicUrl: getPublicUrl(res.data)
        });
        setShareModalVisibility(true);
      } else {
        console.error('No data found to share', res?.data);
      }
    } catch (error) {
      console.error('Error in sharing report', error);
      notification.error({
        message: 'Error',
        description: error?.data?.error || 'Something went wrong',
        duration: 5
      });
      setLoadingShareData(false);
    }
  };

  const handleShareModalCancel = () => {
    setShareModalVisibility(false);
    setLoadingShareData(false);
  };

  const handleDateChange = (option: string[]) => {
    setDateSelected(option[1]);
    setIsDateSelectionOpen(false);
    //resetting campaigns to null
    setCampaigns([]);
    setSelectedCampaigns([]);
  };

  useEffect(() => {
    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
    return () => {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    };
  }, [dispatch]);

  useEffect(() => {
    if (!isSixSignalActivated) {
      setLoading(false);
    }
  }, [isSixSignalActivated]);

  //TODO: Remove the below useEffect when 6 signal report is accessible to all
  useEffect(() => {
    if (isLoggedIn && email !== 'solutions@factors.ai') {
      history.push('/');
    }
  }, [isLoggedIn, email]);

  useEffect(() => {
    const fetchPublicData = async () => {
      try {
        if (!isLoggedIn) setPageMode('public');
        setLoading(true);
        if (paramQueryId && paramProjectId) {
          const res = (await getSixSignalReportPublicData(
            paramProjectId,
            paramQueryId
          )) as ReportApiResponse;

          setLoading(false);
          if (res?.data?.[1]?.result_group) {
            setData(res?.data?.[1]);
            const _query = res?.data?.[1].query?.six_signal_query_group?.[0];
            if (_query.fr && _query.to) {
              setDateSelected(
                getFormattedRange(_query.fr, _query.to, _query.tz)
              );
            }
          }
        }
      } catch (error) {
        console.error('Error in fetching public data', error);
        notification.error({
          message: 'Error',
          description: error?.data?.error || 'Something went wrong',
          duration: 5
        });
        setLoading(false);
        setData(null);
      }
    };
    if (paramQueryId && paramProjectId) fetchPublicData();
    //navigating back to login page if required parameters are not there
    if (!isLoggedIn && (!paramProjectId || !paramQueryId)) {
      history.push('/login');
    }
  }, [paramQueryId, paramProjectId, isLoggedIn]);

  useEffect(() => {
    if (isLoggedIn && isSixSignalActivated && dateValues && !paramQueryId) {
      setDateSelected(dateValues[0].formattedRange);
    }
  }, [isLoggedIn, dateValues, isSixSignalActivated, paramQueryId]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        if (dateSelected) {
          const dateObj = getDateObjFromSelectedDate();
          if (!dateObj) return;
          const res = (await getSixSignalReportData(
            active_project.id,
            dateObj?.from,
            dateObj?.to,
            active_project?.time_zone || 'Asia/Kolkata'
          )) as ReportApiResponse;
          setLoading(false);
          if (res?.data?.[1]?.result_group) {
            setData(res?.data?.[1]);
          }
        }
      } catch (error) {
        console.error('Error in fetching data', error);
        setLoading(false);
        setData(null);
      }
    };
    if (active_project && active_project?.id && dateSelected) fetchData();
  }, [active_project, dateSelected, dateValues, getDateObjFromSelectedDate]);

  useEffect(() => {
    if (data) {
      const value = parseResultGroupResponse(data.result_group[0]);
      setCampaigns(value.campaigns);
    }
  }, [data]);

  return (
    <div className='flex flex-col'>
      <FaPublicHeader
        showDrawer={showDrawer}
        handleShareClick={handleShareClick}
      />
      <div className='px-24 pt-16 mt-12'>
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
          <div>{/* match account */}</div>
        </div>
        <Divider />
        <div className='flex align-middle gap-4'>
          <div className={style.filter}>
            <Button
              onClick={() => setIsDateSelectionOpen(true)}
              icon={<SVG name={'calendar'} color='#8692A3' size={16} />}
              className={style.customButton}
              disabled={pageMode === 'public' || !isSixSignalActivated}
            >
              {dateSelected ? dateSelected : 'Select Report'}
            </Button>

            {isDateSelectionOpen && (
              // @ts-ignore
              <FaSelect
                options={getDateOptions()}
                onClickOutside={() => setIsDateSelectionOpen(false)}
                allowSearch
                optionClick={(option: string[]) => handleDateChange(option)}
              />
            )}
          </div>
          <QuickFilter
            filters={CHANNEL_QUICK_FILTERS}
            onFilterChange={handleQuickFilterChange}
            selectedFilter={filterValue}
          />
          <div className={style.filter}>
            <Button
              className={style.customButton}
              onClick={() => setIsCampaignSelectVisible(true)}
              icon={<SVG name={'Filter'} color='#8692A3' size={12} />}
            >
              {!seletedCampaigns || !seletedCampaigns?.length
                ? 'Filter by campaign'
                : renderCampaignText()}
            </Button>

            {isCampaignSelectVisible && (
              // @ts-ignore
              <FaSelect
                options={getCampaignOptions()}
                onClickOutside={() => setIsCampaignSelectVisible(false)}
                applClick={handleApplyClick}
                selectedOpts={seletedCampaigns}
                allowSearch={campaigns?.length > 0}
                multiSelect={campaigns?.length > 0}
              >
                {!campaigns?.length ? (
                  <div className='px-2'>
                    <Text type={'title'} level={7} extraClass={'m-0'}>
                      No Campaigns Found!
                    </Text>
                  </div>
                ) : null}
              </FaSelect>
            )}
          </div>
        </div>
        <div className='mt-6'>
          {loading || loadingShareData ? (
            <div className='w-full h-full flex items-center justify-center'>
              <div className='w-full h-64 flex items-center justify-center'>
                <Spin size='large' />
              </div>
            </div>
          ) : (
            <>
              <ReportTable
                data={data?.result_group[0]}
                selectedChannel={filterValue}
                selectedCampaigns={seletedCampaigns}
                isSixSignalActivated={isSixSignalActivated}
                dataSelected={dateSelected}
              />
              {!!data && (
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
      <SideDrawer drawerVisible={drawerVisible} hideDrawer={hideDrawer} />
      {shareModalVisibility && (
        <ShareModal
          visible={shareModalVisibility}
          onCancel={handleShareModalCancel}
          shareData={shareData ? shareData : undefined}
        />
      )}
    </div>
  );
};

export default SixSignalReport;
