export const filterURLValue = (term: string): string => {
  term = term.toLowerCase();
  //  Ignoring
  // Regex to detect https/http is there or not as a protocol
  let testURLRegex: RegExp =
    /^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$/;
  if (testURLRegex.test(term) === true) {
    term = term.split('://')[1];
  }
  // Below one is to remove last "slash(/)" from the pathname of the URL Value
  term = term.replace(/\/$/, '');
  return term;
};
