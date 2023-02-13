import { Button, Collapse, Divider, Drawer } from 'antd';
import { Link } from 'react-router-dom';
import { Text, SVG } from 'Components/factorsComponents';
import FaPublicHeader from 'Components/FaPublicHeader';
import style from './index.module.scss';
import React, { useEffect, useState } from 'react';
import { useDispatch } from 'react-redux';
import { SHOW_ANALYTICS_RESULT } from 'Reducers/types';
import BrandInfo from './components/BrandInfo';
import Properties from './components/Properties';
import RecentActivity from './components/RecentActivity';
import QuickFilter from './components/QuickFilter';

const SixSignalReport = () => {
  const [drawerVisible, setDrawerVisible] = useState(false);
  const [filterValue, setFilterValue] = useState<undefined | string>(undefined);
  const dispatch = useDispatch();
  useEffect(() => {
    dispatch({ type: SHOW_ANALYTICS_RESULT, payload: true });
    return () => {
      dispatch({ type: SHOW_ANALYTICS_RESULT, payload: false });
    };
  }, [dispatch]);

  const showDrawer = () => {
    console.log('show drawer called', drawerVisible);
    setDrawerVisible(true);
  };

  const hideDrawer = () => {
    setDrawerVisible(false);
  };
  const drawerTitle = () => (
    <div className='flex gap-2 items-center'>
      <SVG name={'Eye'} size={24} />
      <Text
        type={'title'}
        level={6}
        weight={'bold'}
        color='grey-2'
        extraClass='mb-0'
      >
        Quickview
      </Text>
    </div>
  );
  const drawerFooter = () => (
    <div className='flex justify-center items-center p-4'>
      <Button
        style={{ width: '100%' }}
        onClick={() => console.log('see journey clicked')}
        size='large'
        type='primary'
      >
        See Full Journey
      </Button>
    </div>
  );
  const renderPanelHeader = (header: string) => (
    <div className='px-5'>
      <Text type={'title'} level={7} color='grey-2' weight={'bold'}>
        {header}
      </Text>
    </div>
  );
  return (
    <div className='flex flex-col'>
      <FaPublicHeader showDrawer={showDrawer} />
      <div className='px-24 pt-16 mt-12'>
        <div className='flex justify-between align-middle'>
          <div className='flex align-middle gap-2'>
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
                  <Link
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
                  </Link>
                </div>
              </div>
            </div>
          </div>
          <div>{/* match account */}</div>
        </div>
        <Divider />
        <div className='flex align-middle gap-2'>
          <div>{/* date picker */}</div>
          <QuickFilter
            filters={[
              { id: 'all', label: 'All' },
              { id: 'paid', label: 'Paid Search' },
              { label: 'Organic', id: 'organic' },
              { label: 'Direct', id: 'direct' }
            ]}
            onFilterChange={(id) => setFilterValue(id)}
            selectedFilter={filterValue}
          />
        </div>
      </div>

      <Drawer
        title={drawerTitle()}
        closable={true}
        onClose={hideDrawer}
        width={300}
        visible={drawerVisible}
        closeIcon={<SVG name={'Remove'} color='#8692A3' />}
        className={style.drawerStyle}
        headerStyle={{ background: '#F5F6F8' }}
        bodyStyle={{ padding: 0 }}
        footer={drawerFooter()}
      >
        <div className='flex flex-col'>
          <BrandInfo
            name='Wayne Enterprises'
            description='Subscription management software'
            logo='https://freestencilgallery.com/wp-content/uploads/2017/05/Wayne-Enterprises-Logo-Stencil-Thumb.jpg'
            links={[
              {
                source: 'https://cdn-icons-png.flaticon.com/512/174/174857.png',
                href: 'www.lindedin.com'
              }
            ]}
          />
          <Divider className={style.divider} />
          <Collapse
            defaultActiveKey={['0']}
            expandIconPosition={'right'}
            className='fa-six-signal-panel'
            ghost
          >
            <Collapse.Panel
              header={renderPanelHeader('Properties')}
              className='fa-six-signal-panel-item'
            >
              <Properties
                properties={[
                  { name: '6signal Name', value: 'Wayne Enterprises' },
                  { name: '6signal Domain', value: 'North America' },
                  { name: '6signal Name', value: 'Wayne Enterprises' },
                  { name: '6signal Domain', value: 'North America' }
                ]}
              />
            </Collapse.Panel>
          </Collapse>
          <Divider className={style.divider} />
          <Collapse
            expandIconPosition={'right'}
            className='fa-six-signal-panel'
            ghost
          >
            <Collapse.Panel
              header={renderPanelHeader('Engaged With')}
              className='fa-six-signal-panel-item'
            >
              <div className='flex justify-start flex-wrap gap-1 px-5 pb-0'>
                {[
                  'Webinar',
                  'Google Ads',
                  'Webinar',
                  'Google Ads',
                  'Website',
                  'E-book download',
                  'Website'
                ].map((item) => (
                  <div className={style.engagedContainer}>
                    <Text
                      type={'paragraph'}
                      mini
                      weight='bold'
                      extraClass='m-0'
                      color='grey'
                    >
                      {item}
                    </Text>
                  </div>
                ))}
              </div>
            </Collapse.Panel>
          </Collapse>
          <Divider className={style.divider} />

          <Collapse
            expandIconPosition={'right'}
            className='fa-six-signal-panel'
            ghost
          >
            <Collapse.Panel
              header={renderPanelHeader('Recent Activites')}
              className='fa-six-signal-panel-item'
            >
              <RecentActivity
                recentActivities={[
                  'www.factors.ai/',
                  'www.factors.ai/pricing',
                  'www.factors.ai/lp/leadfeeder',
                  'www.factors.ai/blog'
                ]}
              />
            </Collapse.Panel>
          </Collapse>
        </div>
      </Drawer>
    </div>
  );
};

export default SixSignalReport;
