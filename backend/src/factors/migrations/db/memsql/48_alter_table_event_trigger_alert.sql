alter table event_trigger_alerts add column teams_channel_associated_by text;

update event_trigger_alerts set teams_channel_associated_by = created_by;