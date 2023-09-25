from constants import *
from collections import OrderedDict
import copy

class DataTransformation:

    # it take organization id for report rows and fetches org details like name location domain and append it to the member company report rows
    
    def enrich_company_details(records, map_id_to_company_data):
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

    
    def enrich_campaign_group_info(records, campaign_group):
        updated_records = []
        campaign_group_dict_to_add = {
            'campaign_group_id': campaign_group['id'],
            'campaign_group_name': campaign_group['value']['name'],
            'campaign_group_status': campaign_group['value']['status']
        }
        for record in records:
            record.update(campaign_group_dict_to_add)
            updated_records.append(record)
        return updated_records

    def update_result_with_metadata(response, doc_type, campaign_group_meta, 
                                                campaign_meta, creative_meta):
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
    def distribute_data_across_timerange(records, timerange):
        distributed_records_map_with_timestamp = OrderedDict()
        len_timerange = len(timerange)
        for record in records:
            impressions, clicks, costInUsd, costInLocalCurrency = int(record['impressions']), int(record['clicks']), float(record['costInUsd']), float(record['costInLocalCurrency'])
            impr_quotient, clicks_quotient = impressions//len_timerange, clicks//len_timerange
            impr_rem, clicks_rem = impressions%len_timerange, clicks%len_timerange
            costInUsdDistributed, costInLocalCurrencyDistributed = costInUsd/len_timerange, costInLocalCurrency/len_timerange
            for timestamp in timerange:
                updated_record = copy.deepcopy(record)
                impr_to_update = impr_quotient
                if impr_rem >0:
                    impr_to_update += 1
                    impr_rem -=1
                clicks_to_update = clicks_quotient
                if clicks_rem >0:
                    clicks_to_update += 1
                    clicks_rem -=1
                
                to_update = {'impressions': impr_to_update, 'impr_check': impressions, 'clicks': clicks_to_update, 
                                       'costInUsd': costInUsdDistributed, 'costInLocalCurrency': costInLocalCurrencyDistributed}
                updated_record.update(to_update)
                if timestamp not in distributed_records_map_with_timestamp:
                    distributed_records_map_with_timestamp[timestamp] = []
                distributed_records_map_with_timestamp[timestamp].append(updated_record)
        return distributed_records_map_with_timestamp
    
    def enrich_campaign_company_fields_for_member_company_data(access_token, 
                            map_of_id_to_company_data, records, campaign_group):
        
        updated_records = DataTransformation.enrich_campaign_group_info(records, 
                                                                    campaign_group)

        final_records = DataTransformation.enrich_company_details(updated_records, 
                                                                map_of_id_to_company_data)
        return final_records

        
        

        
    