import anchorme from 'anchorme';

function isValidURL(str: string) {
  return anchorme.validate.url(str);
}

function addProtocolIfMissing(url: string) {
  if (!/^https?:\/\//i.test(url)) {
    return 'http://' + url;
  }
  return url;
}

const truncateURL = (url: string) => {
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
};

export default truncateURL;
