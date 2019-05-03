import moment from 'moment';

const COLORS = [
    "rgba(30, 139, 195, 1.0)",
    "rgba(39, 174, 96, 1.0)",
    "rgba(41, 128, 185, 1.0)",
    "rgba(142, 68, 173, 1.0)",
    "rgba(230, 126, 34, 1.0)",
    "rgba(231, 76, 60, 1.0)",
    "rgba(211, 84, 0, 1.0)",
    "rgba(243, 156, 18, 1.0)",
];

export function isStaging() {
    return ENV === "staging";
}

export function isProduction() {
    return ENV === "production"
}

export function getHostURL() {
    let host = BUILD_CONFIG.backend_host;
    return (host[host.length-1] === "/") ? host : host+"/";
}

export function deepEqual(x, y) {
    return JSON.stringify(x) === JSON.stringify(y);
}

export function trimQuotes(v) {
    if (v == null) return v;
    if (v[0] == '"') v = v.slice(1, v.length);
    if (v[v.length - 1] == '"') v = v.slice(0, v.length - 1);
    return v;
}

export function makeSelectOpts(values) {
    var opts = [];
    for(let i in values) {
        opts.push({label: values[i], value: values[i]});
    }
    return opts
}

export function makeSelectOpt(value, label) {
    if(!label) label = value;
    return { value: value, label: label }
}

export function removeElementByIndex(list=[], index) {
    let _list = [ ...list ] 
    _list.splice(index, 1);
    return _list; // new list.
}

// Create opts from src opts.
// opts src: { <value>: <label> }
export function createSelectOpts(opts) {
    let ropts = [];
    for(let k in opts) ropts.push(makeSelectOpt(k, opts[k]));
    return ropts;
}

// Selected opt from src opts.
// opts src: { <value>: <label> }
export function getSelectedOpt(opt, src) {
    if(!opt) return null;
    if(!src) return makeSelectOpt(opt);
    return { value: opt, label: src[opt] };
}

export function firstToUpperCase(str) {
    return str.charAt(0).toUpperCase() + str.slice(1);
}
    
export function getColor(index) {
    if (index == undefined || index == null) {
        // default color.
        return COLORS[0];
    }

    let ci = ((index + 1) % COLORS.length) - 1;
    return COLORS[ci];
}

export function isNumber(numString) {
    return numString.match(/^[+-]?\d+(\.\d+)?$/)
}

export function getChartScaleWithSpace(scale) {
    if (!scale || scale < 10) return 100;
    let multi10 = Math.pow(10, Math.floor(Math.log10(scale)))
    let buff = multi10 - (scale % multi10);
    buff = buff + multi10; 
    return scale + buff;
}

export function isSingleCountResult(result) {
    let rowKeys = Object.keys(result.rows);
    return rowKeys.length == 1 && result.rows[rowKeys[0]].length == 1;
}

export function slideUnixTimeWindowToCurrentTime(from, to) {
    let diff = to - from;
    let resultTo =  moment(new Date()).unix();
    let resultFrom = resultTo - diff;
    return { from: resultFrom, to: resultTo };
}