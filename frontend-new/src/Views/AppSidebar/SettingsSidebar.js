import React, { useMemo } from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import {
  configureMenuItems,
  settingsMenuItems
} from 'Components/FaHeader/FaHeader';
import { isConfigurationUrl } from './appSidebar.helpers';
import SidebarMenuItem from './SidebarMenuItem';
import { WhiteListedAccounts } from 'Routes/constants'; 
import { useSelector } from 'react-redux';

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

  const agentState = useSelector((state) => state.agent);
  const activeAgent = agentState?.agent_details?.email;

  const menuList = useMemo(() => {
    if (isConfigurationUrl(pathname)) {
      return configureMenuItems;
    }
    return settingsMenuItems;
  }, [pathname]);

  return (
    <div className='flex flex-col row-gap-1 px-2'>
      {menuList.map((item) => {
        if(item?.whitelisted && !WhiteListedAccounts.includes(activeAgent)){
          return null
        }
        return <SettingItem item={item} />;
      })}
    </div>
  );
};

export default SettingsSidebar;
