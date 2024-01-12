import { isString } from 'lodash';

const isValidDomain = (urlText: string, domains: string[]) => {
  for (let i = 0; i < domains.length; i++) {
    if (urlText.startsWith(domains[i])) {
      const truncatedURL = urlText.slice(domains[i].length) || '';
      return truncatedURL.length > 1 ? truncatedURL : urlText;
    }
  }
  return false;
};

const truncateURL = (urlString: string = '', domainList: string[] = []) => {
  if (isString(urlString)) {
    let urlArray = urlString.split(',');
    urlArray = urlArray.map((url) => {
      const spaceTrimmedURL = url.trim();
      const truncatedURL = isValidDomain(spaceTrimmedURL, domainList);
      if (truncatedURL) return truncatedURL;
      return spaceTrimmedURL;
    });
    return urlArray.join(', ');
  }
  return urlString;
};
export default truncateURL;
