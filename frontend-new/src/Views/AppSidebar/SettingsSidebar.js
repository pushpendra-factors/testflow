import React, { useMemo } from 'react';
import { useHistory, useLocation } from 'react-router-dom';
import {
  settingsMenuItems,
  getConfigureMenuItems,
  settingsCategorisedMap
} from 'Components/FaHeader/FaHeader';
import { WhiteListedAccounts } from 'Routes/constants';
import { useSelector } from 'react-redux';
import { PathUrls } from 'Routes/pathUrls';
import SidebarMenuItem from './SidebarMenuItem';
import { checkMatchPath, isConfigurationUrl } from './appSidebar.helpers';

const SettingItem = ({ item }) => {
  const location = useLocation();
  const history = useHistory();
  const { pathname } = location;

  const handleItemClick = () => {
    history.push(item.url);
  };
  const isActive =
    item.url === PathUrls.SettingsIntegration
      ? pathname === item.url ||
        checkMatchPath(pathname, PathUrls.SettingsIntegrationURLID)
      : pathname === item.url;

  return (
    <SidebarMenuItem
      text={item.label}
      isActive={isActive}
      icon={item.icon}
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
      return getConfigureMenuItems(activeAgent);
    }
    return settingsMenuItems;
  }, [pathname, activeAgent]);

  return (
    <div className='flex flex-col gap-y-1 px-2'>
      {settingsCategorisedMap(activeAgent).map((item) => {
        if (item?.whitelisted && !WhiteListedAccounts.includes(activeAgent)) {
          return null;
        }
        return (
          <>
            <SettingItem item={item} />
            <div className={`border-bottom--thin-2`}></div>
            {item.items.map((subItem) => {
              return <SettingItem item={subItem} />;
            })}
          </>
        );
      })}
    </div>
  );
};

export default SettingsSidebar;
