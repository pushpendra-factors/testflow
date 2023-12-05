import React from 'react';
import { Select, Tabs } from 'antd';
import styles from './index.module.scss';
const { TabPane } = Tabs;
import { Number as NumFormat } from 'factorsComponents';

function CampaignMetricsDropdown({ metrics, currValue, setCurrMetricsValue, metricsValue }) {
  return (
    <div className='flex flex-row ml-4 -mt-2'>
      {metrics.map((d, index) => (
        <>
          <div className='basis-1/2 m-4 pt-4'>
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
                  {d}
                </p>
                <p
                  className={`${
                    currValue === index ? styles.text2Active : styles.text2
                  }`}
                >
                  {/* {Math.trunc(metricsValue?.[index]*100)/100} */}
                  <NumFormat number={metricsValue?.[index]} />
                </p>
              </div>
            </div>
          </div>
        </>
      ))}
    </div>
  );
}

export default CampaignMetricsDropdown;
