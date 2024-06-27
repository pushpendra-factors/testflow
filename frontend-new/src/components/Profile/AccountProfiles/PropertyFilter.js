import React, { useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button } from 'antd';
import { SVG } from 'Components/factorsComponents';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';
import { setNewSegmentModeAction } from 'Reducers/accountProfilesView/actions';
import styles from './index.module.scss';
import FiltersBox from './FiltersBox';

function PropertyFilter({
  profileType = 'account',
  filtersExpanded,
  setFiltersExpanded,
  selectedFilters,
  setSelectedFilters,
  resetSelectedFilters,
  appliedFilters,
  applyFilters,
  areFiltersDirty,
  disableDiscardButton,
  isActiveSegment,
  setSaveSegmentModal,
  onClearFilters
}) {
  const dispatch = useDispatch();
  const { newSegmentMode: accountsNewSegmentMode } = useSelector(
    (state) => state.accountProfilesView
  );
  const { newSegmentMode: profilesNewSegmentMode } = useSelector(
    (state) => state.userProfilesView
  );

  const newSegmentMode =
    profileType === 'account' ? accountsNewSegmentMode : profilesNewSegmentMode;

  const toggleFilters = useCallback(() => {
    setFiltersExpanded((curr) => !curr);
    dispatch(setNewSegmentModeAction(false));
  }, []);

  const handleCancel = useCallback(() => {
    resetSelectedFilters();
  }, []);

  const renderShowFiltersButton = () => (
    <Button
      className={cx(
        'flex items-center justify-center px-2 button-shadow',
        styles['collapse-button']
      )}
      onClick={toggleFilters}
    >
      <SVG size={16} name='filterOutlined' color='#8C8C8C' />
      {`View ${
        appliedFilters.filters.length +
        appliedFilters.eventsList.length +
        appliedFilters.secondaryFilters.length
      } filter(s)`}
    </Button>
  );

  const renderFilterButton = () => (
    <Button
      className={cx(
        'flex items-center justify-center button-shadow',
        styles['filter-button']
      )}
      onClick={toggleFilters}
    >
      <SVG size={16} name='filterOutlined' color='#8C8C8C' />
      Filter
    </Button>
  );

  const shouldShowFilterButtons = () =>
    appliedFilters.filters.length +
      appliedFilters.eventsList.length +
      appliedFilters.secondaryFilters.length >
    0;

  if (filtersExpanded === false && newSegmentMode === false) {
    return shouldShowFilterButtons()
      ? renderShowFiltersButton()
      : renderFilterButton();
  }

  return (
    <div className='flex flex-col gap-y-4 w-full'>
      <ControlledComponent controller={!newSegmentMode}>
        <Button
          className={cx(
            'flex items-center justify-center button-shadow',
            styles['collapse-button']
          )}
          onClick={toggleFilters}
        >
          <SVG size={16} name='filterOutlined' color='#8C8C8C' />
          Hide Filters
        </Button>
      </ControlledComponent>
      <FiltersBox
        profileType={profileType}
        isActiveSegment={isActiveSegment}
        selectedFilters={selectedFilters}
        setSelectedFilters={setSelectedFilters}
        appliedFilters={appliedFilters}
        applyFilters={applyFilters}
        setSaveSegmentModal={setSaveSegmentModal}
        areFiltersDirty={areFiltersDirty}
        onCancel={handleCancel}
        onClearFilters={onClearFilters}
        disableDiscardButton={disableDiscardButton}
      />
    </div>
  );
}

export default PropertyFilter;
