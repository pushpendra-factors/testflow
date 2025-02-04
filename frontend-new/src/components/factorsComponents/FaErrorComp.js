import React from 'react';
import lazyWithRetry from 'Utils/lazyWithRetry';
// import Lottie from 'react-lottie';
import * as Sentry from '@sentry/react';
import { Button } from 'antd';
import { Text } from '.';
import animationData from '../../assets/lottie/38064-error-cone.json';

const Lottie = lazyWithRetry(
  () => import(/* webpackChunkName: "animation" */ 'react-lottie')
);

const defaultOptions = {
  loop: true,
  autoplay: true,
  animationData,
  rendererSettings: {
    preserveAspectRatio: 'xMidYMid slice'
  }
};

const FaErrorLog = (error, info) => {
  console.log('Factors error log:', error);
  console.log('error info', info);
  Sentry && Sentry.captureException(error);
};

const FaErrorComp = ({ size, className, type, title, subtitle }) => {
  const sizeCal = (size) => {
    switch (size) {
      case 'large':
        return 200;
      case 'medium':
        return 150;
      case 'small':
        return 100;
      default:
        return 100;
    }
  };
  const finalSize = sizeCal(size);

  if (title) {
    window.Intercom &&
      window.Intercom(
        'showNewMessage',
        `Hey, got ${title}! Can you guys help me out?`
      );
  }

  const refreshPage = () => {
    window.location = '/';
  };

  return (
    <div
      className={`w-full flex flex-col justify-center items-center ${className}`}
    >
      <Lottie options={defaultOptions} height={finalSize} width={finalSize} />
      {title && (
        <Text
          type='title'
          align='center'
          level={5}
          weight='bold'
          extraClass='ml-2 m-0'
        >
          {title}
        </Text>
      )}
      {subtitle && (
        <Text
          type='title'
          align='center'
          level={7}
          weight='thin'
          color='grey'
          extraClass='ml-2 m-0'
        >
          {subtitle}
        </Text>
      )}
      <Button size='large' className='mt-4' onClick={() => refreshPage()}>
        Try Again!
      </Button>
    </div>
  );
};

export { FaErrorComp, FaErrorLog };
