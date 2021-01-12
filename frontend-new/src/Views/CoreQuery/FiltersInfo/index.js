import React from "react";
import DurationInfo from "../../../components/DurationInfo";

function FiltersInfo({ durationObj, breakdown, handleDurationChange }) {
  return null;
  return (
    <div className="flex items-center leading-4">
      <div className="mr-1">Data from </div>
      <DurationInfo
        durationObj={durationObj}
        handleDurationChange={handleDurationChange}
      />
      {breakdown.length ? (
        <div className="ml-1">shown as top 5 groups</div>
      ) : null}
    </div>
  );
}

export default FiltersInfo;
