import React, { useCallback } from 'react';
import { useSelector } from 'react-redux';
import cx from 'classnames';
import { Button, Dropdown, Menu } from 'antd';
import { selectGroupsList } from 'Reducers/groups/selectors';
import FiltersBox from './FiltersBox';
import styles from './index.module.scss';
import { Text, SVG } from 'Components/factorsComponents';
import { INITIAL_FILTERS_STATE } from './accountProfiles.constants';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';

function PropertyFilter({
  viewMode,
  filtersLimit = 3,
  profileType,
  source,
  filters = [],
  applyFilters,
  filtersExpanded,
  setFiltersExpanded,
  filtersList,
  listEvents,
  setListEvents,
  setFiltersList,
  appliedFilters,
  eventProp,
  setSaveSegmentModal,
  selectedAccount,
  setSelectedAccount,
  setAppliedFilters,
  setEventProp
}) {
  const groupsList = useSelector((state) => selectGroupsList(state));
  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);

  const handleAccountChange = (account) => {
    setSelectedAccount((current) => {
      return {
        ...current,
        account
      };
    });
    setFiltersList(INITIAL_FILTERS_STATE.filters);
    setEventProp(INITIAL_FILTERS_STATE.eventProp);
    setListEvents(INITIAL_FILTERS_STATE.eventsList);
    setAppliedFilters(INITIAL_FILTERS_STATE);
  };

  const analyseMenu = (
    <Menu className={styles['dropdown-menu']}>
      {groupsList.map((elem) => {
        return (
          <Menu.Item
            className={styles['dropdown-menu-item']}
            onClick={() => handleAccountChange(elem)}
            key={elem[1]}
          >
            <Text type='title' extraClass='mb-0'>
              {elem[0]}
            </Text>
          </Menu.Item>
        );
      })}
    </Menu>
  );

  const toggleFilters = useCallback(() => {
    setFiltersExpanded((curr) => !curr);
  }, [setFiltersExpanded]);

  if (filtersExpanded === false && newSegmentMode === false) {
    if (appliedFilters.filters.length > 0) {
      return (
        <Button
          className={cx(
            'flex items-center justify-center col-gap-1',
            styles['collapse-button']
          )}
          type='text'
          onClick={toggleFilters}
        >
          <Text type='title' extraClass='mb-0' weight='medium' color='grey-6'>
            View {appliedFilters.filters.length} filter(s)
          </Text>
          <SVG size={16} name='chevronDown' color='#8C8C8C' />
        </Button>
      );
    }

    return (
      <Button
        className={cx(
          'flex items-center justify-center col-gap-1',
          styles['filter-button']
        )}
        onClick={toggleFilters}
      >
        <SVG size={16} name='filter' color='#8C8C8C' />
        <Text
          type='title'
          extraClass='mb-0'
          weight='medium'
          color='character-primary'
        >
          Filter
        </Text>
      </Button>
    );
  }

  if (selectedAccount.account == null) return null;

  return (
    <div className='flex flex-col row-gap-4 w-full'>
      <ControlledComponent controller={newSegmentMode === false}>
        <Button
          className={cx(
            'flex items-center justify-center col-gap-1',
            styles['collapse-button']
          )}
          type='text'
          onClick={toggleFilters}
        >
          <Text type='title' extraClass='mb-0' weight='medium' color='grey-6'>
            Collapse conditions
          </Text>
          <SVG size={16} name='chevronDown' color='#8C8C8C' />
        </Button>
      </ControlledComponent>
      <div className='flex col-gap-2 items-center'>
        <Text type='title' extraClass='mb-0'>
          Include
        </Text>
        <Dropdown overlay={analyseMenu}>
          <div className='flex items-center col-gap-1'>
            <Text
              level={6}
              color='character-primary'
              weight='bold'
              type='title'
              extraClass='mb-0'
            >
              {selectedAccount.account[0]}
            </Text>{' '}
            <SVG size={16} name='caretDown' color='#8c8c8c' />
          </div>
        </Dropdown>
      </div>
      <FiltersBox
        source={selectedAccount.account[1]}
        filtersList={filtersList}
        profileType={profileType}
        setFiltersList={setFiltersList}
        appliedFilters={appliedFilters}
        applyFilters={applyFilters}
        setSaveSegmentModal={setSaveSegmentModal}
        listEvents={listEvents}
        setListEvents={setListEvents}
        eventProp={eventProp}
        setEventProp={setEventProp}
        onCancel={toggleFilters}
      />
    </div>
  );
}
export default PropertyFilter;
