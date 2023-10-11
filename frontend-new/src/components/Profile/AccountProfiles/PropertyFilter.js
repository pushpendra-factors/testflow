import React, { useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button, Dropdown, Menu } from 'antd';
import FiltersBox from './FiltersBox';
import styles from './index.module.scss';
import { Text, SVG } from 'Components/factorsComponents';
import { INITIAL_FILTERS_STATE } from './accountProfiles.constants';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { setNewSegmentModeAction } from 'Reducers/accountProfilesView/actions';

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
  setEventProp,
  areFiltersDirty,
  resetSelectedFilters,
  onClearFilters
}) {
  const dispatch = useDispatch();
  const { newSegmentMode } = useSelector((state) => state.accountProfilesView);

  const handleAccountChange = (account) => {
    setSelectedAccount(account);
    setFiltersList(INITIAL_FILTERS_STATE.filters);
    setEventProp(INITIAL_FILTERS_STATE.eventProp);
    setListEvents(INITIAL_FILTERS_STATE.eventsList);
  };

  const toggleFilters = useCallback(() => {
    setFiltersExpanded((curr) => !curr);
    dispatch(setNewSegmentModeAction(false));
  }, [dispatch, setFiltersExpanded]);

  const handleCancel = useCallback(() => {
    toggleFilters();
    resetSelectedFilters();
  }, [resetSelectedFilters, toggleFilters]);

  if (filtersExpanded === false && newSegmentMode === false) {
    if (appliedFilters.filters.length + appliedFilters.eventsList.length > 0) {
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
            View{' '}
            {appliedFilters.filters.length + appliedFilters.eventsList.length}{' '}
            filter(s)
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
            Hide filters
          </Text>
          <SVG size={16} name='chevronDown' color='#8C8C8C' />
        </Button>
      </ControlledComponent>
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
        areFiltersDirty={areFiltersDirty}
        setEventProp={setEventProp}
        onCancel={handleCancel}
        onClearFilters={onClearFilters}
      />
    </div>
  );
}
export default PropertyFilter;
