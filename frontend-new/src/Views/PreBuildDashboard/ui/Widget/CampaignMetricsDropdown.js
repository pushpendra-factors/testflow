import React from 'react';
import { Select, Tabs } from 'antd';
import styles from './index.module.scss';
const { TabPane } = Tabs;

function CampaignMetricsDropdown({ metrics, currValue, onChange }) {
  const data = [1116, 8189, 209.4642];
  return (
    <div className='flex flex-row ml-4 -mt-4'>
      {metrics.map((d, index) => (
        <>
          <div className='basis-1/2 m-4 pt-4'>
            <div
              className={`${styles.container}`}
              onClick={() => onChange(index)}
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
                  {data[index]}
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
