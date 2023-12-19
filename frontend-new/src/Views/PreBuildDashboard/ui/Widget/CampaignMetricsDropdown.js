import React from 'react';
import { Select, Spin, Tabs } from 'antd';
import styles from './index.module.scss';
const { TabPane } = Tabs;
import { Number as NumFormat } from 'factorsComponents';
import { getFormattedKpiValue } from 'Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';

function CampaignMetricsDropdown({
  metrics,
  currValue,
  setCurrMetricsValue,
  metricsValue
}) {
  if (!metricsValue?.length) {
    return (
      <div className='flex justify-center items-center w-full h-full'>
        <Spin size='small' />
      </div>
    );
  }
  return (
    <div className='flex flex-row ml-4'>
      {metrics.map((metric, index) => (
        <>
          <div className='basis-1/2 mx-4 pt-4'>
            <div
              className={`${styles.container}`}
              onClick={() => setCurrMetricsValue(index)}
            >
              <div>
                <p
                  className={`${
                    currValue === index ? styles.text1Active : styles.text1
                  }`}
                >
                  {metric?.label}
                </p>
                <p
                  className={`${
                    currValue === index ? styles.text2Active : styles.text2
                  }`}
                >
                  {metric?.metricType != null && metric?.metricType !== '' ? (
                    getFormattedKpiValue({
                      value: metricsValue?.[index],
                      metricType: metric?.metricType
                    })
                  ) : (
                    <NumFormat number={metricsValue?.[index]} />
                  )}
                </p>
              </div>
              {index != metricsValue?.length - 1 && (
                <div className={`${styles.line}`}></div>
              )}
            </div>
          </div>
        </>
      ))}
    </div>
  );
}

export default CampaignMetricsDropdown;
