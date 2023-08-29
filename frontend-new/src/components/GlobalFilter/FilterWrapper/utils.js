import { PropTextFormat } from "Utils/dataFormatter"

export const getPropertyGroupLabel = (prpGrp)=>{
    const group = prpGrp === '$domains' ? 'all_account' : prpGrp;
    return `${PropTextFormat(group)} Properties`
}