-- added column type for combined data(1) and page data(2), setting existing documents as combined type
--UP
ALTER TABLE google_organic_documents ADD COLUMN type INT;
UPDATE google_organic_documents SET type = 1;

--DOWN
-- ALTER TABLE google_organic_documents DROP COLUMN type;