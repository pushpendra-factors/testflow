import React from 'react';
import { Button, Collapse, Divider, Drawer } from 'antd';
import { Text, SVG } from 'Components/factorsComponents';
import BrandInfo from '../BrandInfo';
import Properties from '../Properties';
import RecentActivity from '../RecentActivity';
import style from './index.module.scss';

const SideDrawer = ({ hideDrawer, drawerVisible }: Props) => {
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
            key='properties'
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
            key='engagement'
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
            key='activities'
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
  );
};

type Props = {
  hideDrawer: () => void;
  drawerVisible: boolean;
};

export default SideDrawer;
