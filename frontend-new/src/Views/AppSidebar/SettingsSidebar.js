import React, { useMemo } from 'react';
import cx from 'classnames';
import { useHistory, useLocation } from 'react-router-dom';
import {
  configureMenuItems,
  settingsMenuItems
} from 'Components/FaHeader/FaHeader';
import { SVG } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { isConfigurationUrl } from './appSidebar.helpers';
import SidebarMenuItem from './SidebarMenuItem';

const SettingItem = ({ item }) => {
  const location = useLocation();
  const history = useHistory();
  const { pathname } = location;

  const handleItemClick = () => {
    history.push(item.url);
  };

  const isActive = pathname === item.url;

  return (
    <div
      onClick={handleItemClick}
      role='button'
      className={cx(
        'py-2 cursor-pointer rounded-md pl-8 pr-2 flex justify-between col-gap-2 items-center',
        {
          [styles['active']]: isActive
        }
      )}
    >
      <SidebarMenuItem text={item.label} />
      {isActive && <SVG size={16} color='#595959' name='arrowright' />}
    </div>
  );
};

const SettingsSidebar = () => {
  const location = useLocation();
  const { pathname } = location;

  const menuList = useMemo(() => {
    if (isConfigurationUrl(pathname)) {
      return configureMenuItems;
    }
    return settingsMenuItems;
  }, [pathname]);

  return (
    <div className='flex flex-col row-gap-1 px-2'>
      {menuList.map((item) => {
        return <SettingItem item={item} />;
      })}
    </div>
  );
};

export default SettingsSidebar;
