import React, { memo } from 'react';
import styles from './index.module.scss';
import { Text, SVG } from '../factorsComponents';
import { Dropdown, Menu } from 'antd';

const AxisOptionsContainer = ({
  xAxisOptions,
  yAxisOptions,
  onXAxisOptionChange,
  onYAxisOptionChange,
  xAxisMetric,
  yAxisMetric,
  visiblePointsCount,
}) => {
  const xAxisMenu = (
    <Menu>
      {xAxisOptions.map((item) => {
        return (
          <Menu.Item key={item.value} onClick={onXAxisOptionChange}>
            <div className={'flex items-center'}>
              <span className='mr-3'>{item.title}</span>
            </div>
          </Menu.Item>
        );
      })}
    </Menu>
  );

  const yAxisMenu = (
    <Menu>
      {yAxisOptions.map((item) => {
        return (
          <Menu.Item key={item.value} onClick={onYAxisOptionChange}>
            <div className={'flex items-center'}>
              <span className='mr-3'>{item.title}</span>
            </div>
          </Menu.Item>
        );
      })}
    </Menu>
  );

  return (
    <div className={`flex items-center mt-2 ${styles.axisOptionsContainer}`}>
      <div className='w-1/5 py-3 text-center first-col'>
        <Text
          weight='bold'
          color='grey-2'
          extraClass={`mb-0 text-sm`}
          type='title'
        >
          Metrics to Plot
        </Text>
      </div>
      <div className='w-3/5 py-3 pl-8 second-col flex'>
        <div className='w-1/2 xAxisOptions flex items-center'>
          <Text
            weight='medium'
            color='grey-2'
            extraClass={`mb-0 text-sm mr-2`}
            type='title'
          >
            X-axis:
          </Text>
          <Dropdown placement='topCenter' overlay={xAxisMenu}>
            <div className='flex items-center cursor-pointer'>
              <Text
                weight='medium'
                type='title'
                color='grey-2'
                extraClass='mb-0 text-sm'
              >
                {xAxisMetric.length <= 25
                  ? xAxisMetric
                  : xAxisMetric.substr(0, 25) + '...'}
              </Text>
              <SVG name={'dropdown'} size={25} color='#3E516C' />
            </div>
          </Dropdown>
        </div>
        <div className='w-1/2 yAxisOptions flex items-center'>
          <Text
            weight='medium'
            color='grey-2'
            extraClass={`mb-0 text-sm mr-2`}
            type='title'
          >
            Y-axis:
          </Text>
          <Dropdown placement='topCenter' overlay={yAxisMenu}>
            <div className='flex items-center cursor-pointer'>
              <Text
                weight='medium'
                type='title'
                color='grey-2'
                extraClass='mb-0 text-sm'
              >
                {yAxisMetric.length <= 25
                  ? yAxisMetric
                  : yAxisMetric.substr(0, 25) + '...'}
              </Text>
              <SVG name={'dropdown'} size={25} color='#3E516C' />
            </div>
          </Dropdown>
        </div>
      </div>
      <div className='w-1/5 py-3 text-center'>
        <Text
          weight='medium'
          color='grey-2'
          extraClass={`mb-0 text-sm`}
          type='title'
        >
          Showing {visiblePointsCount}{' '}
          {visiblePointsCount > 1 ? 'points' : 'point'}
        </Text>
      </div>
    </div>
  );
};

export default memo(AxisOptionsContainer);
