import React, { ReactElement, ReactNode, useEffect, useState } from 'react';
import cx from 'classnames';
import { SVG, Text } from 'Components/factorsComponents';
import { useHistory, useLocation } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import useQuery from 'hooks/useQuery';
import { useSelector } from 'react-redux';
import { featureLock } from 'Routes/feature';
import { isAlertsUrl } from './appSidebar.helpers';
import styles from './index.module.scss';

interface AppContentHeaderProps {
  heading: ReactNode;
  actions: ReactNode;
}
export const AppContentHeader = ({
  heading,
  actions
}: AppContentHeaderProps) => (
  <div className={cx('flex justify-between w-full border-b bg-white mb-6 p-4')}>
    {heading}

    {actions}
  </div>
);
