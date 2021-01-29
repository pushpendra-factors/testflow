import React from "react";
import { Select } from "antd";
import styles from '../../../components/SaveQuery/index.module.scss';

function CampaignMetricsDropdown({ metrics, currValue, onChange }) {
  return (
    <div className="mr-2 flex items-center">
      <div className="mr-2">Show</div>
      <div>
        <Select
          onChange={onChange}
          className={styles.singleSelectStyles}
          value={currValue}
        >
          {metrics.map((d, index) => {
            return (
              <Select.Option value={index} key={index}>
                {d}
              </Select.Option>
            );
          })}
        </Select>
      </div>
      <div className="ml-2">from</div>
    </div>
  );
}

export default CampaignMetricsDropdown;
