import React, { useState, useEffect } from 'react';
import {
  Badge, Button, Col
} from 'antd';
import { Text } from 'factorsComponents';
import { FullscreenOutlined, RightOutlined, LeftOutlined } from '@ant-design/icons';

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

function WidgetCard({ id, setwidgetModal, widthSize = 2 }) {
  const [currentWidth, setCurrentWidth] = useState(widthSize);

  const resizeWidth = (operator) => {
    console.log('currentWidth', currentWidth);
    if (operator === '+') {
    // console.log("increment");
      if (currentWidth !== 3) {
        setCurrentWidth(currentWidth + 1);
      } else {
        return 3;
      }
    } else {
    // console.log("decrement");
      if (currentWidth !== 0) {
        setCurrentWidth((currentWidth - 1 === 0) ? 1 : currentWidth - 1);
      } else {
        return 1;
      }
    }
  };
  const calcWidth = (size) => {
    // console.log("calcWidth",size);
    switch (size) {
      case 1: return 6;
      case 2: return 12;
      case 3: return 24;
      default: return 12;
    }
  };
  useEffect(() => {
    calcWidth(currentWidth);
  });

  return (
        <Col span={calcWidth(currentWidth)} style={{ transition: 'all 0.1s' }}>
          <div className={'fa-dashboard--widget-card'}>
            <div className={'fa-widget-card--resize-container'}>
              <span className={'fa-widget-card--resize-contents'}>
              {currentWidth < 3 && <a onClick={() => resizeWidth('+')}><RightOutlined /></a>}
                {currentWidth > 1 && <a onClick={() => resizeWidth('-')}><LeftOutlined /></a> }

              </span>
            </div>
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
        </Col>
  );
}

export default WidgetCard;
