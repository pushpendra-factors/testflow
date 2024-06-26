UPDATE plan_details
SET feature_list::demandbase = '{"expiry":9223372036854775807,"granularity":"","is_enabled_feature":true,"limit":0}'
WHERE id IN (2,3,4,5);