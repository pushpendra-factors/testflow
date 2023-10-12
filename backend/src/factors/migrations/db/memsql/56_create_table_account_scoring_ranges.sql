CREATE TABLE IF NOT EXISTS account_scoring_ranges(
    project_id bigint NOT NULL,
    date text NOT NULL,
    bucket text ,
    created_at timestamp(6) NOT NULL,
    updated_at timestamp(6) NOT NULL,
    KEY (project_id, date) USING CLUSTERED COLUMNSTORE,
    PRIMARY KEY (project_id, date)
);
