import React, { useCallback } from "react";
import { Button } from "antd";

function CampaignsQueryComposer({ runCampaignsQuery }) {
  const handleRunQuery = useCallback(() => {
    runCampaignsQuery(false);
  }, [runCampaignsQuery]);

  return (
    <Button size={"large"} type="primary" onClick={handleRunQuery}>
      Analyse
    </Button>
  );
}

export default CampaignsQueryComposer;
