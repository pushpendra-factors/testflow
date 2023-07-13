import React, { useMemo } from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import {
  settingsMenuItems,
  getConfigureMenuItems
} from 'Components/FaHeader/FaHeader';
import { isConfigurationUrl } from './appSidebar.helpers';
import SidebarMenuItem from './SidebarMenuItem';
import useAgentInfo from 'hooks/useAgentInfo';
import { WhiteListedAccounts } from 'Routes/constants';

const SettingItem = ({ item }) => {
  const location = useLocation();
  const history = useHistory();
  const { pathname } = location;

  const handleItemClick = () => {
    history.push(item.url);
  };

  const isActive = pathname === item.url;

  return (
    <SidebarMenuItem
      text={item.label}
      isActive={isActive}
      onClick={handleItemClick}
    />
  );
};

const SettingsSidebar = () => {
  const location = useLocation();
  const { pathname } = location;
  const { email } = useAgentInfo();

  const menuList = useMemo(() => {
    if (isConfigurationUrl(pathname)) {
      return getConfigureMenuItems(email);
    }
    return settingsMenuItems;
  }, [pathname, email]);

  return (
    <div className='flex flex-col row-gap-1 px-2'>
      {menuList.map((item) => {
        if (item?.whitelisted && !WhiteListedAccounts.includes(email)) {
          return null;
        }
        return <SettingItem item={item} />;
      })}
    </div>
  );
};

export default SettingsSidebar;
