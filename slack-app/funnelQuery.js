const generateFunnelHeaders = (json) => {
    res = [];
    res.push(json.rows[0][0]=="$no_group"?"Grouping":"Users")
    res.push("Total Conversion");
    res.push("Conversion Time");

    json.meta.query.ewp.forEach((element, index) => {
        res.push(element.na);
        if (index < json.meta.query.ewp.length - 1) {
            res.push(`In between time(${index} and ${index+1})`);
        }

    });
    //console.log(res);
    return res;
}

const generateFunnelData = (json) => {
    res1 = [];
    json.rows.forEach((element, index) => {
        var res=[];
        var i = 0;
        if(json.rows[0][0]!="$no_group")
            res.push("All");
        else{
            if(index==0)
                res.push("Overall");
            else{
                var a=`${element[i++]}`;
                while(i<(element.length -json.meta.query.ewp.length*2)){
                    a=a+`, ${element[i++]}`;
                }
                res.push(`${a}`)
            }
            
        }
        res.push(element[element.length - 1] + "%");
        if(json.rows[0][0]!="$no_group")
            res.push(getOverAllDuration(json.meta));
        else{
            const durationMetric = json.meta.metrics.find(
                (elem) => elem.title === "MetaStepTimeInfo"
              );
              const firstEventIdx = durationMetric.headers.findIndex(
                (elem) => elem === "step_0_1_time"
              );

              var count=0;
              var x=firstEventIdx;
              while(x<durationMetric.rows[index].length){
                  count=count+durationMetric.rows[index][x++];
              }
              res.push(formatDuration(count));
              
        }
        i=element.length -json.meta.query.ewp.length*2;
        var idx=0;
        if(json.rows[0][0]=="$no_group")idx=index;
        json.meta.query.ewp.forEach((element1, index) => {
            if (index == 0) res.push(element[i++] + " (100%)");
            else if (index == json.meta.query.ewp.length - 1 ) res.push(element[i++] + " (" + element[element.length - 1] + "%)");
            else res.push(element[i++] + " (" + element[i++] + "%)");
            if (index < json.meta.query.ewp.length - 1) {
                res.push(getStepDuration(json.meta,idx, index, index + 1));
            }
        });
        res1.push(res);
    });
    return res1;
}


const formatDuration = (seconds) => {
    seconds = Number(seconds);
    if (seconds < 60) {
        return Math.floor(seconds) + 's';
    }
    if (seconds < 3600) {
        const minutes = Math.floor(seconds / 60);
        const remains = Math.floor(seconds % 60);
        return `${minutes}m ${remains}s`;
    }
    if (seconds < 86400) {
        const hours = Math.floor(seconds / 3600);
        const remains = seconds % 3600;
        const minutes = Math.floor(remains / 60);
        return `${hours}h ${minutes}m`;
    }
    const days = Math.floor(seconds / 86400);
    const remains = seconds % 86400;
    const hours = Math.floor(remains / 3600);
    return `${days}d ${hours}h`;
};


const getOverAllDuration = (durationsObj) => {
    if (durationsObj && durationsObj.metrics) {
        const durationMetric = durationsObj.metrics.find(
            (d) => d.title === "MetaStepTimeInfo"
        );
        if (durationMetric && durationMetric.rows && durationMetric.rows.length) {
            try {
                let total = 0;
                
                durationMetric.rows[0].forEach((r) => {
                    total += Number(r);
                });
                return formatDuration(total);
            } catch (err) {
                return "NA";
            }
        }
    }
    return "NA";
};

const getStepDuration = (durationsObj,idx, index1, index2) => {
    let durationVal = "NA";
    if (durationsObj && durationsObj.metrics) {
        const durationMetric = durationsObj.metrics.find(
            (d) => d.title === "MetaStepTimeInfo"
        );
        if (
            durationMetric &&
            durationMetric.headers &&
            durationMetric.headers.length
        ) {
            try {
                const stepIndex = durationMetric.headers.findIndex(
                    (elem) => elem === `step_${index1}_${index2}_time`
                );
                if (stepIndex > -1) {
                    durationVal = formatDuration(durationMetric.rows[idx][stepIndex]);
                }
            } catch (err) {
                console.log(err);
            }
        }
    }
    return durationVal;
};

module.exports = {
    generateFunnelHeaders, generateFunnelData
}