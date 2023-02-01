import Icon from '@ant-design/icons/lib/components/Icon';
import { SVG, Text } from 'Components/factorsComponents';
import React from 'react';
const DemoSDK = () => {
  return (
    <div
      style={{
        display: 'grid',
        placeContent: 'center',
        marginTop: '110px',
        textAlign: 'center'
      }}
    >
      <div style={{ margin: '0 auto ' }}>
        {' '}
        <img src="https://s3.amazonaws.com/www.factors.ai/assets/img/product/JSSDK_Demoproject.svg" alt="JSSDK_Demoproject" />
      </div>

      <Text
        type={'title'}
        level={5}
        weight={'bold'}
        color={'#3E516C'}
        extraClass={'m-0 mt-2'}
      >
        You are still inside the demo project
      </Text>
      <Text type={'title'} level={5} color={'#3E516C'} extraClass={'m-0 mt-2'}>
        To access your javascript SDK, jump into your main project using the
        account card on the top right. We're rooting for you!
      </Text>
    </div>
  );
};
export default DemoSDK;
