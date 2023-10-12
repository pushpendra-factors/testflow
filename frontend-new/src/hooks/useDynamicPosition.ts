import { PlacementType } from 'Components/GenericComponents/FaSelect/types';
import { throttle } from 'lodash';
import { RefObject, useEffect, useState } from 'react';

//Dynamically Positions only For Top and Bottom Based on Screen and Scroll Adjustments.

const checkForBottom = (
  relativeElement: RefObject<HTMLElement>,
  height: number
) => {
  const targetRect = relativeElement?.current?.getBoundingClientRect();
  if (targetRect) return targetRect.bottom + height < window.innerHeight;
  return false;
};
const checkForTop = (
  relativeElement: RefObject<HTMLElement>,
  height: number
) => {
  const targetRect = relativeElement?.current?.getBoundingClientRect();
  if (targetRect) return targetRect.top - 32 - height > 60;
  return false;
};

const calculatePosition = (
  relativeElement: RefObject<HTMLElement>,
  defaultPosition: PlacementType,
  height: number
) => {
  let position: PlacementType = 'Bottom';
  switch (defaultPosition) {
    case 'Bottom':
      if (checkForBottom(relativeElement, height)) {
        position = defaultPosition;
      } else if (checkForTop(relativeElement, height)) {
        position = 'Top';
      }
      break;
    case 'BottomLeft':
      if (checkForBottom(relativeElement, height)) {
        position = defaultPosition;
      } else if (checkForTop(relativeElement, height)) {
        position = 'TopLeft';
      }
      break;
    case 'BottomRight':
      if (checkForBottom(relativeElement, height)) {
        position = defaultPosition;
      } else if (checkForTop(relativeElement, height)) {
        position = 'TopRight';
      }
      break;
    case 'Top':
      if (checkForTop(relativeElement, height)) {
        position = defaultPosition;
      } else if (checkForBottom(relativeElement, height)) {
        position = 'Bottom';
      }
      break;
    case 'TopLeft':
      if (checkForTop(relativeElement, height)) {
        position = defaultPosition;
      } else if (checkForBottom(relativeElement, height)) {
        position = 'BottomLeft';
      }
      break;
    case 'TopRight':
      if (checkForTop(relativeElement, height)) {
        position = defaultPosition;
      } else if (checkForBottom(relativeElement, height)) {
        position = 'BottomRight';
      }
      break;
  }
  return position;
};

//Note:- Try to pass height prop as maxHeight,[Used To calculate Initial Height], Later Dropdown height is used.

const useDynamicPosition = (
  relativeElement: RefObject<HTMLElement> | null,
  targetElement: RefObject<HTMLElement>,
  defaultPosition: PlacementType = 'Bottom',
  height: number
) => {
  const [position, setPosition] = useState<PlacementType | null>();

  const handleEvent = throttle(function findPosition() {
    if (relativeElement) {
      const dropdownHeight =
        targetElement.current?.getBoundingClientRect()?.height || height;
      const calculatedPosition = calculatePosition(
        relativeElement,
        defaultPosition,
        dropdownHeight
      );
      setPosition(calculatedPosition);
    }
  }, 1000);

  useEffect(() => {
    if (relativeElement?.current && !targetElement.current) {
      const initilPos = calculatePosition(
        relativeElement,
        defaultPosition,
        height
      );
      setPosition(initilPos);
    }
    if (relativeElement?.current && targetElement?.current) {
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
  }, [targetElement?.current, relativeElement?.current]);
  return position;
};

export default useDynamicPosition;
