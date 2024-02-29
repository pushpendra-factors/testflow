from constants.constants import *
from collections import OrderedDict
import copy

class DataTransformation:

    # it take organization id for report rows and fetches org details like name location domain and append it to the member company report rows
    
    def enrich_company_details_to_company_insights(records, map_id_to_company_data):
        for data in records:
            id = data['pivotValues'][0].split(':')[3]
            data['id'] = id
            if id not in map_id_to_company_data:
                data['vanityName'] = '$none'
                data['localizedName'] = '$none'
                data['localizedWebsite'] = '$none'
                data['preferredCountry'] = '$none'
                data['companyHeadquarters'] = '$none'
        
            else:
                if 'vanityName' in map_id_to_company_data[id]:
                    data['vanityName'] = map_id_to_company_data[id]['vanityName']
                else:
                    data['vanityName'] = '$none'

                if 'localizedName' in map_id_to_company_data[id]:
                    data['localizedName'] = map_id_to_company_data[id]['localizedName']
                else:
                    data['localizedName'] = '$none'

                if 'localizedWebsite' in map_id_to_company_data[id]:
                    data['localizedWebsite'] = map_id_to_company_data[id]['localizedWebsite']
                else:
                    data['localizedWebsite'] = '$none'
                
                if 'name' in map_id_to_company_data[id] and (
                    'preferredLocale' in map_id_to_company_data[id]['name']) and (
                    'country' in map_id_to_company_data[id]['name']['preferredLocale']):
                    data['preferredCountry'] = map_id_to_company_data[id]['name']['preferredLocale']['country']
                else:
                    data['preferredCountry'] = '$none'
                
                if 'locations' in map_id_to_company_data[id]:
                    for location in map_id_to_company_data[id]['locations']:
                        if 'locationType' in location and (
                            location['locationType'] == 'HEADQUARTERS') and (
                            'address' in location) and (
                            'country' in location['address']):
                            data['companyHeadquarters'] = location['address']['country']
                            break
                        else:
                            data['companyHeadquarters'] = '$none'

        return records

    def enrich_dependencies_to_company_insights(company_insights, campaign_group_info_map, map_id_to_company_data):
        cg_enriched_company_insights = DataTransformation.enrich_campaign_group_info_to_company_insights(
                                                                company_insights, campaign_group_info_map)
        mc_enriched_company_insights = DataTransformation.enrich_company_details_to_company_insights(
                                                    cg_enriched_company_insights, map_id_to_company_data)
        return mc_enriched_company_insights
    
    def enrich_campaign_group_info_to_company_insights(records, campaign_group_info_map):
        updated_records = []
        for record in records:
            record.update(campaign_group_info_map[record['campaign_group_id']])
            updated_records.append(record)
        return updated_records
    

    def update_insights_with_metadata(response, doc_type, campaign_group_meta, campaign_meta, creative_meta):
        final_response = []
        for data in response:
            id = data['pivotValues'][0].split(':')[3]
            data.update({'id': id})
            if doc_type == CAMPAIGN_GROUP_INSIGHTS:
                if id in campaign_group_meta:
                    data.update(campaign_group_meta[id])
            elif doc_type == CAMPAIGN_INSIGHTS:
                if id in campaign_meta:
                    data.update(campaign_meta[id])
                    if campaign_meta[id][CAMPAIGN_GROUP_ID] in campaign_group_meta:
                        data.update(
                            campaign_group_meta[campaign_meta[id][CAMPAIGN_GROUP_ID]])
            elif doc_type == CREATIVE_INSIGHTS:
                if id in creative_meta:
                    data.update(creative_meta[id])
                    if creative_meta[id][CAMPAIGN_GROUP_ID] in campaign_group_meta:
                        data.update(
                            campaign_group_meta[creative_meta[id][CAMPAIGN_GROUP_ID]])
                    if creative_meta[id][CAMPAIGN_ID] in campaign_meta:
                        data.update(
                            campaign_meta[creative_meta[id][CAMPAIGN_ID]])
            final_response.append(data)
        
        return final_response

    
    def update_hierarchical_data(metadata, doc_type, campaign_group_meta, 
                                                campaign_meta, creative_meta):
        if doc_type == CAMPAIGN_GROUPS:
            for data in metadata:
                campaign_group_meta[str(data['id'])] = {
                                                CAMPAIGN_GROUP_ID: str(data['id']), 
                                                'campaign_group_name': data['name'],
                                                'campaign_group_status': data['status']}
            return campaign_group_meta
        
        if doc_type == CAMPAIGNS:
            for data in metadata:
                campaign_group_id = str(data['campaignGroup'].split(':')[3])
                campaign_meta[str(data['id'])] = {
                                                'campaign_group_id': campaign_group_id,
                                                'campaign_id': str(data['id']),
                                                'campaign_name': data['name'], 
                                                'campaign_status': data['status'], 
                                                'campaign_type': data['type']}
            return campaign_meta

        if doc_type == CREATIVES:
            for data in metadata:
                campaign_id = str(data['campaign'].split(':')[3])
                campaign_group_id = campaign_meta[campaign_id][CAMPAIGN_GROUP_ID]
                creative_meta[str(data['id'])] = {
                                                'campaign_group_id': campaign_group_id, 
                                                'campaign_id': campaign_id ,
                                                'creative_id': str(data['id']), 
                                                'creative_status': data['status'], 
                                                'creative_type': data['type']}
            return creative_meta
        return {}
    
    # split each row evenly in 7 or len_timerange parts and maps it to the timestamp
    # result : {timestamp1: [set of records], timestamp2: [set of records].....}
    def distribute_data_across_timerange(records, timerange):
        distributed_records_map_with_timestamp = OrderedDict()
        len_timerange = len(timerange)
        for record in records:
            impressions, clicks, costInUsd, costInLocalCurrency = int(record['impressions']), int(record['clicks']), float(record['costInUsd']), float(record['costInLocalCurrency'])
            impr_map_with_timestamp = DataTransformation.distribute_metric_across_given_timerange(impressions, timerange)
            clicks_map_with_timestamp = DataTransformation.distribute_metric_across_given_timerange(clicks, timerange)
            costInUsdDistributed, costInLocalCurrencyDistributed = costInUsd/len_timerange, costInLocalCurrency/len_timerange
            for timestamp in timerange:
                updated_record = copy.deepcopy(record)
                
                to_update = {
                            'impressions': impr_map_with_timestamp[timestamp], 
                            'clicks': clicks_map_with_timestamp[timestamp], 
                            'costInUsd': costInUsdDistributed, 
                            'costInLocalCurrency': costInLocalCurrencyDistributed
                            }
                updated_record.update(to_update)
                if timestamp not in distributed_records_map_with_timestamp:
                    distributed_records_map_with_timestamp[timestamp] = []
                distributed_records_map_with_timestamp[timestamp].append(updated_record)
        return distributed_records_map_with_timestamp
          
    def distribute_metric_across_given_timerange(metric, timerange):
        metric_map_with_timestamp = {}
        base_value_for_each_day = metric//len(timerange)
        remaining_value = metric%len(timerange)
        for timestamp in timerange:
            metric_map_with_timestamp[timestamp] = base_value_for_each_day
            if remaining_value > 0:
                metric_map_with_timestamp[timestamp] += 1
                remaining_value -= 1
        return metric_map_with_timestamp

    def transform_metadata_based_on_doc_type(metadata, doc_type, campaign_group_info, campaign_info, creative_info):
        updated_metadata = []
        if doc_type == CAMPAIGN_GROUPS:
            for data in metadata:

                dict_to_add = {CAMPAIGN_GROUP_ID: str(data['id']), 
                                'campaign_group_name': data['name'],
                                'campaign_group_status': data['status']}
                campaign_group_info[str(data['id'])] = dict_to_add
                data.update(dict_to_add)
                updated_metadata.append(data)
        
        if doc_type == CAMPAIGNS:
            for data in metadata:
                campaign_group_id = str(data['campaignGroup'].split(':')[3])
                dict_to_add = {
                                'campaign_group_id': campaign_group_id,
                                'campaign_id': str(data['id']),
                                'campaign_name': data['name'], 
                                'campaign_status': data['status'], 
                                'campaign_type': data['type']
                                }
                campaign_info[str(data['id'])] = dict_to_add
                data.update(dict_to_add)
                updated_metadata.append(data)
        
        if doc_type == CREATIVES:
            for data in metadata:
                campaign_id = str(data['campaign'].split(':')[3])
                campaign_group_id = campaign_info[campaign_id][CAMPAIGN_GROUP_ID]
                dict_to_add = {
                                'campaign_group_id': campaign_group_id, 
                                'campaign_id': campaign_id ,
                                'creative_id': str(data['id']), 
                                'creative_status': data['status'], 
                                'creative_type': data['type']
                            }
                creative_info[str(data['id'])] = dict_to_add
                data.update(dict_to_add)
                updated_metadata.append(data)
        return updated_metadata
        