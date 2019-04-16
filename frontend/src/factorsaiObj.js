import factorsai from 'factorsai';
import { isProduction } from './util'

if (isProduction()) {
    factorsai.init(BUILD_CONFIG.factors_sdk_token); 
} else {
    // host changed to support other environments.
    factorsai.init(BUILD_CONFIG.factors_sdk_token, {host: BUILD_CONFIG.backend_host}); 
}

export default factorsai;