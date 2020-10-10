import React from 'react';
import {
  Badge, Button
} from 'antd';
import { Text } from 'factorsComponents';
import { FullscreenOutlined } from '@ant-design/icons';

const Titles = [
  {
    title: 'Conversion Funnel',
    subTitle: 'User count grouped by Gender, City. Showing 5 of 20 groups'
  },
  {
    title: 'Conversion Funnel',
    subTitle: 'User count grouped by Gender, City. Showing 5 of 20 groups'
  },
  {
    title: 'Leads by First, Last and Most Engaged',
    subTitle: 'User count grouped by First, Last and Most Engaged.'
  },
  {
    title: 'Website Monitoring',
    subTitle: 'User count grouped by City, Gender.'
  }
];

function WidgetCard({ id, setwidgetModal }) {
  return (
        <div className={'fa-dashboard--widget-card'}>
            <div className={'fa-widget-card--top flex justify-between items-start'}>
                <div>
                    <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>{Titles[id].title}</Text>
                    <Text type={'paragraph'} mini color={'grey'} extraClass={'m-0'}>{Titles[id].subTitle}</Text>
                </div>
                <div className={'flex flex-col justify-start items-start fa-widget-card--top-actions'}>
                    <Button onClick={() => setwidgetModal(true)} icon={<FullscreenOutlined />} type="text" />
                </div>
            </div>
            <div className={'fa-widget-card--legend flex justify-center items-center'}>
                <Badge status="success" text="Add to Wishlist, Chennai" />
                <Badge status="warning" text="Add to Wishlist. Chennai" />
            </div>
            <div className={'fa-widget-card--visuals flex justify-center items-center'}>
                <img src={`../../assets/charts/chart-${id}.png`} />
            </div>
        </div>
  );
}

export default WidgetCard;
