import React, {
  ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState
} from 'react';
import styles from './index.module.scss';
import { Text } from '../../factorsComponents';
import { OptionType, SingleSelectOptionClickCallbackType } from './types';
import { filterSearchFunction } from './utils';
import useKey from 'hooks/useKey';
import { HighlightSearchText } from 'Utils/dataFormatter';

interface SingleSelectProps {
  options: OptionType[];
  optionClickCallback?: SingleSelectOptionClickCallbackType;
  allowSearch: boolean;
  searchOption: OptionType | null;
  allowSearchTextSelection: boolean;
  searchTerm: string;
  extraClass?: string;
}
export default function SingleSelect({
  options,
  optionClickCallback,
  allowSearch,
  searchOption,
  allowSearchTextSelection,
  searchTerm,
  extraClass = ''
}: SingleSelectProps) {
  const [renderedOptionsIndex, setRenderOptionsIndex] = useState(12);
  const [hoveredOptionIndex, setHoveredOptionIndex] = useState(0);
  const [filteredOptions, setFilteredOptions] = useState<OptionType[]>([]);

  const handleOptionClick = (op: OptionType) => {
    if (optionClickCallback) optionClickCallback(op);
  };
  const dropdownRef = useRef(null);

  //Initialise Initial Values
  useEffect(() => {
    const filteredOpts = options.filter((op) =>
      filterSearchFunction(op, searchTerm)
    );
    removeHoverStateOfOption();
    setFilteredOptions(filteredOpts);
    setHoveredOptionIndex(0);
  }, [options, searchTerm, allowSearchTextSelection]);

  //KeyBoard Accessibilty
  const handleKeyArrowDown = useCallback(() => {
    const optionsLength =
      searchOption && allowSearchTextSelection
        ? filteredOptions.length + 1
        : filteredOptions.length;
    removeHoverStateOfOption();
    setHoveredOptionIndex((prevIndex) =>
      prevIndex < optionsLength - 1 ? prevIndex + 1 : 0
    );
  }, [hoveredOptionIndex, filteredOptions]);
  const handleKeyArrowUp = useCallback(() => {
    const optionsLength =
      searchOption && allowSearchTextSelection
        ? filteredOptions.length + 1
        : filteredOptions.length;
    if (hoveredOptionIndex > 0 || renderedOptionsIndex >= optionsLength - 1) {
      removeHoverStateOfOption();
      setHoveredOptionIndex((prevIndex) =>
        prevIndex > 0 ? prevIndex - 1 : optionsLength - 1
      );
    }
  }, [hoveredOptionIndex, filteredOptions]);
  const handleKeyEnter = useCallback(() => {
    if (searchOption && allowSearchTextSelection) {
      if (hoveredOptionIndex > 0)
        handleOptionClick(filteredOptions[hoveredOptionIndex - 1]);
      else {
        handleOptionClick(searchOption);
      }
    } else {
      handleOptionClick(filteredOptions[hoveredOptionIndex]);
    }
  }, [hoveredOptionIndex, filteredOptions]);
  const keyPressCallback = useCallback(
    (key: string) => {
      switch (key) {
        case 'Enter':
          handleKeyEnter();
          break;
        case 'ArrowUp':
          handleKeyArrowUp();
          break;
        case 'ArrowDown':
          handleKeyArrowDown();
          break;
      }
    },
    [hoveredOptionIndex, filteredOptions, allowSearchTextSelection]
  );
  useKey(['ArrowDown', 'ArrowUp', 'Enter'], keyPressCallback);

  //Scroll To Hovered Option Index
  const scrollToSelectedOption = () => {
    if (dropdownRef.current) {
      const selectedOptionElement =
        dropdownRef.current.children[hoveredOptionIndex];
      if (selectedOptionElement) {
        const { offsetTop } = selectedOptionElement;
        dropdownRef.current.scrollTop = offsetTop;
      }
    }
  };

  //Apply Hover Class to Option
  const addHoverStateOfOption = () => {
    if (dropdownRef.current) {
      const selectedOptionElement =
        dropdownRef.current.children[hoveredOptionIndex];
      if (selectedOptionElement) {
        (selectedOptionElement as Element).classList.add(styles.hoveredOption);
      }
    }
  };

  //Remove Hover Class to Option
  const removeHoverStateOfOption = () => {
    if (dropdownRef.current) {
      const selectedOptionElement =
        dropdownRef.current.children[hoveredOptionIndex];
      if (selectedOptionElement) {
        (selectedOptionElement as Element).classList.remove(
          styles.hoveredOption
        );
      }
    }
  };

  //Observability Intersection For Infinite Scrolling.
  useEffect(() => {
    //observer for tracking Last Element In Dropdown
    const observer = new IntersectionObserver(
      (entries) => {
        const lastOption = entries[0];

        if (!lastOption.isIntersecting) return;

        const keys = Object.keys(lastOption.target);
        const reactkey = keys.find((key) =>
          key.startsWith('__reactInternalInstance')
        );
        if (reactkey) {
          const optionKey: string = lastOption.target[reactkey].key;
          observer.unobserve(lastOption.target);
          if (optionKey) {
            //optionKey format is "op"+{number}
            if (loadNewOptions(Number(optionKey.substring(2)))) {
              const optionsElement = dropdownRef.current;
              if (optionsElement) {
                let lastItem = (optionsElement as Element).lastChild;
                observer.observe(lastItem as Element);
              }
            }
          }
        }
      },
      {
        root: dropdownRef.current,
        rootMargin: '50px'
      }
    );
    const optionsElement = dropdownRef.current;
    let lastElement;
    if (optionsElement) {
      lastElement = (optionsElement as Element).lastChild;
    }
    if (lastElement) {
      observer.observe(lastElement as Element);
    }
    return () => {
      const optionsElement = dropdownRef.current;
      let lastElement;
      if (optionsElement) {
        lastElement = (optionsElement as Element).lastChild;
      }
      if (lastElement) {
        observer.unobserve(lastElement as Element);
      }
    };
  }, [filteredOptions]);

  //Load New Options For Infinite Scrolling.
  const loadNewOptions = (lastIndex: number) => {
    const optionsElement = dropdownRef.current;
    const optionsLength = filteredOptions.length;
    //All options are renderred.
    if (lastIndex >= optionsLength - 1) {
      return false;
    }
    //Add more Options.
    if (optionsElement) {
      setRenderOptionsIndex((prevIndex) => prevIndex + 12);
      return true;
    }
    return false;
  };

  useEffect(() => {
    addHoverStateOfOption();
    scrollToSelectedOption();
  }, [filteredOptions, hoveredOptionIndex]);

  //Create renderOptions from filteredOptions.
  const renderFilteredOptions = useMemo(() => {
    let rendOpts: ReactNode[] = [];
    let index = 0;
    if (searchOption && allowSearchTextSelection) {
      // Adding Select Option Based On SearchTerm
      rendOpts.push(
        <div
          key={'op' + index}
          className={`${extraClass} ${
            allowSearch
              ? 'fa-select-group-select--options'
              : 'fa-select--options'
          }`}
          onClick={() => handleOptionClick(searchOption)}
        >
          <Text level={7} type={'title'} extraClass={'m-0'} weight={'thin'}>
            Select:
          </Text>
          <span className={`ml-1 ${styles.optText}`}>{searchOption.label}</span>
        </div>
      );
      index += 1;
    }
    filteredOptions.forEach((op) => {
      rendOpts.push(
        <div
          key={'op' + index}
          title={op.label}
          onClick={() => {
            handleOptionClick(op);
          }}
          className={`${extraClass} ${
            allowSearch
              ? 'fa-select-group-select--options'
              : 'fa-select--options'
          }`}
        >
          {op.labelNode ? (
            op.labelNode
          ) : searchTerm.length > 0 ? (
            <HighlightSearchText text={op.label} highlight={searchTerm} />
          ) : (
            <Text level={7} type={'title'} weight={'regular'}>
              {op.label}
            </Text>
          )}
        </div>
      );
      index += 1;
    });
    return rendOpts;
  }, [filteredOptions]);

  //renderOptions Till Index.
  const renderOptions = useMemo(() => {
    return renderFilteredOptions.slice(0, renderedOptionsIndex);
  }, [filteredOptions, renderedOptionsIndex]);

  return (
    <div
      className='flex flex-col'
      ref={dropdownRef}
      style={{ overflowY: 'auto' }}
    >
      {renderOptions}
    </div>
  );
}
