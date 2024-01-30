import anchorme from 'anchorme';

export const isValidURL = (str: string = '') => {
  return anchorme.validate.url(str);
};
