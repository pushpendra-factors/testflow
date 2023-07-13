import React, { ComponentType } from 'react';
import { FEATURES } from '../constants/plans.constants';
import { Spin } from 'antd';
import useFeatureLock from 'hooks/useFeatureLock';

interface HOCProps {
  featureName: typeof FEATURES[keyof typeof FEATURES];
  LockedComponent: React.FC;
}

const LoaderComponent = () => (
  <div className='w-full h-full flex items-center justify-center'>
    <div className='w-full h-64 flex items-center justify-center'>
      <Spin size='large' />
    </div>
  </div>
);

const withFeatureLockHOC = <P extends object>(
  WrappedComponent: ComponentType<P>,
  hocProps: HOCProps
) => {
  const FeatureLockWrapper = (props: any) => {
    const { featureName, LockedComponent } = hocProps;
    const { isFeatureLocked, isLoading } = useFeatureLock(featureName);

    if (isLoading) return <LoaderComponent />;
    if (isFeatureLocked) return <LockedComponent />;

    return <WrappedComponent {...props} />;
  };

  return FeatureLockWrapper;
};

export default withFeatureLockHOC;
