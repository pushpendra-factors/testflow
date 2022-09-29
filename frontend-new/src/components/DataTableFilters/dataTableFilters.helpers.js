import { DEFAULT_CATEGORY_VALUE } from './dataTableFilters.constants';

export const getUpdatedFiltersOnCategoryChange = ({
  categoryIndex,
  selectedCategoryValue,
  currentFilters
}) => {
  const newCategory = {
    ...DEFAULT_CATEGORY_VALUE,
    field: selectedCategoryValue,
    key: new Date().getTime()
  };
  if (currentFilters.categories == null) {
    return {
      categoryCombinationOperator: 'AND',
      categories: [newCategory]
    };
  }
  if (categoryIndex == null) {
    return {
      ...currentFilters,
      categories: [...currentFilters.categories, newCategory]
    };
  }
  return {
    ...currentFilters,
    categories: currentFilters.categories.map((category, index) => {
      if (index !== categoryIndex) {
        return category;
      }
      return {
        ...category,
        field: selectedCategoryValue,
        values: []
      };
    })
  };
};

export const getUpdatedFiltersOnEqualityOperatorChange = ({
  categoryIndex,
  selectedOperator,
  currentFilters
}) => {
  return {
    ...currentFilters,
    categories: currentFilters.categories.map((category, index) => {
      if (index !== categoryIndex) {
        return category;
      }
      return {
        ...category,
        equalityOperator: selectedOperator
      };
    })
  };
};

export const getUpdatedFiltersOnCategoryDelete = ({
  categoryIndex,
  currentFilters
}) => {
  if (currentFilters.categories.length === 1) {
    return {};
  }
  return {
    ...currentFilters,
    categories: currentFilters.categories.filter((_, index) => {
      return categoryIndex !== index;
    })
  };
};

export const getUpdatedFiltersOnValueChange = ({
  categoryIndex,
  value,
  currentFilters,
  categoryFieldType = 'categorical'
}) => {
  if (categoryFieldType === 'numerical' || categoryFieldType === 'percentage') {
    return {
      ...currentFilters,
      categories: currentFilters.categories.map((category, index) => {
        if (index !== categoryIndex) {
          return category;
        }
        return {
          ...category,
          values:
            categoryFieldType === 'percentage' && Number(value) > 100
              ? 100
              : value
        };
      })
    };
  }
  return {
    ...currentFilters,
    categories: currentFilters.categories.map((category, index) => {
      if (index !== categoryIndex) {
        return category;
      }
      const isValuePresent = category.values.indexOf(value) > -1;
      return {
        ...category,
        values: isValuePresent
          ? category.values.filter((v) => v !== value)
          : [...category.values, value]
      };
    })
  };
};
