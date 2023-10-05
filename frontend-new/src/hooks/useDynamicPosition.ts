import { PlacementType } from 'Components/GenericComponents/FaSelect/types';
import { throttle } from 'lodash';
import { RefObject, useEffect, useRef, useState } from 'react';

//Dynamically Positions only For Top and Bottom Based on Screen and Scroll Adjustments.
const useDynamicPosition = (
  targetElement: RefObject<HTMLElement>,
  defaultPosition: PlacementType = 'Bottom'
) => {
  const [position, setPosition] = useState(defaultPosition);
  const positionRef = useRef(defaultPosition);
  useEffect(() => {
    positionRef.current = position;
  }, [position]);
  const checkForBottom = () => {
    const currentPosition = positionRef.current;
    const dropdownRect = targetElement?.current?.getBoundingClientRect();
    if (dropdownRect) {
      if (
        currentPosition === 'Top' ||
        currentPosition === 'TopLeft' ||
        currentPosition === 'TopRight'
      ) {
        return (
          dropdownRect.bottom + 32 + dropdownRect.height < window.innerHeight
        );
      }
      return dropdownRect.top + dropdownRect.height < window.innerHeight;
    }
    return true;
  };
  const checkForTop = () => {
    const currentPosition = positionRef.current;
    const dropdownRect = targetElement?.current?.getBoundingClientRect();
    const pageTopOffset = 60;
    if (dropdownRect) {
      if (
        currentPosition === 'Top' ||
        currentPosition === 'TopLeft' ||
        currentPosition === 'TopRight'
      ) {
        return dropdownRect.top > pageTopOffset;
      }
      return dropdownRect.top - 32 - dropdownRect.height > pageTopOffset;
    }
    return true;
  };
  const handleEvent = throttle(function findPosition() {
    const dropdownRect = targetElement?.current?.getBoundingClientRect();
    if (dropdownRect) {
      switch (defaultPosition) {
        case 'Bottom':
          if (checkForBottom()) {
            setPosition(defaultPosition);
          } else if (checkForTop()) {
            setPosition('Top');
          } else {
            setPosition('Bottom');
          }
          break;
        case 'BottomLeft':
          if (checkForBottom()) {
            setPosition(defaultPosition);
          } else if (checkForTop()) {
            setPosition('TopLeft');
          } else {
            setPosition('Bottom');
          }
          break;
        case 'BottomRight':
          if (checkForBottom()) {
            setPosition(defaultPosition);
          } else if (checkForTop()) {
            setPosition('TopRight');
          } else {
            setPosition('Bottom');
          }
          break;
        case 'Top':
          if (checkForTop()) {
            setPosition(defaultPosition);
          } else if (checkForBottom()) {
            setPosition('Bottom');
          } else {
            setPosition('Bottom');
          }
          break;
        case 'TopLeft':
          if (checkForTop()) {
            setPosition(defaultPosition);
          } else if (checkForBottom()) {
            setPosition('BottomLeft');
          } else {
            setPosition('Bottom');
          }
          break;
        case 'TopRight':
          if (checkForTop()) {
            setPosition(defaultPosition);
          } else if (checkForBottom()) {
            setPosition('BottomRight');
          } else {
            setPosition('Bottom');
          }
          break;
      }
    }
  }, 1000);
  useEffect(() => {
    if (targetElement) {
      window.addEventListener('resize', handleEvent);
      //Only For Query Composer.
      const myElement = document.querySelector(
        '.fa-modal--full-width .ant-modal-content'
      );
      if (myElement) {
        myElement?.addEventListener('scroll', handleEvent);
      } else {
        window.addEventListener('scroll', handleEvent);
      }
      setTimeout(() => handleEvent(), 500);
    }

    return () => {
      window.removeEventListener('resize', handleEvent);
      const myElement = document.querySelector(
        '.fa-modal--full-width .ant-modal-content'
      );
      if (myElement) {
        myElement?.removeEventListener('scroll', handleEvent);
      } else {
        window.removeEventListener('scroll', handleEvent);
      }
    };
  }, [targetElement]);
  return position;
};

export default useDynamicPosition;
