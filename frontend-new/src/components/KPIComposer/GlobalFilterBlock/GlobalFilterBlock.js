import React, { useState, useMemo, memo } from 'react';
import isEmpty from 'lodash/isEmpty';
import { QUERY_TYPE_EVENT, QUERY_TYPE_FUNNEL } from '../../../utils/constants';
import ComposerBlock from '../../QueryCommons/ComposerBlock';
import GlobalFilter from '../GlobalFilter';
import { getUserProperties } from '../../../reducers/coreQuery/middleware';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { areKpisInSameGroup } from '../../../utils/kpiQueryComposer.helpers';

const GlobalFilterBlock = ({
  queryType,
  queries,
  queryOptions,
  setGlobalFiltersOption,
  activeProject,
  selectedMainCategory,
  KPIConfigProps,
  setQueryOptions,
  DefaultQueryOptsVal,
  getUserProperties,
  propertyMaps
}) => {
  const [filterBlockOpen, setFilterBlockOpen] = useState(true);
  const isSameKPIGrp = useMemo(() => {
    return areKpisInSameGroup({ kpis: queries });
  }, [queries]);

  // useEffect(() => {
  //   if (
  //     ((!isSameKPIGrp && isArray(queries) && queries.length > 1) ||
  //       isEmpty(queries)) &&
  //     !isEqual(
  //       get(queryOptions, 'globalFilters', []),
  //       get(DefaultQueryOptsVal, 'globalFilters', [])
  //     )
  //   ) {
  //     // we will not show global filters when kpis selected are from different groups. hence we reset global filters
  //     setQueryOptions((currState) => {
  //       return {
  //         ...currState,
  //         globalFilters: DefaultQueryOptsVal.globalFilters
  //       };
  //     });
  //   }
  // }, []);

  if (isEmpty(queries)) {
    return null;
  }

  if (queryType === QUERY_TYPE_EVENT && queries.length < 1) {
    return null;
  }
  if (queryType === QUERY_TYPE_FUNNEL && queries.length < 2) {
    return null;
  }

  return (
    <ComposerBlock
      blockTitle={'FILTER BY'}
      isOpen={filterBlockOpen}
      showIcon={true}
      onClick={() => setFilterBlockOpen(!filterBlockOpen)}
      extraClass={'no-padding-l'}
    >
      <div key={0} className={'fa--query_block borderless no-padding '}>
        <GlobalFilter
          filters={queryOptions.globalFilters}
          setGlobalFilters={setGlobalFiltersOption}
          onFiltersLoad={[
            () => {
              getUserProperties(activeProject.id, queryType);
            }
          ]}
          selectedMainCategory={selectedMainCategory}
          KPIConfigProps={KPIConfigProps}
          propertyMaps={propertyMaps}
          isSameKPIGrp={isSameKPIGrp}
        />
      </div>
    </ComposerBlock>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  propertyMaps: state.kpi.kpi_property_mapping
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      getUserProperties
    },
    dispatch
  );

export default memo(
  connect(mapStateToProps, mapDispatchToProps)(GlobalFilterBlock)
);
