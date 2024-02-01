-- add three optional columns used for koco: consolidate data warehouse access controller
ALTER TABLE task ADD COLUMN IF NOT EXISTS eks_name_space VARCHAR;
ALTER TABLE task ADD COLUMN IF NOT EXISTS eks_service_accounbt VARCHAR;
ALTER TABLE task ADD COLUMN IF NOT EXISTS s3_bucket_list VARCHAR;
