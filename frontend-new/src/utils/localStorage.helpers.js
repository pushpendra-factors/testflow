export const setItemToLocalStorage = (key, payload) => {
  localStorage.setItem(key, payload);
};

export const getItemFromLocalStorage = (key) => {
  return localStorage.getItem(key);
};

export const removeItemFromLocalStorage = (key) => {
  localStorage.removeItem(key);
};

export const clearLocalStorage = (key) => {
  localStorage.clear();
};
