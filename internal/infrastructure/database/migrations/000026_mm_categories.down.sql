ALTER TABLE mm_materials DROP CONSTRAINT IF EXISTS mm_materials_category_id_fkey;
ALTER TABLE mm_materials
    ADD CONSTRAINT mm_materials_category_id_fkey
    FOREIGN KEY (category_id) REFERENCES product_categories(id);
DROP TABLE IF EXISTS mm_categories;
