import React, { useState } from 'react';
import { SVG } from 'factorsComponents';
import { Button, Popover } from 'antd';

function ChartConfigPopover({children}) {
  const [visible, setVisible] = useState(false);

  return (
    <Popover
      onVisibleChange={setVisible}
      placement='bottomRight'
      trigger='click'
      content={children}
      visible={visible}
    >
      <Button
        onClick={setVisible?.bind(null, true)}
        size='large'
        icon={<SVG name='controls' />}
        type='text'
      />
    </Popover>
  );
}

export default ChartConfigPopover;
