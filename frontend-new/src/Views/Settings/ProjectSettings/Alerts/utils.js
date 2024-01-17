export const getErrorMsg = (errorItem, string) => {
    let index = errorItem?.failed_at?.indexOf(string);
    let msg = errorItem?.details[index] ? errorItem?.details[index] : null
    return msg
}

export const SLACK = "Slack";
export const WEBHOOK = "WH";
export const MS_TEAMS = "Teams";

export const getMsgPayloadMapping = (groupBy) => {
    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property){
        var obj = {}
        groupBy.map((item)=>{
             obj[item.property] = `{{$${item.property}}}`
        })
        return obj
    }
    else return null
}