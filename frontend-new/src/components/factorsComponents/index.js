import lazyWithRetry from 'Utils/lazyWithRetry';
import Text from './Text';
import Spiner from './Spiner';
import Number from './Number';
import { FaErrorComp, FaErrorLog } from './FaErrorComp';

const SVG = lazyWithRetry(() => import('./SVG'));

export { Text, SVG, Spiner, Number, FaErrorComp, FaErrorLog };
