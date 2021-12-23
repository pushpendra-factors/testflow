import React from "react";
import styles from "./index.module.scss";
import { Number as NumFormat } from "../factorsComponents";
import {  useSelector } from 'react-redux';

function ChartHeader({ total, query, bgColor, smallFont = false }) {

  const { eventNames } = useSelector((state) => state.coreQuery);

  const displayQueryName = (q) => {
    return eventNames[q] || q;
  };

  return (
    <div className="flex flex-col items-center justify-center">
      <div className={`flex items-center ${smallFont ? "mb-2" : "mb-4"}`}>
        <div
          style={{ backgroundColor: bgColor }}
          className={`mr-1 ${styles.eventCircle}`}
        ></div>
        <div className={styles.eventText}>
          {query.length > 20 ? displayQueryName(query).slice(0, 20) + "..." : displayQueryName(query)}
        </div>
      </div>
      <div
        className={`${smallFont ? styles.smallerTotalText : styles.totalText}`}
      >
        <NumFormat shortHand={total > 10000} number={total} />
      </div>
    </div>
  );
}

export default ChartHeader;
