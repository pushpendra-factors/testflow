import React from "react";
import { Spin } from "antd";

function PageSuspenseLoader() {
  return (
    <div className="flex items-center min-h-screen justify-center w-min-screen">
      <Spin size={"large"} className={"fa-page-loader"} />
    </div>
  );
}

export default PageSuspenseLoader;
