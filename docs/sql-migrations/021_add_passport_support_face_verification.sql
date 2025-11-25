-- Migration 021: Add Passport Support for Face Verification
-- Date: November 21, 2025
-- Purpose: Allow foreign providers to use passport for face verification

-- Add document_type column to face_verifications table
ALTER TABLE face_verifications 
ADD COLUMN document_type VARCHAR(20) DEFAULT 'national_id' CHECK (document_type IN ('national_id', 'passport'));

-- Add document_id to reference the uploaded document
ALTER TABLE face_verifications 
ADD COLUMN document_id INTEGER;

-- Add foreign key to provider_documents table (if verification references existing document)
ALTER TABLE face_verifications 
ADD CONSTRAINT fk_document 
FOREIGN KEY (document_id) 
REFERENCES provider_documents(document_id) 
ON DELETE SET NULL;

-- Update existing records to have document_type = 'national_id' (for Thai providers)
UPDATE face_verifications 
SET document_type = 'national_id' 
WHERE document_type IS NULL;

-- Make document_type NOT NULL after setting defaults
ALTER TABLE face_verifications 
ALTER COLUMN document_type SET NOT NULL;

-- Add index for document_type queries
CREATE INDEX idx_face_verifications_document_type ON face_verifications(document_type);

-- Add comment for clarity
COMMENT ON COLUMN face_verifications.document_type IS 'Type of identification document: national_id (Thai ID card) or passport (Foreign passport)';
COMMENT ON COLUMN face_verifications.document_id IS 'References provider_documents.document_id for the ID card or passport document';
