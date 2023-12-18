import anchorme from 'anchorme';

export const isValidURL = (str: string) => {
  return anchorme.validate.url(str);
};

function addProtocolIfMissing(url: string) {
  if (!/^https?:\/\//i.test(url)) {
    return 'http://' + url;
  }
  return url;
}
function urlHasBracketsOrBraces(inputString: string) {
  const regex = /[\[\](){}]/;
  return regex.test(inputString);
}

const truncateURL = (urlString: string) => { 
  let urlArray = urlString.split(',');
  urlArray = urlArray.map((urlText) => {
    const url = urlText.trim();

    //check URL is containing brakcets or braces. (not handled in anchorme npm package, hence added as spl conditional check)
    if(urlHasBracketsOrBraces(url)){
      return url
    }    
    //check URL is valid using anchorme npm package
    if (!isValidURL(url)) {
      return url;
    }
    const urlWithProtocol = addProtocolIfMissing(url);
    const urlObject = new URL(urlWithProtocol);
    const path = urlObject.pathname;

    // Check if there's a subdirectory
    const parts = path.split('/').filter(Boolean);

    if (parts.length > 0) {
      return `/${parts.slice(0).join('/')}`;
    } else {
      return url;
    }
  });
  return urlArray.join(', ');
};
export default truncateURL;