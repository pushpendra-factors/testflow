import React, { useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import cx from 'classnames';
import { Button } from 'antd';
import { Text, SVG } from 'Components/factorsComponents';
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
  }, [dispatch, setFiltersExpanded]);

  const handleCancel = useCallback(() => {
    // toggleFilters();
    resetSelectedFilters();
  }, [resetSelectedFilters, toggleFilters]);

  if (filtersExpanded === false && newSegmentMode === false) {
    if (
      appliedFilters.filters.length +
        appliedFilters.eventsList.length +
        appliedFilters.secondaryFilters.length >
      0
    ) {
      return (
        <Button
          className={cx(
            'flex items-center justify-center gap-x-1',
            styles['collapse-button']
          )}
          type='text'
          onClick={toggleFilters}
        >
          <Text type='title' extraClass='mb-0' weight='medium' color='grey-6'>
            {`View ${
              appliedFilters.filters.length +
              appliedFilters.eventsList.length +
              appliedFilters.secondaryFilters.length
            } filter(s)`}
          </Text>
          <SVG size={16} name='chevronDown' color='#8C8C8C' />
        </Button>
      );
    }

    return (
      <Button
        className={cx(
          'flex items-center justify-center gap-x-1',
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

  if (!selectedFilters?.account?.length) return null;

  return (
    <div className='flex flex-col gap-y-4 w-full'>
      <ControlledComponent controller={newSegmentMode === false}>
        <Button
          className={cx(
            'flex items-center justify-center gap-x-1',
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
