import moment from 'moment';
import { runQuery, getFunnelData } from "../../reducers/coreQuery/services";

export const getDataFromServer = (query, unitId, dashboardId, durationObj, refresh, activeProjectId) => {
    if (query.query.query_group) {
        let queryGroup = query.query.query_group;
        if (durationObj.from && durationObj.to) {
            queryGroup = queryGroup.map(elem => {
                return {
                    ...elem,
                    fr: moment(durationObj.from).startOf('day').utc().unix(),
                    to: moment(durationObj.to).endOf('day').utc().unix(),
                    gbt: elem.gbt ? durationObj.frequency : ''
                };
            });
        } else {
            queryGroup = queryGroup.map(elem => {
                return {
                    ...elem,
                    fr: moment().startOf('week').utc().unix(),
                    to: moment().utc().unix(),
                    gbt: elem.gbt ? durationObj.frequency : ''
                };
            });
        }
        if (refresh) {
            return runQuery(activeProjectId, queryGroup);
        } else {
            return runQuery(activeProjectId, queryGroup, { refresh: false, unit_id: unitId, id: dashboardId });
        }
    } else {
        let funnelQuery = query.query;
        if (durationObj.from && durationObj.to) {
            funnelQuery = {
                ...funnelQuery,
                fr: moment(durationObj.from).startOf('day').utc().unix(),
                to: moment(durationObj.to).endOf('day').utc().unix()
            };
        } else {
            funnelQuery = {
                ...funnelQuery,
                fr: moment().startOf('week').utc().unix(),
                to: moment().utc().unix()
            };
        }
        if (refresh) {
            return getFunnelData(activeProjectId, funnelQuery);
        } else {
            return getFunnelData(activeProjectId, funnelQuery, { refresh: false, unit_id: unitId, id: dashboardId });
        }
    }
}