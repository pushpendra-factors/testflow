import React from 'react';
import { Select } from 'antd';
import styles from '../../../components/SaveQuery/index.module.scss';

function CampaignMetricsDropdown({ metrics, currValue, onChange }) {
  return (
    <div className="flex items-center gap-x-2">
      <div>Show</div>
      <div>
        <Select
          onChange={onChange}
          className={styles.singleSelectStyles}
          value={currValue}
        >
          {metrics.map((d, index) => (
            <Select.Option value={index} key={d}>
              {d}
            </Select.Option>
          ))}
        </Select>
      </div>
      <div>from</div>
    </div>
  );
}

export default CampaignMetricsDropdown;
