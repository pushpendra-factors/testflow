import React from "react";
import styles from "./index.module.scss";

function ChartHeader({ total, query, bgColor, smallFont = false }) {
  return (
    <div className="flex flex-col items-center justify-center">
      <div className={`flex items-center ${smallFont ? "mb-2" : "mb-4"}`}>
        <div
          style={{ backgroundColor: bgColor }}
          className={`mr-1 ${styles.eventCircle}`}
        ></div>
        <div className={styles.eventText}>
          {query.length > 20 ? query.slice(0, 20) + "..." : query}
        </div>
      </div>
      <div
        className={`${smallFont ? styles.smallerTotalText : styles.totalText}`}
      >
        {total}
      </div>
    </div>
  );
}

export default ChartHeader;
