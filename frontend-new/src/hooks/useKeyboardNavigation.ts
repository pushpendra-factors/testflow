import React from 'react';

function useKeyboardNavigation(
  ref: React.RefObject<HTMLObjectElement>,
  event: any
) {
  // Conditional PreventDefault()
  // This prevents page from scrolling if shifted focus using up/down
  // And Makes it possible to enter into search box
  if (event.keyCode === 40 || event.keyCode === 38) event.preventDefault();
  // Define focusable selectors
  if (!ref.current) return;

  const focusableSelectors = '[tabindex]:not([tabindex="-1"])';
  const focusableElements: any = Array.from(
    ref.current.querySelectorAll(focusableSelectors)
  );
  const activeElement = document.activeElement;
  const currentIndex = focusableElements.indexOf(activeElement);

  if (currentIndex === -1) return; // Active element is not in the focusable list

  let newIndex;

  if (event.keyCode === 40) {
    newIndex = (currentIndex + 1) % focusableElements.length;
  } else if (event.keyCode === 38) {
    newIndex =
      (currentIndex - 1 + focusableElements.length) % focusableElements.length;
  } else {
    return; // If the key is not left or right arrow, do nothing
  }
  focusableElements[newIndex]?.focus();
}
export default useKeyboardNavigation;
