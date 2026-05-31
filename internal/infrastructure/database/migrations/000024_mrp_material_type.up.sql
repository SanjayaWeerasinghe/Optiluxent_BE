ALTER TABLE mrp_material_master
    ADD COLUMN material_type VARCHAR(20) NOT NULL DEFAULT 'RAW_MATERIAL'
        CHECK (material_type IN ('RAW_MATERIAL','SEMI_FINISHED','SERVICE'));

CREATE INDEX IF NOT EXISTS idx_mrp_material_master_type ON mrp_material_master(tenant_id, material_type);
